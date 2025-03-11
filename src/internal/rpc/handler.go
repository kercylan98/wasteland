package rpc

type Handler interface {
	Handle(message []byte)
}

type HandlerFn func(message []byte)

func (f HandlerFn) Handle(message []byte) {
	f(message)
}
