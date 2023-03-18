package tproxy

import (
	"context"
	"net"
)

func ListenTCP(addr *net.TCPAddr) (net.Listener, error) {
	var lc net.ListenConfig
	lc.Control = tcpTransparentControl
	l, err := lc.Listen(context.Background(), "tcp", addr.String())
	if err != nil {
		return nil, err
	}
	return l, nil
}
