package socks

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"net"
	"slices"

	"github.com/cerfical/socks2http/internal/proxy/addr"
)

const (
	ip4AddrType      = 0x01
	hostnameAddrType = 0x03
	ip6AddrType      = 0x04
)

var ErrInvalidVersion = fmt.Errorf("invalid version")

func v5ReadAddr(r *bufio.Reader) (*addr.Addr, error) {
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
		return nil, fmt.Errorf("invalid address type (%#02x)", addrType)
	}

	addr.Port, err = readPort(r)
	if err != nil {
		return nil, fmt.Errorf("decode port: %w", err)
	}

	return &addr, nil
}

func v5EncodeAddr(a *addr.Addr) ([]byte, error) {
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
			return nil, fmt.Errorf("hostname too long (%v)", n)
		}
		bytes = append(bytes, hostnameAddrType, byte(n))
		bytes = append(bytes, []byte(a.Host)...)
	}
	bytes = binary.BigEndian.AppendUint16(bytes, a.Port)
	return bytes, nil
}

func v5CheckReserved(r *bufio.Reader) error {
	rsv, err := r.ReadByte()
	if err != nil {
		return fmt.Errorf("decode reserved field: %w", err)
	}
	if rsv != 0 {
		return fmt.Errorf("non-zero value for reserved field (%#02x)", rsv)
	}
	return nil
}

func v4ReadHostname(r *bufio.Reader, ip net.IP) (string, error) {
	if bytes.HasPrefix(ip[:3], []byte{0, 0, 0}) {
		if ip[3] == 0 {
			// The address is empty
			return "", nil
		}

		// The address is a hostname
		addr, err := r.ReadString('\x00')
		if err != nil {
			return "", err
		}
		return addr[:len(addr)-1], nil // Remove the null terminator
	}

	// The address is an IPv4 address
	return ip.String(), nil
}

func v4ReadAddr(r *bufio.Reader) (net.IP, uint16, error) {
	port, err := readPort(r)
	if err != nil {
		return nil, 0, fmt.Errorf("decode port: %w", err)
	}

	var ip4 [4]byte
	if _, err := io.ReadFull(r, ip4[:]); err != nil {
		return nil, 0, fmt.Errorf("decode IPv4 address: %w", err)
	}
	return net.IP(ip4[:]), port, nil
}

func v4EncodeAddr(a *addr.Addr) (addr []byte, hostname []byte, err error) {
	port := binary.BigEndian.AppendUint16(nil, a.Port)
	if a.Host == "" {
		// Address not specified
		return append(port, 0, 0, 0, 0), nil, nil
	}

	ip := net.ParseIP(a.Host)
	if ip == nil {
		// Hostname
		return append(port, 0, 0, 0, 1), []byte(a.Host), nil
	}

	ip4 := ip.To4()
	if ip4 == nil {
		// IPv6 address
		return nil, nil, fmt.Errorf("not an IPv4 address: %v", ip)
	}
	// IPv4 address
	return append(port, ip4...), nil, nil

}

func readPort(r *bufio.Reader) (uint16, error) {
	var port [2]byte
	if _, err := io.ReadFull(r, port[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(port[:]), nil
}

func checkVersion(r *bufio.Reader, vs ...Version) (Version, error) {
	versionBytes, err := r.Peek(1)
	if err != nil {
		return 0, fmt.Errorf("decode version: %w", err)
	}

	ver := Version(versionBytes[0])
	if !slices.Contains(vs, ver) {
		return 0, fmt.Errorf("%w (%v)", ErrInvalidVersion, ver)
	}
	_, _ = r.Discard(1)

	return ver, nil
}

func badVersion(v Version) error {
	return fmt.Errorf("%w (%v)", ErrInvalidVersion, v)
}
