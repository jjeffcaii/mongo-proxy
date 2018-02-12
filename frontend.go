package pxmgo

import (
	"errors"
	"log"
	"net"
)

type implFrontend struct {
	addr     string
	listener net.Listener
}

func (p *implFrontend) Serve(handler func(Context)) error {
	if p.listener != nil {
		return errors.New("listener has been created already")
	}
	listen, err := net.Listen("tcp", p.addr)
	if err != nil {
		return err
	}
	defer listen.Close()
	p.listener = listen
	for {
		c, err := p.listener.Accept()
		if err != nil {
			log.Println("accept connection failed:", err)
			break
		}
		go func() {
			ctx := newContext(c)
			defer ctx.Close()
			handler(ctx)
		}()
	}
	return nil
}

func (p *implFrontend) Close() error {
	if p.listener == nil {
		return nil
	}
	if err := p.listener.Close(); err != nil {
		return err
	}
	p.listener = nil
	return nil
}

func NewServer(addr string) Endpoint {
	return &implFrontend{addr: addr}
}
