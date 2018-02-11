package protocol

import (
	"bytes"
)

type OpReply struct {
	*Op
	ResponseFlags  int32
	CursorID       int64
	StartingFrom   int32
	NumberReturned int32
	Documents      []Document
}

func (p *OpReply) Append(buffer *bytes.Buffer) (int, error) {
	cache := &bytes.Buffer{}
	writer := newWriter(cache).
		writeInt32(p.ResponseFlags).
		writeInt64(p.CursorID).
		writeInt32(p.StartingFrom).
		writeInt32(p.NumberReturned)
	for _, doc := range p.Documents {
		writer.writeDocument(doc)
	}
	wrote, err := writer.end()
	if err != nil {
		return 0, err
	}
	old := p.OpHeader.MessageLength
	wrote += HeaderLength
	p.OpHeader.MessageLength = int32(wrote)
	defer func() {
		p.OpHeader.MessageLength = old
	}()
	bf := &bytes.Buffer{}
	if _, err := p.OpHeader.Append(bf); err != nil {
		return 0, err
	}
	if _, err := cache.WriteTo(bf); err != nil {
		return 0, err
	}
	if _, err := bf.WriteTo(buffer); err != nil {
		return 0, err
	}
	return wrote, nil
}

func (p *OpReply) Encode() ([]byte, error) {
	buff := &bytes.Buffer{}
	if _, err := p.Append(buff); err != nil {
		return nil, err
	}
	return buff.Bytes(), nil
}

func (p *OpReply) Decode(bs []byte) error {
	v0 := &Header{}
	if err := v0.Decode(bs); err != nil {
		return err
	}
	totals := len(bs)
	if int(v0.MessageLength) != totals {
		return &errMessageLength{int(v0.MessageLength), totals}
	}
	offset := HeaderLength
	v1 := readInt32(bs, offset)
	offset += 4
	v2 := readInt64(bs, offset)
	offset += 8
	v3 := readInt32(bs, offset)
	offset += 4
	v4 := readInt32(bs, offset)
	offset += 4
	v5 := make([]Document, 0)
	for ; offset < totals; {
		doc, size, err := readDocument(bs, offset)
		if err != nil {
			return err
		}
		offset += size
		v5 = append(v5, doc)
	}
	if offset != totals {
		return &errMessageOffset{offset, totals}
	}
	p.OpHeader = v0
	p.ResponseFlags = v1
	p.CursorID = v2
	p.StartingFrom = v3
	p.NumberReturned = v4
	p.Documents = v5
	return nil
}

func NewOpReply() *OpReply {
	return &OpReply{
		Op: &Op{},
	}
}
