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
	Unregister(operator, meta ProcessId)
}

type processRegistryConfig struct {
	Meta          Meta
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

func (i *processRegistryImpl) Unregister(operator, meta ProcessId) {
	if meta == nil {
		return
	}
	logger := i.config.LoggerProvide.Provide()
	logger.Debug("*processRegistryImpl.Unregister", log.String("event", "unregister"), log.String("process", meta.Path()))
	if target, loaded := i.processes.LoadAndDelete(meta.Path()); loaded {
		switch cast := target.(type) {
		case ProcessLifecycle:
			if operator == nil {
				operator = meta
			}
			logger.Debug("*processRegistryImpl.Unregister", log.String("event", "terminate"), log.String("process", meta.Path()))
			cast.Terminate(operator)
		}
	}
}
