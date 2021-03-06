package protocol

import (
	"bytes"
	"strings"
)

type OpQuery struct {
	*Op
	Flags                int32
	FullCollectionName   string
	NumberToSkip         int32
	NumberToReturn       int32
	Query                Document
	ReturnFieldsSelector Document
}

func (p *OpQuery) TableName() (*TableName, bool) {
	sp := strings.Split(p.FullCollectionName, ".")
	if len(sp) == 2 {
		return &TableName{sp[0], sp[1]}, true
	}
	return nil,false
}

func (p *OpQuery) Encode() ([]byte, error) {
	bf := bytes.Buffer{}
	if _, err := p.Append(&bf); err == nil {
		return bf.Bytes(), nil
	} else {
		return nil, err
	}
}

func (p *OpQuery) Decode(bs []byte) error {
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
	v4 := readInt32(bs, offset)
	offset += 4
	v5, size, err := readDocument(bs, offset)
	if err != nil {
		return err
	}
	offset += size
	var v6 Document
	if offset < totals {
		v6, size, err = readDocument(bs, offset)
		if err != nil {
			return err
		}
		offset += size
	}
	if offset != totals {
		return &errMessageOffset{offset, totals}
	}
	p.OpHeader = v0
	p.Flags = v1
	p.FullCollectionName = v2
	p.NumberToSkip = v3
	p.NumberToReturn = v4
	p.Query = v5
	p.ReturnFieldsSelector = v6
	return nil
}

func (p *OpQuery) Append(buffer *bytes.Buffer) (int, error) {
	cache := &bytes.Buffer{}
	wrote, err := newWriter(cache).
		writeInt32(p.Flags).
		writeString(p.FullCollectionName).
		writeInt32(p.NumberToSkip).
		writeInt32(p.NumberToReturn).
		writeDocument(p.Query).
		writeDocument(p.ReturnFieldsSelector).
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

func NewOpQuery() *OpQuery {
	return &OpQuery{
		Op: &Op{},
	}
}
