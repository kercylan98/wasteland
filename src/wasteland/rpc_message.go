package wasteland

import (
	"encoding/gob"
)

var defaultRPCMessageBuilder = RPCMessageBuilderFN(func(sender, target, agent ResourceLocator, priority MessagePriority, typeName string, message []byte) RPCMessage {
	return &rpcMessage{
		S: sender,
		T: target,
		A: agent,
		P: priority,
		M: message,
		N: typeName,
	}
})

func DefaultRPCMessageBuilder() RPCMessageBuilder {
	return defaultRPCMessageBuilder
}

func init() {
	gob.RegisterName("w:resourceLocator", &resourceLocator{})
	gob.RegisterName("w:rpcMessage", &rpcMessage{})
}

type RPCMessage interface {
	// Sender 返回消息的发送者
	Sender() ResourceLocator

	// Target 返回消息的目标
	Target() ResourceLocator

	// Agent 返回消息的代理
	Agent() ResourceLocator

	// Priority 返回消息的优先级
	Priority() MessagePriority

	// Message 返回消息体
	Message() []byte

	// MessageType 返回消息类型
	MessageType() string
}

type RPCMessageBuilder interface {
	// Build 构建一个新的 RPC 消息
	Build(sender, target, agent ResourceLocator, priority MessagePriority, typeName string, message []byte) RPCMessage
}

type RPCMessageBuilderFN func(sender, target, agent ResourceLocator, priority MessagePriority, typeName string, message []byte) RPCMessage

func (fn RPCMessageBuilderFN) Build(sender, target, agent ResourceLocator, priority MessagePriority, typeName string, message []byte) RPCMessage {
	return fn(sender, target, agent, priority, typeName, message)
}

type rpcMessage struct {
	S ResourceLocator
	T ResourceLocator
	A ResourceLocator
	P MessagePriority
	N string
	M []byte
}

func (r *rpcMessage) Sender() ResourceLocator {
	return r.S
}

func (r *rpcMessage) Target() ResourceLocator {
	return r.T
}

func (r *rpcMessage) Agent() ResourceLocator {
	return r.A
}

func (r *rpcMessage) Priority() MessagePriority {
	return r.P
}

func (r *rpcMessage) Message() []byte {
	return r.M
}

func (r *rpcMessage) MessageType() string {
	return r.N
}
