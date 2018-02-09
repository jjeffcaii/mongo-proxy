package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"

	"github.com/sbunce/bson"
)

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
		return fmt.Errorf("at least %d bytes", HeaderLength)
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

type xwriter struct {
	buffer *bytes.Buffer
	wrote  int
}

func (p *xwriter) writeInt64(v int64) *xwriter {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(v))
	wrote, err := p.buffer.Write(b)
	if err != nil {
		panic(err)
	}
	p.wrote += wrote
	return p
}

func (p *xwriter) writeInt32(v int32) *xwriter {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v))
	wrote, err := p.buffer.Write(b)
	if err != nil {
		panic(err)
	}
	p.wrote += wrote
	return p
}

func (p *xwriter) writeString(v string) *xwriter {
	wrote, err := p.buffer.WriteString(v)
	if err != nil {
		panic(err)
	}
	p.wrote += wrote
	if err = p.buffer.WriteByte(0); err != nil {
		panic(err)
	}
	p.wrote++
	return p
}

func (p *xwriter) writeDocument(doc Document) *xwriter {
	if doc == nil {
		return p
	}
	doc.MustEncode()
	b, err := doc.Encode()
	if err != nil {
		panic(err)
	}
	wrote, err := p.buffer.Write(b)
	if err != nil {
		panic(err)
	}
	p.wrote += wrote
	return p
}

func (p *xwriter) end() (int, error) {
	return p.wrote, nil
}

func newWriter(buffer *bytes.Buffer) *xwriter {
	return &xwriter{buffer: buffer}
}

func readInt32(bs []byte, offset int) int32 {
	return int32(binary.LittleEndian.Uint32(bs[offset:offset+4]))
}

func readInt64(bs []byte, offset int) int64 {
	return int64(binary.LittleEndian.Uint64(bs[offset:offset+8]))
}

func readString(bs []byte, offset int) string {
	var i = offset
	for {
		if b := bs[i]; b == 0 {
			break
		}
		i++
	}
	return string(bs[offset:i])
}

func readDocument(bs []byte, offset int) (Document, int, error) {
	l := int(binary.LittleEndian.Uint32(bs[offset:offset+4]))
	slice, err := bson.ReadSlice(bytes.NewReader(bs[offset:offset+l]))
	if err != nil {
		return nil, 0, err
	}
	return slice, l, nil
}

func ParseOpCode(bs []byte) OpCode {
	if l := len(bs); l < 16 {
		log.Printf("cannot parse OpCode: %d bytes is not enough", l)
		return -1
	}
	return OpCode(binary.LittleEndian.Uint32(bs[12:16]))
}

func ToMap(d Document) map[string]interface{} {
	c := make(map[string]interface{})
	for _, p := range d {
		c[p.Key] = p.Val
	}
	return c
}
