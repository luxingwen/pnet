package transport

import (
	"fmt"
	"net"
	"time"
)

type TCPTransport struct{}

func NewTCPTransport() *TCPTransport {
	t := &TCPTransport{}
	return t
}

func (t *TCPTransport) Dial(addr string, dialTimeout time.Duration) (net.Conn, error) {
	return net.DialTimeout(t.GetNetwork(), addr, dialTimeout)
}

func (t *TCPTransport) Listen(port uint16) (net.Listener, error) {
	laddr := fmt.Sprintf(":%d", port)
	return net.Listen(t.GetNetwork(), laddr)
}

func (t *TCPTransport) GetNetwork() string {
	return "tcp"
}

func (t *TCPTransport) String() string {
	return "tcp"
}
