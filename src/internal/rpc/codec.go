package rpc

import (
	"bytes"
	"encoding/gob"
)

func init() {
	gob.Register(&gobCodecMessage{})
}

// Codec 是用于对消息发送前和接收后进行编码和解码的接口
type Codec interface {
	// Encode 编码消息
	Encode(m any) (typeName string, bytes []byte, err error)

	// Decode 解码消息
	Decode(typeName string, data []byte) (m any, err error)
}

// CodecProvider 是用于创建 Codec 的接口
type CodecProvider interface {
	// Provide 创建 Codec
	Provide() Codec
}

// CodecProviderFN 是 CodecProvider 的函数实现
type CodecProviderFN func() Codec

// Provide 实现 CodecProvider 接口
func (fn CodecProviderFN) Provide() Codec {
	return fn()
}

func newCodec() *codec {
	codec := &codec{
		encoderBuf: new(bytes.Buffer),
		decoderBuf: new(bytes.Buffer),
	}
	codec.encoder = gob.NewEncoder(codec.encoderBuf)
	codec.decoder = gob.NewDecoder(codec.decoderBuf)
	return codec
}

type gobCodecMessage struct {
	M any
}

type codec struct {
	encoderBuf *bytes.Buffer
	decoderBuf *bytes.Buffer
	encoder    *gob.Encoder
	decoder    *gob.Decoder
}

func (c *codec) Encode(v any) (typeName string, data []byte, err error) {
	v = &gobCodecMessage{M: v}

	defer c.encoderBuf.Reset()

	if err := c.encoder.Encode(v); err != nil {
		return "", nil, err
	}
	data = make([]byte, c.encoderBuf.Len())
	copy(data, c.encoderBuf.Bytes())
	return "", data, nil
}

func (c *codec) Decode(typeName string, data []byte) (m any, err error) {
	c.decoderBuf.Write(data)
	defer c.decoderBuf.Reset()

	var dst = new(gobCodecMessage)
	if err := c.decoder.Decode(dst); err != nil {
		return nil, err
	}
	return dst.M, nil
}
