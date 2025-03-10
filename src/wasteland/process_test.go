package wasteland_test

import "github.com/kercylan98/wasteland/src/wasteland"

var (
	_ wasteland.Process             = (*TestFnProcess)(nil)
	_ wasteland.ProcessLifecycle    = (*TestFnProcess)(nil)
	_ wasteland.ProcessHandler      = (*TestFnProcess)(nil)
	_ wasteland.ProcessMessageAgent = (*TestFnProcess)(nil)
)

type TestFnProcess struct {
	ID                   wasteland.ProcessId
	OnInitialize         func()
	OnTerminate          func(operator wasteland.ProcessId)
	OnHandleMessage      func(sender wasteland.ProcessId, priority wasteland.MessagePriority, message wasteland.Message)
	OnHandleAgentMessage func(sender, target wasteland.ProcessId, priority wasteland.MessagePriority, message wasteland.Message)
}

func (t *TestFnProcess) GetID() wasteland.ProcessId {
	return t.ID
}

func (t *TestFnProcess) Initialize() {
	if t.OnInitialize != nil {
		t.OnInitialize()
	}
}

func (t *TestFnProcess) Terminate(operator wasteland.ProcessId) {
	if t.OnTerminate != nil {
		t.OnTerminate(operator)
	}
}

func (t *TestFnProcess) HandleMessage(sender wasteland.ProcessId, priority wasteland.MessagePriority, message wasteland.Message) {
	if t.OnHandleMessage != nil {
		t.OnHandleMessage(sender, priority, message)
	}
}

func (t *TestFnProcess) HandleAgentMessage(sender, target wasteland.ProcessId, priority wasteland.MessagePriority, message wasteland.Message) {
	if t.OnHandleAgentMessage != nil {
		t.OnHandleAgentMessage(sender, target, priority, message)
	}
}
