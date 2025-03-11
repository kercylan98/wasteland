package wasteland

import "github.com/kercylan98/wasteland/src/internal/rpc"

func init() {
	rpc.RegisterName("w:metaImpl", &metaImpl{})
	rpc.RegisterName("w:processIdImpl", &processIdImpl{})
	rpc.RegisterName("w:rpcMessage", &rpcMessage{})
}

type rpcMessage struct {
	Sender   ProcessId
	Target   ProcessId
	Agent    ProcessId
	Priority MessagePriority
	Message  Message
}
