package ipt2socks

import (
	"github.com/0990/ipt2socks/tproxy"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	listener net.Listener

	tcpListenAddr *net.TCPAddr
	udpListenAddr *net.UDPAddr

	proxyDialer proxyDialer

	cfg Config
}

func NewServer(c Config) (*Server, error) {
	listenAddr := c.ListenAddr
	taddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	uaddr, err := net.ResolveUDPAddr("udp", listenAddr)
	if err != nil {
		return nil, err
	}

	dialer, err := newProxyDialer(c.Proxy, c.UDPTimeout)
	if err != nil {
		return nil, err
	}

	return &Server{
		tcpListenAddr: taddr,
		udpListenAddr: uaddr,
		proxyDialer:   dialer,
		cfg:           c,
	}, nil
}

func (s *Server) Run() error {
	err := s.listen()
	if err != nil {
		return err
	}
	go s.serve()
	go runUDPRelayServer(s.udpListenAddr, s.proxyDialer, time.Duration(s.cfg.UDPTimeout)*time.Second)
	return nil
}

func (s *Server) listen() error {
	l, err := tproxy.ListenTCP(s.tcpListenAddr)
	if err != nil {
		return err
	}
	s.listener = l
	return nil
}

func (s *Server) serve() {
	var tempDelay time.Duration

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			logrus.WithError(err).Error("HandleListener Accept")
			if ne, ok := err.(*net.OpError); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				logrus.Errorf("http: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return
		}
		go s.connHandler(conn)
	}
}

func (s *Server) connHandler(conn net.Conn) {
	defer conn.Close()
	dst := conn.LocalAddr()

	log := logrus.WithFields(logrus.Fields{
		"dst": dst.String(),
		"src": conn.RemoteAddr().String(),
	})

	proxyConn, err := s.proxyDialer.Dial("tcp", dst.String())
	if err != nil {
		log.WithError(err).Error("sock5 dial fail")
		return
	}
	defer proxyConn.Close()

	log.Debug("proxy success")

	var streamWait sync.WaitGroup
	streamWait.Add(2)

	streamConn := func(dst io.Writer, src io.Reader) {
		io.Copy(dst, src)
		streamWait.Done()
	}

	go streamConn(proxyConn, conn)
	go streamConn(conn, proxyConn)

	streamWait.Wait()
}
