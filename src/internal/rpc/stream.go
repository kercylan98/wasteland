package rpc

import (
	"github.com/kercylan98/wasteland/src/internal/protobuf/protobuf"
)

var _ Stream = (*streamImpl)(nil)

type StreamHandler interface {
	Send(message *protobuf.Message) error
	Recv() (*protobuf.Message, error)
	CloseSend() error
}

type Stream interface {
	StreamHandler

	Initialize(addr string)

	GetAddr() string

	Close(rpc Serve)
}

func newStream(stream StreamHandler) Stream {
	return &streamImpl{
		StreamHandler: stream,
	}
}

type streamImpl struct {
	StreamHandler
	addr string
}

func (s *streamImpl) GetAddr() string {
	return s.addr
}

func (s *streamImpl) Initialize(addr string) {
	s.addr = addr
}

func (s *streamImpl) Close(rpc Serve) {
	_ = s.StreamHandler.CloseSend()
	if s.addr != "" {
		rpc.Unbind(s)
	}
}
