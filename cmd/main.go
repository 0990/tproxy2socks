package main

import (
	"flag"
	"fmt"
	"github.com/0990/ipt2socks"
	"github.com/sirupsen/logrus"
	"net/url"
	"os"
	"os/signal"
	"strings"
)

var proxy = flag.String("proxy", "", "Use this proxy [protocol://]host[:port]")
var listen = flag.String("listen", "", "listen addr")
var udpTimeout = flag.Int("udptimeout", 60, "udp timeout second")

func main() {
	flag.Parse()

	cfg, err := parseCfg(*proxy, *listen, *udpTimeout)
	if err != nil {
		logrus.Fatal(err)
	}

	server, err := ipt2socks.NewServer(cfg)
	if err != nil {
		logrus.Fatalln(err)
	}
	err = server.Run()
	if err != nil {
		logrus.Fatalln(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	s := <-c
	fmt.Println("quit,Got signal:", s)
}

func parseCfg(proxy string, listen string, udpTimeout int) (ipt2socks.Config, error) {
	proxyAddr, err := parseProxy(proxy)
	if err != nil {
		return ipt2socks.Config{}, err
	}

	return ipt2socks.Config{
		ProxyAddr:  proxyAddr,
		ListenAddr: listen,
		UDPTimeout: int32(udpTimeout),
	}, nil
}

func parseProxy(s string) (string, error) {
	if !strings.Contains(s, "://") {
		s = fmt.Sprintf("%s://%s", "socks5" /* default protocol */, s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}

	protocol := strings.ToLower(u.Scheme)

	switch protocol {
	case "socks5":
		return u.Host, nil
	default:
		return "", fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
