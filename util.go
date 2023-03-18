package ipt2socks

import "net"

func isNetTimeoutErr(err error) bool {
	if err, ok := err.(net.Error); ok && err.Timeout() {
		return true
	}
	return false
}

func isNetCloseErr(err error) bool {
	if err, ok := err.(*net.OpError); ok && err.Err.Error() == "use of closed network connection" {
		return true
	}
	return false
}
