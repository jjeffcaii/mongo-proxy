package protocol

import (
	"bytes"
	"encoding/binary"
)

type Header struct {
	MessageLength int32
	RequestID     int32
	ResponseTo    int32
	OpCode        OpCode
}

func (p *Header) Encode() ([]byte, error) {
	bf := &bytes.Buffer{}
	_, err := newWriter(bf).
		writeInt32(p.MessageLength).
		writeInt32(p.RequestID).
		writeInt32(p.ResponseTo).
		writeInt32(int32(p.OpCode)).
		end()
	if err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (p *Header) Decode(bs []byte) error {
	if len(bs) < HeaderLength {
		return errHeaderLength
	}
	p.MessageLength = int32(binary.LittleEndian.Uint32(bs[:4]))
	p.RequestID = int32(binary.LittleEndian.Uint32(bs[4:8]))
	p.ResponseTo = int32(binary.LittleEndian.Uint32(bs[8:12]))
	p.OpCode = OpCode(binary.LittleEndian.Uint32(bs[12:HeaderLength]))
	return nil
}

func (p *Header) Append(buffer *bytes.Buffer) (int, error) {
	return newWriter(buffer).
		writeInt32(p.MessageLength).
		writeInt32(p.RequestID).
		writeInt32(p.ResponseTo).
		writeInt32(int32(p.OpCode)).
		end()
}
