package pxmgo

import (
	"bytes"
	"errors"
	"io"

	"github.com/jjeffcaii/mongo-proxy/protocol"
)

type Context interface {
	io.Closer
	Use(middlewares ...Middleware) Context
	Send(bs []byte) error
	SendBuffer(bf *bytes.Buffer) error
	SendMessage(msg protocol.Message) error
	Next() <-chan protocol.Message
}

// Endpoint communicate endpoint for routing messages.
type Endpoint interface {
	io.Closer
	Serve(handler func(ctx Context)) error
}

var EOF = io.EOF
var Ignore = errors.New("skip message")

type Identifier struct {
	Username string
	Password string
}

type Authenticator interface {
	Middleware
	Wait() (db *string, ok bool)
}

type Middleware interface {
	// Handle handle request.
	Handle(ctx Context, req protocol.Message) error
}
