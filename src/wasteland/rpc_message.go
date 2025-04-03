package wasteland

import (
	"encoding/gob"
)

func init() {
	gob.RegisterName("w:metaImpl", &metaImpl{})
	gob.RegisterName("w:processIdImpl", &processIdImpl{})
	gob.RegisterName("w:rpcMessage", &rpcMessage{})
}

type rpcMessage struct {
	Sender   ProcessId
	Target   ProcessId
	Agent    ProcessId
	Priority MessagePriority
	Message  Message
}
