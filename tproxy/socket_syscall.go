//go:build linux
// +build linux

package tproxy

import "syscall"

func tcpTransparentControl(network, address string, c syscall.RawConn) error {
	var opErr error
	err := c.Control(func(fd uintptr) {
		opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
	})
	if err != nil {
		return err
	}

	return opErr
}

func udpTransparentControl(network, address string, c syscall.RawConn) error {
	var opErr error
	err := c.Control(func(fd uintptr) {
		opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
		if opErr != nil {
			return
		}
		opErr = syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_RECVORIGDSTADDR, 1)
	})
	if err != nil {
		return err
	}

	return opErr
}
