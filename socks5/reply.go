package socks5

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"

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
	BindAddr addr.Host
}

func (r *Reply) Write(w io.Writer) error {
	// Encode the request header
	bytes := []byte{VersionCode, byte(r.Status), 0x00}

	if ip4, ok := r.BindAddr.ToIPv4(); ok {
		// Append bind IPv4 address
		bytes = append(bytes, ip4AddrType)
		bytes = append(bytes, ip4[:]...)
	} else {
		n := len(r.BindAddr.Hostname)
		if n > math.MaxUint8 {
			return fmt.Errorf("encode bind address: %w (%v)", ErrHostnameTooLong, n)
		}

		// Append bind hostname
		bytes = append(bytes, hostnameAddrType)
		bytes = append(bytes, byte(n))
		bytes = append(bytes, []byte(r.BindAddr.Hostname)...)
	}

	// Append bind port
	bytes = binary.BigEndian.AppendUint16(bytes, r.BindAddr.Port)

	_, err := w.Write(bytes)
	return err
}
