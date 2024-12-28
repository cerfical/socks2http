package socks

import (
	"fmt"
	"net"

	"github.com/cerfical/socks2http/internal/addr"
)

func Connect(conn net.Conn, dest *addr.Addr) error {
	ipv4, err := addr.LookupIPv4(dest.Hostname)
	if err != nil {
		return fmt.Errorf("resolve address %v: %w", dest, err)
	}

	req := Request{Header{
		Version:  V4,
		Command:  ConnectCommand,
		DestIP:   ipv4,
		DestPort: dest.Port,
	}, ""}

	if err := req.Write(conn); err != nil {
		return err
	}
	return ReadReply(conn)
}
