package wasteland_test

import (
	"errors"
	"github.com/kercylan98/go-log/log"
	"github.com/kercylan98/wasteland/src/wasteland"
	"net"
	"testing"
)

func TestProcessRegistryImpl_Register(t *testing.T) {

	registry := wasteland.ExportNewProcessRegistry(wasteland.ExportProcessRegistryConfig{
		Meta: wasteland.ExportNewProcessIdMeta(&net.IPAddr{
			IP: []byte{127, 0, 0, 1},
		}, 1, 1, 0, 0),
		LoggerProvide: log.ProviderFn(func() log.Logger {
			return log.GetDefault()
		}),
	})

	process := &TestFnProcess{
		ID:                   wasteland.ExportNewProcessId(registry.Meta(), "/test"),
		OnInitialize:         nil,
		OnTerminate:          nil,
		OnHandleMessage:      nil,
		OnHandleAgentMessage: nil,
	}

	if err := registry.Register(process); err != nil {
		t.Error(err)
		return
	} else if err = registry.Register(process); err != nil { // SAME PROCESS
		t.Error(err)
		return
	} else if err = registry.Register(&TestFnProcess{
		ID:                   wasteland.ExportNewProcessId(registry.Meta(), "/test"),
		OnInitialize:         nil,
		OnTerminate:          nil,
		OnHandleMessage:      nil,
		OnHandleAgentMessage: nil,
	}); err == nil { // SAME PATH
		t.Error(errors.New("register same path"))
		return
	}

}
