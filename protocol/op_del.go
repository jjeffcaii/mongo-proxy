package protocol

import (
	"bytes"
)

type OpDelete struct {
	*Op
	Zero               int32
	FullCollectionName string
	Flags              int32
	Selector           Document
}

func (p *OpDelete) Encode() ([]byte, error) {
	bf := &bytes.Buffer{}
	if _, err := p.Append(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (p *OpDelete) Append(buffer *bytes.Buffer) (int, error) {
	cache := &bytes.Buffer{}
	wrote, err := newWriter(cache).
		writeInt32(p.Zero).
		writeString(p.FullCollectionName).
		writeInt32(p.Flags).
		writeDocument(p.Selector).
		end()
	if err != nil {
		return 0, err
	}
	old := p.Header.MessageLength
	wrote += HeaderLength
	p.Header.MessageLength = int32(wrote)
	defer func() {
		p.Header.MessageLength = old
	}()
	bf := &bytes.Buffer{}
	if _, err := p.Header.Append(bf); err != nil {
		return 0, err
	}
	if _, err := cache.WriteTo(bf); err != nil {
		return 0, err
	}
	return wrote, nil
}

func (p *OpDelete) Decode(bs []byte) error {
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
	v3 := readInt32(bs, offset)
	offset += 4
	v4, size, err := readDocument(bs, offset)
	if err != nil {
		return err
	}
	offset += size
	if offset != totals {
		return &errMessageOffset{offset, totals}
	}
	p.Header = v0
	p.Zero = v1
	p.FullCollectionName = v2
	p.Flags = v3
	p.Selector = v4
	return nil
}
