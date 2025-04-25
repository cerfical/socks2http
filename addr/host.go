package addr

import (
	"fmt"
	"net"
	"strconv"
)

func NewHost(hostname string, port uint16) *Host {
	return &Host{
		Hostname: hostname,
		Port:     port,
	}
}

func ParseHost(host string) (*Host, error) {
	hostname, port, err := net.SplitHostPort(host)
	if err != nil {
		return nil, err
	}

	portNum, err := ParsePort(port)
	if err != nil {
		return nil, fmt.Errorf("parse port number: %w", err)
	}

	return NewHost(hostname, portNum), nil
}

type Host struct {
	Hostname string
	Port     uint16
}

func (h *Host) String() string {
	port := strconv.Itoa(int(h.Port))
	return net.JoinHostPort(h.Hostname, port)
}

func (h *Host) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *Host) UnmarshalText(text []byte) error {
	host, err := ParseHost(string(text))
	if err != nil {
		return err
	}
	*h = *host
	return nil
}

func (h *Host) ToIPv4() (IPv4, bool) {
	ip := net.ParseIP(h.Hostname)
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

func (h *Host) ResolveToIPv4() (IPv4, error) {
	// Assume localhost, if the hostname is not specified
	if h.Hostname == "" {
		return IPv4{127, 0, 0, 1}, nil
	}

	ip, err := net.ResolveIPAddr("ip4", h.Hostname)
	if err != nil {
		return IPv4{}, err
	}
	return IPv4(ip.IP.To4()), nil
}
