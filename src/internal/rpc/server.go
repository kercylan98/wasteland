package rpc

import (
	"github.com/kercylan98/wasteland/src/internal/protobuf/protobuf"
)

var _ protobuf.RPCServiceServer = (*server)(nil)

type server struct {
	*protobuf.UnimplementedRPCServiceServer
	rpc Serve
}

func (s *server) OpenStream(oss protobuf.RPCService_OpenStreamServer) error {
	stream := newStream(oss.(*protobuf.RPCStream), nil)
	if handshake, err := s.rpc.WaitHandshake(stream); err != nil {
		return err
	} else {
		stream.Initialize(handshake.Address)
		s.rpc.Bind(stream)
		s.rpc.ListenMessage(stream)
		return nil
	}
}
