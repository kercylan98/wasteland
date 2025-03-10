package wasteland

import "net"

type (
	Address   = net.Addr
	Path      = string
	Zone      = uint8
	Version   = uint8
	Timestamp = int64
	Gen       = uint16

	Message         = any
	MessagePriority = int8 // 消息优先级是一个 8 位有符号整数，范围是 -128 到 127，数值越大，优先级越高。
)
