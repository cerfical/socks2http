package socks5

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/cerfical/socks2http/addr"
)

const VersionCode = 0x05
const (
	ip4AddrType      = 0x01
	hostnameAddrType = 0x03
)

type hexByte byte

func (b hexByte) String() string {
	return fmt.Sprintf("%#02x", byte(b))
}

func checkVersion(r *bufio.Reader) error {
	version, err := r.Peek(1)
	if err != nil {
		return err
	}
	if version[0] != VersionCode {
		return fmt.Errorf("%w (%v)", ErrInvalidVersion, hexByte(version[0]))
	}
	r.Discard(1)
	return nil
}

func checkReserved(r *bufio.Reader) error {
	rsv, err := r.ReadByte()
	if err != nil {
		return err
	}
	if rsv != 0 {
		return fmt.Errorf("%w (%v)", ErrNonZeroReservedField, hexByte(rsv))
	}
	return nil
}

func readAddr(r *bufio.Reader) (*addr.Addr, error) {
	addrType, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode address type: %w", err)
	}

	var host addr.Addr
	switch addrType {
	case ip4AddrType:
		var ip4 addr.IPv4
		if _, err := io.ReadFull(r, ip4[:]); err != nil {
			return nil, fmt.Errorf("decode IPv4 address: %w", err)
		}
		host.Host = ip4.String()
	case hostnameAddrType:
		n, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("decode hostname length: %w", err)
		}

		hostname := make([]byte, n)
		if _, err := io.ReadFull(r, hostname); err != nil {
			return nil, fmt.Errorf("decode hostname: %w", err)
		}
		host.Host = string(hostname)
	default:
		return nil, fmt.Errorf("%w (%v)", ErrInvalidAddrType, hexByte(addrType))
	}

	host.Port, err = readPort(r)
	if err != nil {
		return nil, fmt.Errorf("decode port: %w", err)
	}

	return &host, nil
}

func readPort(r *bufio.Reader) (uint16, error) {
	var port [2]byte
	if _, err := io.ReadFull(r, port[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(port[:]), nil
}
