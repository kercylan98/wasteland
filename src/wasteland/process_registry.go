package wasteland

import (
	"fmt"
	"github.com/kercylan98/go-log/log"
	"github.com/puzpuzpuz/xsync/v3"
)

type ProcessRegistry interface {
	// Meta 获取注册表的元数据
	Meta() Meta

	// Register 注册一个进程到注册表
	Register(process Process) (err error)

	// Unregister 从注册表中注销一个进程
	Unregister(operator, target ProcessId)

	// Get 获取一个进程
	Get(meta ProcessId) (process Process, err error)
}

type processRegistryConfig struct {
	Meta          Meta
	Daemon        Process
	LoggerProvide log.Provider
}

func newProcessRegistry(config processRegistryConfig) ProcessRegistry {
	return &processRegistryImpl{
		config:    config,
		processes: xsync.NewMapOf[Path, Process](),
	}
}

type processRegistryImpl struct {
	config    processRegistryConfig
	processes *xsync.MapOf[Path, Process] // 用于存储所有进程的映射表
}

func (i *processRegistryImpl) Get(id ProcessId) (process Process, err error) {
	if id == nil {
		return i.daemon()
	}

	// 通过缓存加载
	cache, implCache := id.(ProcessIdCache)
	if implCache {
		if process = cache.Load(); process != nil {
			lifecycle, impl := process.(ProcessLifecycle)
			if impl {
				if !lifecycle.Terminated() {
					return process, nil
				}
				cache.Store(nil)
			} else {
				// 假设没有实现 ProcessLifecycle 接口的进程是永远不会被终止的，因此不需要检查是否终止
				return process, nil
			}
		}
	}

	// 本地注册表加载
	var loaded bool
	if process, loaded = i.processes.Load(id.Path()); loaded {
		if implCache {
			cache.Store(process)
		}
		return process, nil
	}

	return i.daemon()
}

func (i *processRegistryImpl) daemon() (process Process, err error) {
	if i.config.Daemon == nil {
		return nil, fmt.Errorf("process not found")
	}
	return i.config.Daemon, nil
}

func (i *processRegistryImpl) Meta() Meta {
	return i.config.Meta
}

func (i *processRegistryImpl) Register(process Process) (err error) {
	path := process.GetID().Path()
	if target, loaded := i.processes.LoadOrStore(path, process); loaded {
		if target == process {
			return nil
		}
		return fmt.Errorf("process %s already exists", path)
	}
	logger := i.config.LoggerProvide.Provide()
	logger.Debug("*processRegistryImpl.Register", log.String("event", "register"), log.String("process", path))
	switch cast := process.(type) {
	case ProcessLifecycle:
		logger.Debug("*processRegistryImpl.Register", log.String("event", "initialize"), log.String("process", path))
		cast.Initialize()
	}
	return nil
}

func (i *processRegistryImpl) Unregister(operator, target ProcessId) {
	if target == nil {
		return
	}
	logger := i.config.LoggerProvide.Provide()
	logger.Debug("*processRegistryImpl.Unregister", log.String("event", "unregister"), log.String("process", target.Path()))
	if targetProcess, loaded := i.processes.LoadAndDelete(target.Path()); loaded {
		switch cast := targetProcess.(type) {
		case ProcessLifecycle:
			if operator == nil {
				operator = target
			}
			logger.Debug("*processRegistryImpl.Unregister", log.String("event", "terminate"), log.String("process", target.Path()))
			cast.Terminate(operator)
		}
	}
}
