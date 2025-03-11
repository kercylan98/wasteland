package rpc_test

import (
	"github.com/kercylan98/go-log/log"
	"github.com/kercylan98/wasteland/src/internal/rpc"
	"net"
	"testing"
	"time"
)

func TestNew(t *testing.T) {

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
		return
	}

	r := rpc.New(rpc.Config{
		Handler: rpc.HandlerFn(func(message []byte) {
			t.Log(string(message))
		}),
		Listener: lis,
		Logger: log.ProviderFn(func() log.Logger {
			return log.GetDefault()
		}),
	})

	r.Run()

	lis2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Error(err)
		return
	}

	r2 := rpc.New(rpc.Config{
		Handler: rpc.HandlerFn(func(message []byte) {
			t.Log(string(message))
		}),
		Listener: lis2,
		Logger: log.ProviderFn(func() log.Logger {
			return log.GetDefault()
		}),
	})

	r2.Run()

	t.Log(r2.Get(lis.Addr().String()))

	time.Sleep(time.Hour)

}
