package protocol

import (
	"bytes"
)

type OpUpdate struct {
	*Op
	Zero               int32
	FullCollectionName string
	Flags              int32
	Selector           Document
	Update             Document
}

func (p *OpUpdate) Append(buffer *bytes.Buffer) (int, error) {
	cache := &bytes.Buffer{}
	wrote, err := newWriter(cache).
		writeInt32(p.Zero).
		writeString(p.FullCollectionName).
		writeInt32(p.Flags).
		writeDocument(p.Selector).
		writeDocument(p.Update).
		end()
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

func (p *OpUpdate) Encode() ([]byte, error) {
	bf := &bytes.Buffer{}
	if _, err := p.Append(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (p *OpUpdate) Decode(bs []byte) error {
	v0 := &Header{}
	if err := v0.Decode(bs); err != nil {
		return err
	}
	var totals = len(bs)
	if int(v0.MessageLength) != totals {
		return &errMessageLength{int(v0.MessageLength), totals}
	}
	offset := HeaderLength
	v1 := readInt32(bs, offset)
	offset += 4
	v2 := readString(bs, offset)
	offset += len(v2) + 1
	v3 := readInt32(bs, offset)
	offset += 4
	v4, size, err := readDocument(bs, offset)
	if err != nil {
		return err
	}
	offset += size
	v5, size, err := readDocument(bs, offset)
	if err != nil {
		return err
	}
	offset += size
	if offset != totals {
		return &errMessageOffset{offset, totals}
	}
	p.OpHeader = v0
	p.Zero = v1
	p.FullCollectionName = v2
	p.Flags = v3
	p.Selector = v4
	p.Update = v5
	return nil
}
