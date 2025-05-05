package socks5

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"

	"github.com/cerfical/socks2http/proxy/addr"
)

const VersionCode = 0x05
const (
	ip4AddrType      = 0x01
	hostnameAddrType = 0x03
	ip6AddrType      = 0x04
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

	var addr addr.Addr
	switch addrType {
	case ip4AddrType:
		var ip4 [4]byte
		if _, err := io.ReadFull(r, ip4[:]); err != nil {
			return nil, fmt.Errorf("decode IPv4 address: %w", err)
		}
		addr.Host = net.IP(ip4[:]).String()
	case ip6AddrType:
		var ip6 [16]byte
		if _, err := io.ReadFull(r, ip6[:]); err != nil {
			return nil, fmt.Errorf("decode IPv6 address: %w", err)
		}
		addr.Host = net.IP(ip6[:]).String()
	case hostnameAddrType:
		n, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("decode hostname length: %w", err)
		}

		hostname := make([]byte, n)
		if _, err := io.ReadFull(r, hostname); err != nil {
			return nil, fmt.Errorf("decode hostname: %w", err)
		}
		addr.Host = string(hostname)
	default:
		return nil, fmt.Errorf("%w (%v)", ErrInvalidAddrType, hexByte(addrType))
	}

	addr.Port, err = readPort(r)
	if err != nil {
		return nil, fmt.Errorf("decode port: %w", err)
	}

	return &addr, nil
}

func readPort(r *bufio.Reader) (uint16, error) {
	var port [2]byte
	if _, err := io.ReadFull(r, port[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(port[:]), nil
}

func encodeAddr(a *addr.Addr) ([]byte, error) {
	var bytes []byte
	if ip := net.ParseIP(a.Host); ip != nil {
		var addrType byte
		if ip4 := ip.To4(); ip4 != nil {
			// IPv4 address
			addrType = ip4AddrType
			ip = ip4
		} else {
			// IPv6 address
			addrType = ip6AddrType
		}
		bytes = append(bytes, addrType)
		bytes = append(bytes, ip[:]...)
	} else {
		// Hostname address
		n := len(a.Host)
		if n > math.MaxUint8 {
			return nil, fmt.Errorf("%w (%v)", ErrHostnameTooLong, n)
		}
		bytes = append(bytes, hostnameAddrType, byte(n))
		bytes = append(bytes, []byte(a.Host)...)
	}
	bytes = binary.BigEndian.AppendUint16(bytes, a.Port)
	return bytes, nil
}
