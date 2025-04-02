package addr

import (
	"fmt"
	"net"
	"strconv"
)

func NewHost(hostname string, port int) *Host {
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
	Port     int
}

func (h *Host) String() string {
	port := strconv.Itoa(h.Port)
	return net.JoinHostPort(h.Hostname, port)
}
