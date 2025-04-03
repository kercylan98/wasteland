package rpc

import (
	"bytes"
	"encoding/gob"
)

// Codec 是用于对消息发送前和接收后进行编码和解码的接口
type Codec interface {
	// Encode 编码消息
	Encode(m any) (bytes []byte, err error)

	// Decode 解码消息
	Decode(m any, data []byte) (res any, err error)
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

type codec struct {
	encoderBuf *bytes.Buffer
	decoderBuf *bytes.Buffer
	encoder    *gob.Encoder
	decoder    *gob.Decoder
}

func (c *codec) Encode(v any) ([]byte, error) {
	defer c.encoderBuf.Reset()

	if err := c.encoder.Encode(v); err != nil {
		return nil, err
	}
	var data = make([]byte, c.encoderBuf.Len())
	copy(data, c.encoderBuf.Bytes())
	return data, nil
}

func (c *codec) Decode(dst any, data []byte) (t any, err error) {
	c.decoderBuf.Write(data)
	defer c.decoderBuf.Reset()

	if err := c.decoder.Decode(dst); err != nil {
		return t, err
	}
	return dst, nil
}
