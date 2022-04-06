package multiplexer

import (
	"errors"
	"net"
)

type Multiplexer interface {
	AcceptStream() (net.Conn, error)
	OpenStream() (net.Conn, error)
}

func NewMultiplexer(protocol string, conn net.Conn, isClient bool) (Multiplexer, error) {
	switch protocol {
	case "smux":
		return NewSmux(conn, isClient)
	case "yamux":
		return NewYamux(conn, isClient)
	default:
		return nil, errors.New("Unknown protocol " + protocol)
	}
}
