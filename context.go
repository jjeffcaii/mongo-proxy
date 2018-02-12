package pxmgo

import (
	"bufio"
	"fmt"
	"net"
	"sync/atomic"

	"github.com/jjeffcaii/mongo-proxy/protocol"
)

type errInvalidOp struct {
	code protocol.OpCode
}

func (p *errInvalidOp) Error() string {
	return fmt.Sprintf("bad message with opcode %d", p.code)
}

type implContext struct {
	reqId       int32
	conn        net.Conn
	middlewares []Middleware
	splicer     *splicer
	writer      *bufio.Writer
	queue       chan protocol.Message
}

func (p *implContext) Use(middlewares ...Middleware) Context {
	for _, it := range middlewares {
		if it != nil {
			p.middlewares = append(p.middlewares, it)
		}
	}
	return p
}

func (p *implContext) Next() <-chan protocol.Message {
	return p.queue
}

func (p *implContext) SendMessage(msg protocol.Message) error {
	addrRespTo := &(msg.Header().ResponseTo)
	old := atomic.LoadInt32(addrRespTo)
	atomic.StoreInt32(addrRespTo, p.reqId)
	bs, err := msg.Encode()
	atomic.StoreInt32(addrRespTo, old)
	if err != nil {
		return err
	}
	return p.Send(bs)
}

func (p *implContext) Send(bs []byte) error {
	_, err := p.writer.Write(bs)
	if err != nil {
		return err
	}
	return p.writer.Flush()
}

func (p *implContext) Close() error {
	p.splicer.Close()
	return p.conn.Close()
}

func (p *implContext) nextMessage() (protocol.Message, error) {
	var bs []byte
	data, err := p.splicer.next()
	if err != nil {
		return nil, err
	}
	bs = data.Bytes()
	var msg protocol.Message
	opcode := protocol.ParseOpCode(bs)
	switch opcode {
	case protocol.OpCodeReply:
		msg = protocol.NewOpReply()
		break
	case protocol.OpCodeMsg:
		msg = protocol.NewOpMsg()
		break
	case protocol.OpCodeUpdate:
		msg = protocol.NewOpUpdate()
		break
	case protocol.OpCodeInsert:
		msg = protocol.NewOpInsert()
		break
	case protocol.OpReserved:
		// TODO: RESERVED
		break
	case protocol.OpCodeQuery:
		msg = protocol.NewOpQuery()
		break
	case protocol.OpCodeGetMore:
		msg = protocol.NewOpGetMore()
		break
	case protocol.OpCodeDel:
		msg = protocol.NewOpDelete()
		break
	case protocol.OpCodeKillCursor:
		msg = protocol.NewOpKillCursors()
		break
	case protocol.OpCodeCmd:
		msg = protocol.NewOpCommand()
		break
	case protocol.OpCodeCmdReply:
		msg = protocol.NewOpCommandReply()
		break
	default:
		break
	}
	if msg == nil {
		return nil, &errInvalidOp{opcode}
	}
	if err := msg.Decode(bs); err != nil {
		return nil, err
	}
	p.reqId = msg.Header().RequestID
	for _, it := range p.middlewares {
		err = it.Handle(p, msg)
		if err != nil {
			break
		}
	}
	if err == Ignore {
		return nil, nil
	}
	if err != nil && err != EOF {
		return nil, err
	}
	return msg, nil
}

func newContext(conn net.Conn) Context {
	ctx := &implContext{
		conn:        conn,
		middlewares: make([]Middleware, 0),
		splicer:     NewSplicer(bufio.NewReader(conn)),
		writer:      bufio.NewWriter(conn),
		queue:       make(chan protocol.Message),
	}
	go func(q chan<- protocol.Message) {
		for {
			next, err := ctx.nextMessage()
			if err != nil {
				break
			}
			if next != nil {
				q <- next
			}
		}
		close(q)
	}(ctx.queue)
	return ctx
}
