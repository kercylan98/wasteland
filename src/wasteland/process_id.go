package wasteland

import "sync/atomic"

var (
	_ ProcessIdCache = (*processIdImpl)(nil)
)

type ProcessId interface {
	Meta

	// Path 返回这个进程的资源路径
	Path() Path
}

type ProcessIdCache interface {
	ProcessId

	Load() Process

	Store(process Process)
}

func NewProcessId(meta Meta, path Path) ProcessId {
	return newProcessId(meta, path)
}

func newProcessId(meta Meta, path Path) ProcessId {
	return &processIdImpl{
		Meta:   meta,
		IdPath: path,
	}
}

type processIdImpl struct {
	Meta
	IdPath Path
	cache  atomic.Pointer[Process] // 进程缓存，该字段并非序列化的一部分
}

func (p *processIdImpl) Load() Process {
	v := p.cache.Load()
	if v == nil {
		return nil
	}
	return *v
}

func (p *processIdImpl) Store(process Process) {
	p.cache.Store(&process)
}

func (p *processIdImpl) Path() Path {
	return p.IdPath
}
