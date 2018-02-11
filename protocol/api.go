package protocol

import (
	"bytes"

	"github.com/sbunce/bson"
)

type OpCode int32

const (
	HeaderLength = 16

	OpCodeReply      OpCode = 1
	OpCodeMsg        OpCode = 1000
	OpCodeUpdate     OpCode = 2001
	OpCodeInsert     OpCode = 2002
	OpReserved       OpCode = 2003
	OpCodeQuery      OpCode = 2004
	OpCodeGetMore    OpCode = 2005
	OpCodeDel        OpCode = 2006
	OpCodeKillCursor OpCode = 2007
	OpCodeCmd        OpCode = 2010
	OpCodeCmdReply   OpCode = 2011
)

type Document = bson.Slice
type Pair = bson.Pair

type Message interface {
	Buffered
	Header() *Header
}

type DatabaseSupport interface {
	GetDatabase() *string
}

type Buffered interface {
	Append(buffer *bytes.Buffer) (int, error)
	Encode() ([]byte, error)
	Decode(bs []byte) error
}
