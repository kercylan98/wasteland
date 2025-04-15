package protobuf

type RPCStream struct {
	RPCService_OpenStreamServer
}

func (*RPCStream) CloseSend() error {
	return nil
}
