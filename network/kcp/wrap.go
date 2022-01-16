package kcp

import (
	"net"

	"github.com/xtaci/kcp-go"
)

// GoTCPServerOptionListen gotcp options WithListen
func GoTCPServerOptionListen(addr string) (ln net.Listener, err error) {
	return kcp.Listen(addr)
}

// GoTCPClientOptionDialer gotcp options  WithClientOptionsDialer
func GoTCPClientOptionDialer(network, addr string) (conn net.Conn, err error) {
	return kcp.Dial(addr)
}
