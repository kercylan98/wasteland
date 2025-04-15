package wasteland

import (
	"fmt"
	"github.com/kercylan98/go-log/log"
	"github.com/kercylan98/wasteland/src/internal/protobuf/protobuf"
	"github.com/kercylan98/wasteland/src/internal/rpc"
	"github.com/puzpuzpuz/xsync/v3"
	"net"
)

type ProcessRegistry interface {
	// ResourceLocator 获取注册表的资源定位符
	ResourceLocator() ResourceLocator

	// Run 运行注册表
	Run() (err error)

	// Register 注册一个进程到注册表
	Register(process Process) (err error)

	// Unregister 从注册表中注销一个进程
	Unregister(operator, target ResourceLocator)

	// Get 获取一个进程
	Get(meta ResourceLocator) (process Process, err error)

	// Stop 停止注册表
	Stop()

	// GracefulStop 优雅地停止注册表
	GracefulStop()
}

type ProcessRegistryConfig struct {
	Locator           ResourceLocator
	Daemon            Process
	LoggerProvide     log.Provider
	CodecProvider     CodecProvider
	RPCMessageBuilder RPCMessageBuilder
}

func NewProcessRegistry(config ProcessRegistryConfig) ProcessRegistry {
	if config.LoggerProvide == nil {
		config.LoggerProvide = log.ProviderFn(log.GetDefault)
	}
	return &processRegistryImpl{
		config:    config,
		processes: xsync.NewMapOf[Path, Process](),
	}
}

type processRegistryImpl struct {
	config    ProcessRegistryConfig
	processes *xsync.MapOf[Path, Process] // 用于存储所有进程的映射表
	rpc       rpc.RPC                     // RPC 服务（仅在 addr 不为空时才创建）
}

func (i *processRegistryImpl) Run() (err error) {
	addr := i.config.Locator.Address()
	if addr != "" {
		rpcConfig := rpc.Config{
			CodecProvider: i.config.CodecProvider,
			Handler:       rpc.HandlerFn(i.rpcMessageHandle),
			Logger:        i.config.LoggerProvide,
		}
		if rpcConfig.Listener, err = net.Listen("tcp", addr); err != nil {
			return err
		}

		i.rpc = rpc.New(rpcConfig)
		return i.rpc.Run()
	}
	return
}

func (i *processRegistryImpl) Stop() {
	if i.rpc != nil {
		i.rpc.Stop()
	}
}

func (i *processRegistryImpl) GracefulStop() {
	if i.rpc != nil {
		i.rpc.GracefulStop()
	}
}

func (i *processRegistryImpl) Get(id ResourceLocator) (process Process, err error) {
	if id == nil {
		return i.daemon()
	}

	// 通过缓存加载
	cache, implCache := id.(ResourceCache)
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

	// 远程解析
	if id.Address() != i.config.Locator.Address() {
		process = newRPCProcess(i, id)
		if implCache {
			cache.Store(process)
		}
		return process, nil
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

func (i *processRegistryImpl) ResourceLocator() ResourceLocator {
	return i.config.Locator
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

func (i *processRegistryImpl) Unregister(operator, target ResourceLocator) {
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

func (i *processRegistryImpl) rpcMessageHandle(stream rpc.Stream, message *protobuf.Message_BatchEntry) {
	msg, err := stream.Decode(message.TypeName, message.Message)
	if err != nil {
		i.config.LoggerProvide.Provide().Error("rpcMessageHandle", log.String("event", "decode"), log.Err(err))
		return
	}

	process, err := i.Get(NewResourceLocator(message.ReceiverAddr, message.ReceiverPath))
	if err != nil {
		i.config.LoggerProvide.Provide().Error("rpcMessageHandle", log.String("event", "get"), log.Err(err))
		return
	}

	if handler, cast := process.(ProcessHandler); cast {
		var sender ResourceLocator
		if message.SenderAddr != "" {
			sender = NewResourceLocator(message.SenderAddr, message.SenderPath)
		}
		handler.HandleMessage(sender, message.Priority, msg)
	} else {
		i.config.LoggerProvide.Provide().Warn("rpcMessageHandle", log.String("event", "cast"), log.String("process", message.ReceiverPath))
	}
}
