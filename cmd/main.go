package main

import (
	"flag"
	"fmt"
	"github.com/0990/tproxy2socks"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

var proxy = flag.String("proxy", "socks5://127.0.0.1:1080", "Use this proxy [protocol://]host[:port]")
var listen = flag.String("listen", "0.0.0.0:60080", "listen addr")
var udpTimeout = flag.Int("udptimeout", 60, "udp timeout second")
var logLevel = flag.String("loglevel", "error", "log level,debug,info,warn,error")

func main() {
	flag.Parse()

	level, err := logrus.ParseLevel(*logLevel)
	if err != nil {
		logrus.Fatalln(fmt.Errorf("loglevel not valid:%w", level))
	}

	logrus.SetLevel(level)

	server, err := tproxy2socks.NewServer(tproxy2socks.Config{
		Proxy:      *proxy,
		ListenAddr: *listen,
		UDPTimeout: int32(*udpTimeout),
	})

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
