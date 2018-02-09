package pxmgo

import (
  "bytes"
  "errors"
  "io"

  "github.com/jjeffcaii/go-mongo-proxy/protocol"
)

type Handler func(context Context)

type Context interface {
  Use(middleware Middleware, allow ...protocol.OpCode) Context
  Send(bs []byte) error
  SendBuffer(bf *bytes.Buffer) error
  Next() (protocol.Message, error)
  Close() error
}

// communication endpoint for routing messages.
type Endpoint interface {
  io.Closer
  Serve(handler Handler) error
}

// TODO: optimize -> 链式状态管理太乱了 :-(
var EOF = io.EOF
var END = errors.New("MIDDLEWARE_END")

type OnRes = func(msg protocol.Message) error
type OnNext = func(err error)
type Middleware = func(req protocol.Message, res OnRes, next OnNext)
