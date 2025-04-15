package wasteland

import "sync/atomic"

func NewResourceLocator(address Address, path Path) ResourceLocator {
	return &resourceLocator{
		A: address,
		P: path,
	}
}

// ResourceLocator 是一个资源定位符接口，它提供了获取资源地址和路径的方法
type ResourceLocator interface {
	// Address 返回资源的地址，例如：127.0.0.1:8080
	Address() Address

	// Path 返回资源的路径，例如：/api/v1/resource
	Path() Path
}

// ResourceCache 是资源缓存的接口，它允许对资源进行存储和加载
type ResourceCache interface {
	ResourceLocator

	Load() Process

	Store(process Process)
}

type resourceLocator struct {
	A     Address
	P     Path
	cache atomic.Pointer[Process] // 进程缓存，该字段并非序列化的一部分
}

func (g *resourceLocator) Address() Address {
	return g.A
}

func (g *resourceLocator) Path() Path {
	return g.P
}

func (g *resourceLocator) Load() Process {
	v := g.cache.Load()
	if v == nil {
		return nil
	}
	return *v
}

func (g *resourceLocator) Store(process Process) {
	g.cache.Store(&process)
}
