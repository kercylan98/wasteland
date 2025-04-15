package wasteland

import (
	"github.com/kercylan98/go-log/log"
	"github.com/kercylan98/wasteland/src/internal/protobuf/protobuf"
	"github.com/kercylan98/wasteland/src/internal/rpc"
	"math"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"
)

const (
	rpcProcessStateIdle uint32 = iota
	rpcProcessStateActive
	rpcMessageBatchLimit = 256
)

var (
	_ Process        = (*rpcProcess)(nil)
	_ ProcessHandler = (*rpcProcess)(nil)
)

func newRPCProcess(registry *processRegistryImpl, id ResourceLocator) Process {
	return &rpcProcess{
		id:       id,
		registry: registry,
	}
}

type rpcProcess struct {
	id             ResourceLocator                // 指向远端进程的 ID
	registry       *processRegistryImpl           // 注册表
	stream         rpc.Stream                     // 远程流
	batch          []*protobuf.Message_BatchEntry // 批量消息
	rw             sync.RWMutex                   // 读写锁
	state          atomic.Uint32                  // 状态
	recoveryWaiter sync.WaitGroup                 // 恢复等待组
}

func (r *rpcProcess) GetID() ResourceLocator {
	return r.id
}

func (r *rpcProcess) HandleMessage(sender ResourceLocator, priority MessagePriority, message Message) {
	typeName, buf, err := r.stream.Encode(message)
	if err != nil {
		r.registry.config.LoggerProvide.Provide().Error("remote", log.String("event", "send"), log.String("addr", r.id.Address()), log.String("info", "encode message error"), log.Err(err))
		return
	}

	r.rw.Lock()
	r.batch = append(r.batch, &protobuf.Message_BatchEntry{
		SenderAddr:   sender.Address(),
		SenderPath:   sender.Path(),
		ReceiverAddr: r.id.Address(),
		ReceiverPath: r.id.Path(),
		TypeName:     typeName,
		Message:      buf,
	})
	r.rw.Unlock()
	r.activation()
}

func (r *rpcProcess) activation() {
	if r.state.CompareAndSwap(rpcProcessStateIdle, rpcProcessStateActive) {
		go func() {
			for {
				stop := r.send()
				r.state.Store(rpcProcessStateIdle)
				if stop {
					break
				}
				r.rw.RLock()
				empty := len(r.batch) == 0
				r.rw.RUnlock()
				if empty {
					break
				} else if !r.state.CompareAndSwap(rpcProcessStateIdle, rpcProcessStateActive) {
					break
				}
			}
		}()
	}
}

func (r *rpcProcess) send() (stop bool) {
	const baseDelay = time.Millisecond * 100
	const maxDelay = time.Second
	const multiplier = 2.0
	const randomization = 0.5

	var logger = r.registry.config.LoggerProvide.Provide()

	for {
		r.rw.Lock()
		n := len(r.batch)
		var batch []*protobuf.Message_BatchEntry
		if n < rpcMessageBatchLimit {
			batch = r.batch
			r.batch = nil
		} else {
			batch = r.batch[:rpcMessageBatchLimit]
			r.batch = r.batch[rpcMessageBatchLimit:]
		}
		r.rw.Unlock()
		if len(batch) == 0 {
			break
		}

		// 尝试发送消息
		var stream = r.stream
		var err error
		var once sync.Once
		var failCount int
		var m *protobuf.Message

		for {
			// 尚未持有远程流，尝试获取
			if stream == nil {
				stream, err = r.registry.rpc.Get(r.id.Address())
				r.stream = stream
				if err != nil {
					logger.Error("remote", log.String("event", "send"), log.String("addr", r.id.Address()), log.String("info", "get remote stream error"), log.Err(err))
					break
				}
			}

			// 如果获取远程流失败或者发送消息失败，进入下一次重试
			if err == nil {
				if m == nil {
					m = &protobuf.Message{MessageType: &protobuf.Message_Batch_{Batch: &protobuf.Message_Batch{Entries: batch}}}
				}

				if err = stream.Send(m); err != nil {
					r.stream = nil
					r.registry.rpc.CloseStream(stream)
				} else {
					break // 发送成功，退出循环
				}
			}

			// 发送失败，等待循环重试，阻塞后续消息，避免消息大量堆积
			once.Do(func() {
				r.recoveryWaiter.Add(1)
			})

			// 退避重试，该重试没有必要存在上限，因为即便是达到次数上限跳出循环，也会被下一次消息发送重新激活而继续阻塞
			failCount++
			delay := float64(baseDelay) * math.Pow(multiplier, float64(failCount))
			jitter := (rand.Float64() - 0.5) * randomization * float64(baseDelay)
			sleepDuration := time.Duration(delay + jitter)
			if sleepDuration > maxDelay {
				sleepDuration = maxDelay
			}

			logger.Error("remote", log.String("event", "send"), log.String("addr", r.id.Address()), log.Int("times", failCount), log.String("info", "send message error"), log.Err(err))

			time.Sleep(sleepDuration)
		}

		// 解除等待
		if failCount > 0 {
			r.recoveryWaiter.Done()
		}
	}
	return
}
