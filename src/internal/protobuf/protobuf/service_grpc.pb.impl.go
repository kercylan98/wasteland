package protobuf

type RPCStream = rPCServiceOpenStreamServer

func (*rPCServiceOpenStreamServer) CloseSend() error {
	return nil
}
