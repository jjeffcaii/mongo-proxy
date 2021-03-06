package protocol

import (
	"bytes"
	"fmt"

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

type TableName struct {
	Database   string
	Collection string
}

func (p TableName) String() string {
	return fmt.Sprintf("%s.%s", p.Database, p.Collection)
}

type DatabaseSupport interface {
	TableName() (tbl *TableName, ok bool)
}

type Buffered interface {
	Append(buffer *bytes.Buffer) (int, error)
	Encode() ([]byte, error)
	Decode(bs []byte) error
}
