package protocol

import (
	"bytes"
)

type OpCommand struct {
	*Op
	Database    string
	CommandName string
	Metadata    Document
	CommandArgs Document
	InputDocs   []Document
}

func (p *OpCommand) Encode() ([]byte, error) {
	bf := &bytes.Buffer{}
	if _, err := p.Append(bf); err != nil {
		return nil, err
	}
	return bf.Bytes(), nil
}

func (p *OpCommand) Append(buffer *bytes.Buffer) (int, error) {
	cache := &bytes.Buffer{}
	writer := newWriter(cache).
		writeString(p.Database).
		writeString(p.CommandName).
		writeDocument(p.Metadata).
		writeDocument(p.CommandArgs)
	for _, it := range p.InputDocs {
		writer.writeDocument(it)
	}
	wrote, err := writer.end()
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
	if _, err := bf.WriteTo(buffer); err != nil {
		return 0, err
	}
	return wrote, nil
}

func (p *OpCommand) Decode(bs []byte) error {
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
	v2 := readString(bs, offset)
	offset += len(v2) + 1
	v3, size, err := readDocument(bs, offset)
	if err != nil {
		return err
	}
	offset += size
	v4, size, err := readDocument(bs, offset)
	if err != nil {
		return err
	}
	offset += size
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
	p.Header = v0
	p.Database = v1
	p.CommandName = v2
	p.Metadata = v3
	p.CommandArgs = v4
	p.InputDocs = v5
	return nil
}
