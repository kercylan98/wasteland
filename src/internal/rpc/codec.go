package rpc

import (
	"bytes"
	"encoding/gob"
)

func RegisterType(t any) {
	gob.Register(t)
}

func RegisterName(name string, t any) {
	gob.RegisterName(name, t)
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
