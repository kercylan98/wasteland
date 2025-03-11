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

	Encode(m any) (bytes []byte, err error)

	Decode(m any, data []byte) (err error)

	Close(rpc Serve)
}

func newStream(stream StreamHandler) Stream {
	return &streamImpl{
		StreamHandler: stream,
	}
}

type streamImpl struct {
	StreamHandler
	addr  string
	codec *codec
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

func (s *streamImpl) Initialize(addr string) {
	s.addr = addr
	s.codec = newCodec()
}

func (s *streamImpl) Close(rpc Serve) {
	_ = s.StreamHandler.CloseSend()
	if s.addr != "" {
		rpc.Unbind(s)
	}
}
