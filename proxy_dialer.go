package tproxy2socks

import (
	"errors"
	"fmt"
	"github.com/0990/socks5"
	"net"
	"net/url"
	"strings"
)

type proxyDialer interface {
	Dial(network, addr string) (net.Conn, error)
}

func newProxyDialer(proxy string, udpTimeout int32) (proxyDialer, error) {
	protocol, addr, err := parseProxy(proxy)
	if err != nil {
		return nil, err
	}

	switch protocol {
	case "socks5":
		return socks5.NewSocks5Client(socks5.ClientCfg{
			ServerAddr: addr,
			UserName:   "",
			Password:   "",
			UDPTimout:  int(udpTimeout),
			TCPTimeout: 60,
		}), nil
	case "socks4":
		return socks5.NewSocks4Client(socks5.ClientCfg{
			ServerAddr: addr,
			UserName:   "",
			Password:   "",
			UDPTimout:  int(udpTimeout),
			TCPTimeout: 60,
		}), nil
	default:
		return nil, errors.New("not support proxy type")
	}
}

func parseProxy(s string) (string, string, error) {
	if !strings.Contains(s, "://") {
		s = fmt.Sprintf("%s://%s", "socks5" /* default protocol */, s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", "", err
	}

	protocol := strings.ToLower(u.Scheme)

	switch protocol {
	case "socks5", "socks4":
		return protocol, u.Host, nil
	default:
		return "", "", fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
