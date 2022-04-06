package transport

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Address struct {
	Transport Transport
	Host      string
	Port      uint16
}

func NewAddress(protocol, host string, port uint16, supportedTransports []Transport) (*Address, error) {
	transport, err := NewTransport(protocol, supportedTransports)
	if err != nil {
		return nil, err
	}

	addr := &Address{
		Transport: transport,
		Host:      host,
		Port:      port,
	}

	return addr, nil
}

func (addr *Address) String() string {
	return fmt.Sprintf("%s://%s", addr.Transport, addr.ConnRemoteAddr())
}

func (addr *Address) ConnRemoteAddr() string {
	s := addr.Host
	if addr.Port > 0 {
		s += fmt.Sprintf(":%d", addr.Port)
	}
	return s
}

func (addr *Address) Dial(dialTimeout time.Duration) (net.Conn, error) {
	return addr.Transport.Dial(addr.ConnRemoteAddr(), dialTimeout)
}

func Parse(rawAddr string, supportedTransports []Transport) (*Address, error) {
	u, err := url.Parse(rawAddr)
	if err != nil {
		return nil, err
	}

	transport, err := NewTransport(u.Scheme, supportedTransports)
	if err != nil {
		return nil, err
	}

	host, portStr, err := SplitHostPort(u.Host)
	if err != nil {
		return nil, err
	}

	port := 0
	if len(portStr) > 0 {
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, err
		}
	}

	addr := &Address{
		Transport: transport,
		Host:      host,
		Port:      uint16(port),
	}

	return addr, nil
}

func SplitHostPort(hostport string) (host, port string, err error) {
	if !strings.Contains(hostport, ":") {
		hostport += ":"
	}
	return net.SplitHostPort(hostport)
}
