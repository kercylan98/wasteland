package wasteland

type Process interface {
	// GetID 返回这个进程的唯一标识
	GetID() ResourceLocator
}

type ProcessLifecycle interface {
	Process

	// Initialize 初始化进程
	Initialize()

	// Terminate 终止进程，当进程被终止时调用，参数是发起终止的进程 ID，否则使用自身 ID
	Terminate(operator ResourceLocator)

	// Terminated 检查进程是否已经终止
	Terminated() bool
}

type ProcessHandler interface {
	Process

	HandleMessage(sender ResourceLocator, priority MessagePriority, message Message)
}
