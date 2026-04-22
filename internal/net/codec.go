
package net

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	MaxMessageSize = 1024 * 1024 * 10
	HeaderSize      = 8
)

type Message struct {
	MsgID uint32
	Data  []byte
}

type Codec struct {
}

func NewCodec() *Codec {
	return &Codec{}
}

func (c *Codec) Encode(msg *Message) ([]byte, error) {
	if msg == nil {
		return nil, errors.New("message is nil")
	}

	dataLen := len(msg.Data)
	if dataLen > MaxMessageSize {
		return nil, errors.New("message too large")
	}

	buf := make([]byte, HeaderSize+dataLen)
	binary.BigEndian.PutUint32(buf[0:4], msg.MsgID)
	binary.BigEndian.PutUint32(buf[4:8], uint32(dataLen))
	copy(buf[HeaderSize:], msg.Data)

	return buf, nil
}

func (c *Codec) Decode(reader io.Reader) (*Message, error) {
	header := make([]byte, HeaderSize)
	if _, err := io.ReadFull(reader, header); err != nil {
		return nil, err
	}

	msgID := binary.BigEndian.Uint32(header[0:4])
	dataLen := binary.BigEndian.Uint32(header[4:8])

	if dataLen > MaxMessageSize {
		return nil, errors.New("message too large")
	}

	data := make([]byte, dataLen)
	if _, err := io.ReadFull(reader, data); err != nil {
		return nil, err
	}

	return &Message{
		MsgID: msgID,
		Data:  data,
	}, nil
}
