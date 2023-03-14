package ipt2socks

import "net"

type proxyDialer interface {
	Dial(network, addr string) (net.Conn, error)
}
