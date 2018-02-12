package pxmgo

import (
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type staticBackend struct {
	mutex *sync.Mutex
	addr  string
	conn  net.Conn
}

func (p *staticBackend) Serve(handler func(Context)) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.conn != nil {
		return errors.New("endpoint has been fired already")
	}
	tcpConn, err := net.DialTimeout("tcp", p.addr, 15*time.Second)
	if err != nil {
		log.Println("connect backend failed:", err)
		return err
	}
	p.conn = tcpConn
	ctx := newContext(tcpConn)
	defer ctx.Close()
	handler(ctx)
	return io.EOF
}

func (p *staticBackend) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	return nil
}

type BackendOption struct {
}

func NewBackend(addr string, options ...*BackendOption) Endpoint {
	// TODO: support options
	return &staticBackend{
		mutex: &sync.Mutex{},
		addr:  addr,
	}
}
