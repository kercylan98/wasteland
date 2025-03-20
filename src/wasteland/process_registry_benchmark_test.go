package wasteland_test

import (
	"github.com/kercylan98/wasteland/src/wasteland"
	"testing"
)

type TestProcess struct {
	id wasteland.ProcessId
}

func (t *TestProcess) GetID() wasteland.ProcessId {
	return t.id
}

func BenchmarkProcessRegistryImpl_Get(b *testing.B) {
	registry := wasteland.NewProcessRegistry(wasteland.ProcessRegistryConfig{
		Meta:          wasteland.NewMeta("127.0.0.1:7777"),
		Daemon:        nil,
		LoggerProvide: nil,
	})

	pid := wasteland.NewProcessId(registry.Meta(), "/")

	if err := registry.Register(&TestProcess{id: pid}); err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.Get(pid)
	}
	b.StopTimer()
}
