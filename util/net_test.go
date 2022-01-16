package util

import (
	"net"
	"testing"
)

func TestGetAddr1(t *testing.T) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(l.Addr().String())
	t.Log(net.SplitHostPort(l.Addr().String()))
	l.Close()
}

func TestGetAddr2(t *testing.T) {
	addr, err := net.ResolveTCPAddr("tcp", "0.0.0.0:19653")
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(l.Addr().String())
	t.Log(net.SplitHostPort(l.Addr().String()))

	l.Close()
}
