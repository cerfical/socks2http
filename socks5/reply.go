package socks5

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"

	"github.com/cerfical/socks2http/addr"
)

func ReadReply(r *bufio.Reader) (*Reply, error) {
	if err := checkVersion(r); err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}

	status, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode status: %w", err)
	}

	if err := checkReserved(r); err != nil {
		return nil, fmt.Errorf("decode reserved field: %w", err)
	}

	bindAddr, err := readAddr(r)
	if err != nil {
		return nil, fmt.Errorf("decode bind address: %w", err)
	}

	return &Reply{Status(status), *bindAddr}, nil
}

type Reply struct {
	Status   Status
	BindAddr addr.Addr
}

func (r *Reply) Write(w io.Writer) error {
	// Encode the request header
	bytes := []byte{VersionCode, byte(r.Status), 0x00}

	if ip := net.ParseIP(r.BindAddr.Host); ip != nil {
		var addrType byte
		if ip4 := ip.To4(); ip4 != nil {
			// Bind address is an IPv4 address
			addrType = ip4AddrType
			ip = ip4
		} else {
			// Bind address is an IPv6 address
			addrType = ip6AddrType
		}
		bytes = append(bytes, addrType)
		bytes = append(bytes, ip[:]...)
	} else {
		// Bind address is a hostname
		n := len(r.BindAddr.Host)
		if n > math.MaxUint8 {
			return fmt.Errorf("encode bind address: %w (%v)", ErrHostnameTooLong, n)
		}
		bytes = append(bytes, hostnameAddrType)
		bytes = append(bytes, byte(n))
		bytes = append(bytes, []byte(r.BindAddr.Host)...)
	}

	// Append bind port
	bytes = binary.BigEndian.AppendUint16(bytes, r.BindAddr.Port)

	_, err := w.Write(bytes)
	return err
}
