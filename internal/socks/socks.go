package socks

import (
	"fmt"
	"net"

	"github.com/cerfical/socks2http/internal/addr"
)

func Connect(conn net.Conn, dest *addr.Addr) error {
	ipv4, port, err := resolveAddr(dest)
	if err != nil {
		return fmt.Errorf("resolve address %v: %w", dest, err)
	}

	req := Request{
		Version:  V4,
		Command:  ConnectCommand,
		DestIP:   ipv4,
		DestPort: port,
	}

	if err := req.Write(conn); err != nil {
		return err
	}
	return ReadReply(conn)
}

func resolveAddr(a *addr.Addr) (addr.IPv4Addr, uint16, error) {
	ipv4, err := addr.LookupIPv4(a.Hostname)
	if err != nil {
		return addr.IPv4Addr{}, 0, fmt.Errorf("lookup hostname: %w", err)
	}

	port, err := addr.ParsePort(a.Port)
	if err != nil {
		return addr.IPv4Addr{}, 0, fmt.Errorf("parse port number: %w", err)
	}
	return ipv4, port, nil
}
