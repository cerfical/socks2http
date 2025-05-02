package addr

import (
	"fmt"
	"net"
	"strconv"
)

func New(host string, port uint16) *Addr {
	return &Addr{
		Host: host,
		Port: port,
	}
}

func Parse(addr string) (*Addr, error) {
	if addr == "" {
		return New("", 0), nil
	}

	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, fmt.Errorf("split host from port: %w", err)
	}

	portNum, err := ParsePort(port)
	if err != nil {
		return nil, fmt.Errorf("parse port %v: %w", port, err)
	}

	return New(host, portNum), nil
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
	host, err := Parse(string(text))
	if err != nil {
		return err
	}
	*a = *host
	return nil
}

func (a *Addr) ToIPv4() (IPv4, bool) {
	ip := net.ParseIP(a.Host)
	if ip == nil {
		// Not an IP address
		return IPv4{}, false
	}

	ip4 := ip.To4()
	if ip4 == nil {
		// Not an IPv4 address
		return IPv4{}, false
	}
	return IPv4(ip4), true
}

func (a *Addr) ResolveToIPv4() (IPv4, error) {
	// Assume localhost, if the hostname is not specified
	if a.Host == "" {
		return IPv4{127, 0, 0, 1}, nil
	}

	ip, err := net.ResolveIPAddr("ip4", a.Host)
	if err != nil {
		return IPv4{}, err
	}
	return IPv4(ip.IP.To4()), nil
}
