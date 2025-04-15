package rpc

import "github.com/kercylan98/wasteland/src/internal/protobuf/protobuf"

type Handler interface {
	Handle(stream Stream, message *protobuf.Message_BatchEntry)
}

type HandlerFn func(stream Stream, message *protobuf.Message_BatchEntry)

func (f HandlerFn) Handle(stream Stream, message *protobuf.Message_BatchEntry) {
	f(stream, message)
}
