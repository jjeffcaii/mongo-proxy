package protocol

import (
	"bytes"
	"fmt"
)

type OpQuery struct {
	Header               *Header
	Flags                int32
	FullCollectionName   string
	NumberToSkip         int32
	NumberToReturn       int32
	Query                Document
	ReturnFieldsSelector Document
}

func (p *OpQuery) GetDatabase() *string {
	// TODO: extract database
	return nil
}

func (p *OpQuery) GetHeader() *Header {
	return p.Header
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
		return fmt.Errorf("broken message: want=%d, actually=%d", v0.MessageLength, totals)
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
		return fmt.Errorf("broken message: read=%d, total=%d", offset, totals)
	}
	p.Header = v0
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
