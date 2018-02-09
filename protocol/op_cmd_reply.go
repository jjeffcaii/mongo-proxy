package protocol

import (
  "bytes"
  "fmt"
)

type OpCommandReply struct {
  Header       *Header
  Metadata     Document
  CommandReply Document
  OutputDocs   []Document
}

func (p *OpCommandReply) GetHeader() *Header {
  return p.Header
}

func (p *OpCommandReply) Encode() ([]byte, error) {
  bf := &bytes.Buffer{}
  if _, err := p.Append(bf); err != nil {
    return nil, err
  }
  return bf.Bytes(), nil
}

func (p *OpCommandReply) Append(buffer *bytes.Buffer) (int, error) {
  cache := &bytes.Buffer{}
  writer := newWriter(cache).writeDocument(p.Metadata).writeDocument(p.CommandReply)
  for _, it := range p.OutputDocs {
    writer.writeDocument(it)
  }
  wrote, err := writer.end()
  if err != nil {
    return 0, err
  }
  wrote += HeaderLength
  bf := &bytes.Buffer{}
  old := p.Header.MessageLength
  p.Header.MessageLength = int32(wrote)
  defer func() { p.Header.MessageLength = old }()
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

func (p *OpCommandReply) Decode(bs []byte) error {
  v0 := &Header{}
  if err := v0.Decode(bs); err != nil {
    return err
  }
  totals := len(bs)
  if int(v0.MessageLength) != totals {
    return fmt.Errorf("broken message: want=%d, actually=%d", v0.MessageLength, totals)
  }
  offset := HeaderLength
  v1, size, err := readDocument(bs, offset)
  if err != nil {
    return err
  }
  offset += size
  v2, size, err := readDocument(bs, offset)
  if err != nil {
    return err
  }
  offset += size
  v3 := make([]Document, 0)
  for ; offset < totals; {
    v, l, err := readDocument(bs, offset)
    if err != nil {
      return err
    }
    offset += l
    v3 = append(v3, v)
  }
  if offset != totals {
    return fmt.Errorf("broken message: read=%d, total=%d", offset, totals)
  }
  p.Header = v0
  p.Metadata = v1
  p.CommandReply = v2
  p.OutputDocs = v3
  return nil
}
