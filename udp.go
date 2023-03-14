package ipt2socks

import (
	"fmt"
	"github.com/0990/ipt2socks/syncx"
	"github.com/0990/ipt2socks/tproxy"
	"github.com/sirupsen/logrus"
	"net"
	"time"
)

const socketBufSize = 64 * 1024

// send: client->relayer->sender->remote
// receive: client<-relayer<-sender<-remote
func runUDPRelayServer(listenAddr *net.UDPAddr, proxyDialer proxyDialer, timeout time.Duration) {
	relayer, err := tproxy.ListenUDP(listenAddr.String())
	if err != nil {
		return
	}
	defer relayer.Close()

	var senders SenderMap

	for {
		buf := make([]byte, socketBufSize)
		n, srcAddr, dstAddr, err := tproxy.ReadFromUDP(relayer.(*net.UDPConn), buf)
		if err != nil {
			logrus.WithError(err).Error("tproxy.ReadFromUDP(")
			continue
		}

		data := buf[0:n]
		id := srcAddr.String() + "|" + dstAddr.String()

		worker := &UDPWorker{
			relayer:   relayer,
			timeout:   timeout,
			srcAddr:   srcAddr,
			dstAddr:   dstAddr,
			writeData: make(chan []byte, 100),
			onClear: func() {
				senders.Del(id)
			},
		}

		w, load := senders.LoadOrStore(id, worker)
		if !load {
			w.Run(proxyDialer)
		}

		w.Write(data)
	}
}

type SenderMap struct {
	m syncx.Map[string, *UDPWorker]
}

func (p *SenderMap) Del(key string) *UDPWorker {
	if conn, exist := p.m.Load(key); exist {
		p.m.Delete(key)
		return conn
	}

	return nil
}

func (p *SenderMap) LoadOrStore(key string, worker *UDPWorker) (w *UDPWorker, load bool) {
	return p.m.LoadOrStore(key, worker)
}

type UDPWorker struct {
	srcAddr, dstAddr *net.UDPAddr
	relayer          net.PacketConn
	timeout          time.Duration
	onClear          func()
	writeData        chan []byte

	sender net.Conn
}

func (w *UDPWorker) Run(dialer proxyDialer) {
	go func() {
		err := w.run(dialer)
		if err != nil {
			w.Close()
			logrus.WithError(err).Error("UDPWorker run")
		}
	}()
}

func (w *UDPWorker) run(dialer proxyDialer) error {
	sender, err := dialer.Dial("udp", w.dstAddr.String())
	if err != nil {
		return err
	}

	w.sender = sender

	log := logrus.WithFields(logrus.Fields{
		"srcAddr": w.srcAddr.String(),
		"dstAddr": w.dstAddr.String(),
	})

	go func() {
		defer w.Close()

		err := relayToClient(sender, w.relayer, w.srcAddr, w.timeout)
		if err != nil {
			log.WithError(err).Error("relayToClient")
		}
	}()

	go func() {
		defer w.Close()

		for v := range w.writeData {
			data := v
			err := w.write(data)
			if err != nil {
				log.WithError(err).Error("UDPWorker Write")
				return
			}
		}
	}()
	return nil
}

func (w *UDPWorker) Close() {
	w.onClear()
	w.sender.Close()
}

func (w *UDPWorker) Write(data []byte) {
	w.writeData <- data
}

func (w *UDPWorker) write(data []byte) error {
	n, err := w.sender.Write(data)
	if err != nil {
		return err
	}
	if len(data) != n {
		return fmt.Errorf("Not all bytes [%d < %d] in buffer written to remote[%s]", n, len(data), w.dstAddr.String())
	}
	return nil
}

func relayToClient(receiver net.Conn, relayer net.PacketConn, clientAddr net.Addr, timeout time.Duration) error {
	buf := make([]byte, socketBufSize)
	for {
		receiver.SetReadDeadline(time.Now().Add(timeout))
		n, err := receiver.Read(buf)
		if err != nil {
			return err
		}

		_, err = relayer.WriteTo(buf[0:n], clientAddr)
		if err != nil {
			return err
		}
	}
}
