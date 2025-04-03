package rpc

import (
	"github.com/kercylan98/wasteland/src/internal/protobuf/protobuf"
	"google.golang.org/grpc"
)

var _ Stream = (*streamImpl)(nil)

type StreamHandler interface {
	Send(message *protobuf.Message) error
	Recv() (*protobuf.Message, error)
	CloseSend() error
}

type Stream interface {
	StreamHandler

	Initialize(addr string, provider CodecProvider)

	GetAddr() string

	Encode(m any) (bytes []byte, err error)

	Decode(m any, data []byte) (err error)

	Close(rpc Serve)
}

func newStream(stream StreamHandler, cc *grpc.ClientConn) Stream {
	return &streamImpl{
		StreamHandler: stream,
		cc:            cc,
	}
}

type streamImpl struct {
	StreamHandler
	codec Codec
	cc    *grpc.ClientConn
	addr  string
}

func (s *streamImpl) Encode(m any) (bytes []byte, err error) {
	return s.codec.Encode(m)
}

func (s *streamImpl) Decode(m any, data []byte) (err error) {
	m, err = s.codec.Decode(m, data)
	return
}

func (s *streamImpl) GetAddr() string {
	return s.addr
}

func (s *streamImpl) Initialize(addr string, provider CodecProvider) {
	s.addr = addr
	s.codec = provider.Provide()
}

func (s *streamImpl) Close(rpc Serve) {
	_ = s.StreamHandler.CloseSend()
	if s.addr != "" {
		rpc.Unbind(s)
	}
	if s.cc != nil {
		_ = s.cc.Close()
	}
}
