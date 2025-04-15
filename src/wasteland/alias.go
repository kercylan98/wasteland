package wasteland

import "github.com/kercylan98/wasteland/src/internal/rpc"

type (
	Address = string
	Path    = string

	Message         = any
	MessagePriority = int32 // 数值越大，优先级越高。

	Codec           = rpc.Codec
	CodecProvider   = rpc.CodecProvider
	CodecProviderFN = rpc.CodecProviderFN
)
