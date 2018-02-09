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

func (p *staticBackend) Serve(handler Handler) error {
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
	go handler(newContext(tcpConn))
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

func NewStaticBackend(addr string) Endpoint {
	return &staticBackend{
		mutex: &sync.Mutex{},
		addr:  addr,
	}
}
