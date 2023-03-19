package tproxy2socks

import (
	"fmt"
	"github.com/0990/tproxy2socks/syncx"
	"github.com/0990/tproxy2socks/tproxy"
	"github.com/sirupsen/logrus"
	"net"
	"syscall"
	"time"
)

const socketBufSize = 64 * 1024

// send: client->relayer->sender->remote
// receive: client<-relayer<-sender<-remote
func runUDPRelayServer(listenAddr *net.UDPAddr, proxyDialer proxyDialer, timeout time.Duration) {
	r, err := tproxy.ListenUDP(listenAddr.String())
	if err != nil {
		return
	}
	defer r.Close()

	relayer := r.(*net.UDPConn)

	rc, err := relayer.SyscallConn()
	if err != nil {
		return
	}

	rc.Control(func(fd uintptr) {
		err = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	})

	var workers WorkerMap
	for {
		buf := make([]byte, socketBufSize)
		n, srcAddr, dstAddr, err := tproxy.ReadFromUDP(relayer, buf)
		if err != nil {
			logrus.WithError(err).Error("tproxy.ReadFromUDP")
			continue
		}

		id := fmt.Sprint("%s|%s", srcAddr.String(), dstAddr.String())

		data := buf[0:n]

		worker := &UDPWorker{
			timeout:     timeout,
			srcAddr:     srcAddr,
			dstAddr:     dstAddr,
			proxyDialer: proxyDialer,
			onClear: func() {
				workers.Del(id)
			},
		}

		w, load := workers.LoadOrStore(id, worker)
		if !load {
			w.Run()
		}

		w.Write(data)

		w.Logger().WithField("len", len(data)).Debug("client udp")
	}
}

type WorkerMap struct {
	m syncx.Map[string, *UDPWorker]
}

func (p *WorkerMap) Del(key string) *UDPWorker {
	if conn, exist := p.m.Load(key); exist {
		p.m.Delete(key)
		return conn
	}

	return nil
}

func (p *WorkerMap) LoadOrStore(key string, worker *UDPWorker) (w *UDPWorker, load bool) {
	return p.m.LoadOrStore(key, worker)
}

type UDPWorker struct {
	srcAddr, dstAddr *net.UDPAddr
	timeout          time.Duration
	proxyDialer      proxyDialer
	writeData        chan []byte
	onClear          func()
}

func (w *UDPWorker) Run() {
	w.writeData = make(chan []byte, 100)

	go func() {
		err := w.run()
		if err != nil {
			w.onClear()
			w.Logger().WithError(err).Error("UDPWorker run")
		}
	}()
}

func (w *UDPWorker) Logger() *logrus.Entry {
	log := logrus.WithFields(logrus.Fields{
		"src": w.srcAddr.String(),
		"dst": w.dstAddr.String(),
	})
	return log
}

func (w *UDPWorker) Write(data []byte) {
	if len(w.writeData) > 90 {
		logrus.Warn("UDPWorker writeData reach limit")
	}
	w.writeData <- data
}

func (w *UDPWorker) run() error {
	now := time.Now()
	proxy, err := w.proxyDialer.Dial("udp", w.dstAddr.String())
	if err != nil {
		return err
	}

	clientWriter, err := tproxy.DialUDP("udp", w.dstAddr, w.srcAddr)
	if err != nil {
		return err
	}

	log := w.Logger()
	defer func() {
		log.Debug("UDPWorker close")

		w.onClear()
		clientWriter.Close()
		proxy.Close()
		close(w.writeData)
	}()

	go func() {
		for data := range w.writeData {
			n, err := proxy.Write(data)
			if err != nil {
				log.WithError(err).Error("proxy.Write")
				return
			}
			if len(data) != n {
				err := fmt.Errorf("Not all bytes [%d < %d] in buffer written to remote[%s]", n, len(data), w.dstAddr.String())
				log.WithError(err).Error("proxy.Write")
				return
			}
		}
	}()

	go func() {
		buf := make([]byte, socketBufSize)
		for {
			n, addr, err := clientWriter.ReadFromUDP(buf)
			if err != nil {
				if !isNetCloseErr(err) {
					log.WithError(err).Error("ReadFromUDP")
				}
				return
			}
			w.Write(buf[:n])

			log.WithField("len", n).Debug("client more udp")

			if addr.String() != w.srcAddr.String() {
				logrus.WithField("addr", addr.String()).Error("ReadFromUDP addr.String()!=w.srcAddr")
			}
		}
	}()

	log.Debug("UDPWorker start")

	_, _, err = relayToClient(proxy, clientWriter, w.srcAddr, w.timeout, log)
	if err != nil {
		if isNetTimeoutErr(err) {
		} else {
			log.WithField("elapseSec", time.Since(now).Seconds()).WithError(err).Error("relayToClient")
		}
	}
	return nil
}

func relayToClient(receiver net.Conn, writer *net.UDPConn, clientAddr *net.UDPAddr, timeout time.Duration, log *logrus.Entry) (int32, time.Time, error) {
	var lastReadTime time.Time

	var readCount int32
	for {
		buf := make([]byte, socketBufSize)
		receiver.SetReadDeadline(time.Now().Add(timeout))
		n, err := receiver.Read(buf)
		if err != nil {
			return readCount, lastReadTime, err
		}

		readCount++
		lastReadTime = time.Now()

		//log.WithField("len", n).Debug("remote udp")
		_, err = writer.WriteTo(buf[0:n], clientAddr)
		if err != nil {
			return readCount, lastReadTime, err
		}
	}

	return readCount, lastReadTime, nil
}
