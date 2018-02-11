package protocol

import (
	"bytes"
)

type OpKillCursors struct {
	*Op
	Zero              int32
	NumberOfCursorIDs int32
	CursorIDs         []int64
}

func (p *OpKillCursors) Encode() ([]byte, error) {
	bf := &bytes.Buffer{}
	if _, err := p.Append(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (p *OpKillCursors) Append(buffer *bytes.Buffer) (int, error) {
	cache := &bytes.Buffer{}
	writer := newWriter(cache).writeInt32(p.Zero).writeInt32(p.NumberOfCursorIDs)
	for _, it := range p.CursorIDs {
		writer.writeInt64(it)
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

func (p *OpKillCursors) Decode(bs []byte) error {
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
	v2 := readInt32(bs, offset)
	offset += 4
	v3 := make([]int64, 0)
	for ; offset < totals; {
		v3 = append(v3, readInt64(bs, offset))
		offset += 8
	}
	if offset != totals {
		return &errMessageOffset{offset, totals}
	}
	p.OpHeader = v0
	p.Zero = v1
	p.NumberOfCursorIDs = v2
	p.CursorIDs = v3
	return nil
}
