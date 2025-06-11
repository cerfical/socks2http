package addr

import (
	"fmt"
	"net"
	"strconv"
)

func NewAddr(host string, port uint16) *Addr {
	return &Addr{
		Host: host,
		Port: port,
	}
}

func ParseAddr(addr string) (*Addr, error) {
	if addr == "" {
		return NewAddr("", 0), nil
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("split host from port: %w", err)
	}

	portNum, err := ParsePort(port)
	if err != nil {
		return nil, fmt.Errorf("parse port %v: %w", port, err)
	}

	return NewAddr(host, portNum), nil
}

type Addr struct {
	Host string
	Port uint16
}

func (a *Addr) String() string {
	if a.Host == "" && a.Port == 0 {
		return ""
	}
	return net.JoinHostPort(a.Host, strconv.Itoa(int(a.Port)))
}

func (a *Addr) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Addr) UnmarshalText(text []byte) error {
	host, err := ParseAddr(string(text))
	if err != nil {
		return err
	}
	*a = *host
	return nil
}
