package tproxy

import (
	"context"
	"net"
)

type Listener struct {
	base net.Listener
}

func ListenTCP(addr *net.TCPAddr) (net.Listener, error) {
	var lc net.ListenConfig
	lc.Control = tcpTransparentControl
	l, err := lc.Listen(context.Background(), "tcp", addr.String())
	if err != nil {
		return nil, err
	}
	return &Listener{l}, nil
}

func (l *Listener) Accept() (net.Conn, error) {
	return l.AcceptTProxy()
}

func (l *Listener) AcceptTProxy() (*Conn, error) {
	conn, err := l.base.(*net.TCPListener).AcceptTCP()
	if err != nil {
		return nil, err
	}
	return &Conn{TCPConn: conn}, nil
}

func (l *Listener) Addr() net.Addr {
	return l.base.Addr()
}

func (l *Listener) Close() error {
	return l.base.Close()
}

type Conn struct {
	*net.TCPConn
}

func (c *Conn) DstAddr() *net.TCPAddr {
	return c.LocalAddr().(*net.TCPAddr)
}
