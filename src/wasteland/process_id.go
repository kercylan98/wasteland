package wasteland

type ProcessId interface {
	Meta

	// Path 返回这个进程的资源路径
	Path() Path
}

func newProcessId(meta Meta, path Path) ProcessId {
	return &processIdImpl{
		Meta: meta,
		path: path,
	}
}

type processIdImpl struct {
	Meta
	path Path
}

func (p *processIdImpl) Path() Path {
	return p.path
}
