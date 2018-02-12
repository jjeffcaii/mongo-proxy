package pxmgo

import (
	"errors"
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
		return err
	}
	p.conn = tcpConn
	go func(h func(Context), c Context) {
		h(c)
	}(handler, newContext(tcpConn))
	return nil
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
