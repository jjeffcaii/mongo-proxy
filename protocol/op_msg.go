package protocol

import (
	"bytes"
)

type OpMsg struct {
	*Op
	Message string
}

func (p *OpMsg) Encode() ([]byte, error) {
	bf := &bytes.Buffer{}
	if _, err := p.Append(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (p *OpMsg) Append(buffer *bytes.Buffer) (int, error) {
	cache := bytes.Buffer{}
	wrote, err := newWriter(&cache).writeString(p.Message).end()
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

func (p *OpMsg) Decode(bs []byte) error {
	v0 := &Header{}
	if err := v0.Decode(bs); err != nil {
		return err
	}
	totals := len(bs)
	if int(v0.MessageLength) != totals {
		return &errMessageLength{int(v0.MessageLength), totals}
	}
	var offset = HeaderLength
	v1 := readString(bs, offset)
	offset += len(v1) + 1
	if offset != totals {
		return &errMessageOffset{offset, totals}
	}
	p.OpHeader = v0
	p.Message = v1
	return nil
}
