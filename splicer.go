package pxmgo

import (
  "bufio"
  "bytes"
  "encoding/binary"
  "io"

  "github.com/jjeffcaii/mongo-proxy/protocol"
)

type splicer struct {
  isStop bool
  wants  int
  reader *bufio.Reader
  buffer *bytes.Buffer
}

func (p *splicer) stop() {
  p.isStop = true
}

func (p *splicer) next() (*bytes.Buffer, error) {
  var left = p.wants - p.buffer.Len()
  for i := 0; i < left; i++ {
    if p.isStop {
      return nil, io.EOF
    }
    b, err := p.reader.ReadByte()
    if err == io.EOF || p.isStop {
      return nil, io.EOF
    }
    if err != nil {
      return nil, err
    }
    p.buffer.WriteByte(b)
  }
  if p.buffer.Len() != protocol.HeaderLength {
    old := p.buffer
    // reset
    p.wants = protocol.HeaderLength
    p.buffer = &bytes.Buffer{}
    return old, nil
  }
  b := p.buffer.Bytes()[0:4]
  payloadSize := int(binary.LittleEndian.Uint32(b))
  p.wants += payloadSize - protocol.HeaderLength
  return p.next()
}

func NewSplicer(source *bufio.Reader) *splicer {
  return &splicer{
    wants:  protocol.HeaderLength,
    reader: source,
    buffer: &bytes.Buffer{},
  }
}
