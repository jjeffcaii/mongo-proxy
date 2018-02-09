package protocol

import (
  "bytes"

  "github.com/sbunce/bson"
)

type OpCode int32

const (
  OpCodeReply      OpCode = 1
  OpCodeMsg        OpCode = 1000
  OpCodeUpdate     OpCode = 2001
  OpCodeInsert     OpCode = 2002
  RESERVED         OpCode = 2003
  OpCodeQuery      OpCode = 2004
  OpCodeGetMore    OpCode = 2005
  OpCodeDel        OpCode = 2006
  OpCodeKillCursor OpCode = 2007
  OpCodeCmd        OpCode = 2010
  OpCodeCmdReply   OpCode = 2011
)

const (
  HeaderLength = 16
)

type Document = bson.Slice
type Pair = bson.Pair

type Message interface {
  Buffered
  GetHeader() *Header
}

type DatabaseSupport interface {
  GetDatabase() *string
}

type Buffered interface {
  Append(buffer *bytes.Buffer) (int, error)
  Encode() ([]byte, error)
  Decode(bs []byte) error
}

type Header struct {
  MessageLength int32
  RequestID     int32
  ResponseTo    int32
  OpCode        OpCode
}

type documentBuilder struct {
  pairs []Pair
}

func (p *documentBuilder) Set(key string, val interface{}) *documentBuilder {
  p.pairs = append(p.pairs, Pair{Key: key, Val: val,})
  return p
}

func (p *documentBuilder) Build() Document {
  return p.pairs
}

func NewDocument() *documentBuilder {
  return &documentBuilder{}
}
