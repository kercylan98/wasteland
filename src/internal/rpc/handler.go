package rpc

type Handler interface {
	Handle(stream Stream, message []byte)
}

type HandlerFn func(stream Stream, message []byte)

func (f HandlerFn) Handle(stream Stream, message []byte) {
	f(stream, message)
}
