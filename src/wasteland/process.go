package wasteland

type ProcessMeta interface {
	// GetID 返回这个进程的唯一标识
	GetID() ProcessId
}

type Process interface {
	ProcessMeta
}

type ProcessLifecycle interface {
	Process

	// Initialize 初始化进程
	Initialize()

	// Terminate 终止进程，当进程被终止时调用，参数是发起终止的进程 ID，否则使用自身 ID
	Terminate(operator ProcessId)
}

type ProcessHandler interface {
	Process

	HandleMessage(sender ProcessId, priority MessagePriority, message Message)
}

type ProcessMessageAgent interface {
	Process

	HandleAgentMessage(sender, target ProcessId, priority MessagePriority, message Message)
}
