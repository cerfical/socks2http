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
