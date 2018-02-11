package protocol

import (
	"bytes"
)

type OpInsert struct {
	*Op
	Flags              int32
	FullCollectionName string
	Documents          []Document
}

func (p *OpInsert) Append(buffer *bytes.Buffer) (int, error) {
	cache := &bytes.Buffer{}
	writer := newWriter(cache).writeInt32(p.Flags).writeString(p.FullCollectionName)
	for _, it := range p.Documents {
		writer.writeDocument(it)
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

func (p *OpInsert) Encode() ([]byte, error) {
	bf := &bytes.Buffer{}
	if _, err := p.Append(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (p *OpInsert) Decode(bs []byte) error {
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
	v2 := readString(bs, offset)
	offset += len(v2) + 1
	v3 := make([]Document, 0)
	for ; offset < totals; {
		foo, size, err := readDocument(bs, offset)
		if err != nil {
			return err
		}
		offset += size
		v3 = append(v3, foo)
	}
	if offset != totals {
		return &errMessageOffset{offset, totals}
	}
	p.OpHeader = v0
	p.Flags = v1
	p.FullCollectionName = v2
	p.Documents = v3
	return nil
}

func NewOpInsert() *OpInsert {
	return &OpInsert{
		Op: &Op{},
	}
}
