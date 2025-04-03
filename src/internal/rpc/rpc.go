package rpc

import (
	"context"
	"errors"
	"github.com/kercylan98/go-log/log"
	"github.com/kercylan98/wasteland/src/internal/protobuf/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"net"
	"sync"
	"time"
)

type RPC interface {
	Run() (err error)

	Get(addr string) (stream Stream, err error)

	Stop()

	CloseStream(stream Stream)

	GracefulStop()
}

type Serve interface {
	RPC

	Bind(stream Stream)

	Unbind(stream Stream)

	WaitHandshake(stream Stream) (*protobuf.Message_Handshake, error)

	ListenMessage(stream Stream)
}

type Config struct {
	Handler       Handler
	Listener      net.Listener
	Logger        log.Provider
	CodecProvider CodecProvider
}

func New(config Config) RPC {
	if config.CodecProvider == nil {
		config.CodecProvider = CodecProviderFN(func() Codec {
			return newCodec()
		})
	}
	return &rpcImpl{
		config:  config,
		streams: make(map[string][]Stream),
	}
}

type rpcImpl struct {
	config  Config
	addr    net.Addr
	streams map[string][]Stream
	rw      sync.RWMutex
	grpc    *grpc.Server
}

func (r *rpcImpl) CloseStream(stream Stream) {
	stream.Close(r)
}

func (r *rpcImpl) Run() (err error) {
	r.grpc = grpc.NewServer()
	r.grpc.RegisterService(&protobuf.RPCService_ServiceDesc, &server{rpc: r, codecProvider: r.config.CodecProvider})

	r.addr = r.config.Listener.Addr()
	return r.grpc.Serve(r.config.Listener)
}

func (r *rpcImpl) Stop() {
	r.grpc.Stop()
}

func (r *rpcImpl) GracefulStop() {
	r.grpc.GracefulStop()
}

// Bind 绑定远程流
func (r *rpcImpl) Bind(stream Stream) {
	addr := stream.GetAddr()

	r.rw.Lock()
	defer r.rw.Unlock()

	r.streams[addr] = append(r.streams[addr], stream)
}

// Unbind 解绑远程流
func (r *rpcImpl) Unbind(stream Stream) {
	addr := stream.GetAddr()

	r.rw.Lock()
	defer r.rw.Unlock()

	if streams, ok := r.streams[addr]; ok {
		for i, s := range streams {
			if s == stream {
				r.streams[addr] = append(streams[:i], streams[i+1:]...)
				return
			}
		}
	}
}

func (r *rpcImpl) Get(addr string) (stream Stream, err error) {
	// 采用读锁获取目标，避免每次获取都加写锁，需要使用双重校验来确保不会重复创建
	r.rw.RLock()
	if streams, ok := r.streams[addr]; ok {
		// 随机选择一个远程流
		var index int64
		if len(r.streams) > 0 {
			index = time.Now().UnixMilli() % int64(len(r.streams))
			if index < int64(len(streams)) {
				r.rw.RUnlock()
				return streams[index], nil
			}
		}
	}
	r.rw.RUnlock()
	if stream, err = r.createRemoteStream(addr); err != nil {
		r.config.Logger.Provide().Warn("remote", log.String("event", "dial"), log.String("addr", addr), log.Any("info", "retrying on a continuous basis"), log.Any("err", err))
		return nil, err
	} else {
		return stream, nil
	}
}

// createRemoteStream 创建一个远程流
func (r *rpcImpl) createRemoteStream(addr string) (Stream, error) {
	r.rw.Lock()
	defer r.rw.Unlock()

	// 双重校验，防止重复创建
	streams, exist := r.streams[addr]
	if exist && len(streams) == 5 {
		return streams[time.Now().UnixMilli()%5], nil
	}

	// 创建客户端并打开连接
	stream, err := r.openRemoteStream(addr)
	if err != nil {
		return nil, err
	}

	// 发起握手
	if err = r.execClientHandshake(stream); err != nil {
		stream.Close(r)
		return nil, err
	}

	// 激活
	if handshake, err := r.WaitHandshake(stream); err != nil {
		stream.Close(r)
		return nil, err
	} else {
		stream.Initialize(handshake.Address, r.config.CodecProvider)
		r.Bind(stream)
	}

	// 记录
	streams = append(streams, stream)

	go r.ListenMessage(stream)
	return stream, nil
}

// execClientHandshake 由客户端发送握手消息
func (r *rpcImpl) execClientHandshake(stream Stream) error {
	msg := &protobuf.Message{
		MessageType: &protobuf.Message_Handshake_{Handshake: &protobuf.Message_Handshake{Address: r.addr.String()}},
	}
	if err := stream.Send(msg); err != nil {
		return err
	}
	return nil
}

// WaitHandshake 等待握手消息的到来并回复
//   - 第一条收到的消息必须是来自客户端发起的握手消息，并且回复握手消息
func (r *rpcImpl) WaitHandshake(stream Stream) (*protobuf.Message_Handshake, error) {
	message, err := stream.Recv()
	if err != nil {
		return nil, err
	}

	handshake := message.GetHandshake()
	if handshake == nil {
		return nil, errors.New("waitHandshake message is expected")
	}

	addr := r.addr.String()
	if handshake.Address == addr {
		return nil, errors.New("loop-back connection is not allowed")
	}

	return handshake, stream.Send(&protobuf.Message{
		MessageType: &protobuf.Message_Handshake_{Handshake: &protobuf.Message_Handshake{Address: addr}},
	})
}

// ListenMessage 开始监听远程流消息
func (r *rpcImpl) ListenMessage(stream Stream) {
	defer func() {
		stream.Close(r)
	}()

	var message *protobuf.Message
	var err error
	for {
		message, err = stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}
			return
		}

		r.handleMessage(stream, message)
	}
}

// openRemoteStream 打开 GRPC 远程流
func (r *rpcImpl) openRemoteStream(addr string) (Stream, error) {
	cc, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := protobuf.NewRPCServiceClient(cc)
	server, err := client.OpenStream(context.Background())
	if err != nil {
		return nil, err
	}

	return newStream(server, cc), nil
}

// handleMessage 处理远程流消息
func (r *rpcImpl) handleMessage(stream Stream, message *protobuf.Message) {
	switch m := message.GetMessageType().(type) {
	case *protobuf.Message_Batch_:
		r.onStreamBatchMessage(stream, m.Batch)
	case *protobuf.Message_Farewell_:
		r.onStreamFarewellMessage(stream, m.Farewell)
	}
}

func (r *rpcImpl) onStreamBatchMessage(stream Stream, batch *protobuf.Message_Batch) {
	for _, messageBytes := range batch.Messages {
		r.config.Handler.Handle(stream, messageBytes)
	}
}

func (r *rpcImpl) onStreamFarewellMessage(stream Stream, farewell *protobuf.Message_Farewell) {

}
