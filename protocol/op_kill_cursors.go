package protocol

import (
  "bytes"
  "fmt"
)

type OpKillCursors struct {
  Header            *Header
  Zero              int32
  NumberOfCursorIDs int32
  CursorIDs         []int64
}

func (p *OpKillCursors) GetHeader() *Header {
  return p.Header
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

func (p *OpKillCursors) Decode(bs []byte) error {
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
  v2 := readInt32(bs, offset)
  offset += 4
  v3 := make([]int64, 0)
  for ; offset < totals; {
    v3 = append(v3, readInt64(bs, offset))
    offset += 8
  }
  if offset != totals {
    return fmt.Errorf("broken message: read=%d, total=%d", offset, totals)
  }
  p.Header = v0
  p.Zero = v1
  p.NumberOfCursorIDs = v2
  p.CursorIDs = v3
  return nil
}
