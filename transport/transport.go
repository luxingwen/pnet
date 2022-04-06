package transport

import (
	"errors"
	"net"
	"time"
)

type Transport interface {
	Dial(addr string, dialTimeout time.Duration) (net.Conn, error)
	Listen(port uint16) (net.Listener, error)
	GetNetwork() string
	String() string
}

func NewTransport(protocol string, supportedTransports []Transport) (Transport, error) {
	for _, t := range supportedTransports {
		if protocol == t.String() {
			return t, nil
		}
	}
	return nil, errors.New("Unknown protocol " + protocol)
}
