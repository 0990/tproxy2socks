package main

import (
	"flag"
	"fmt"
	"github.com/0990/ipt2socks"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
)

var proxy = flag.String("proxy", "socks5://127.0.0.1:1080", "Use this proxy [protocol://]host[:port]")
var listen = flag.String("listen", "0.0.0.0:60080", "listen addr")
var udpTimeout = flag.Int("udptimeout", 60, "udp timeout second")
var verbose = flag.Bool("verbose", false, "print verbose log, affect performance")

func main() {
	flag.Parse()

	if *verbose {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	server, err := ipt2socks.NewServer(ipt2socks.Config{
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
