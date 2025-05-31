package socks4

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"slices"

	"github.com/cerfical/socks2http/internal/proxy/addr"
)

const VersionCode = 0x04

type hexByte byte

func (b hexByte) String() string {
	return fmt.Sprintf("%#02x", byte(b))
}

func checkVersion(r *bufio.Reader, version byte) error {
	versionBytes, err := r.Peek(1)
	if err != nil {
		return err
	}
	if v := versionBytes[0]; v != version {
		return fmt.Errorf("%w (%v)", ErrInvalidVersion, hexByte(v))
	}
	return nil
}

func readAddr(r *bufio.Reader) (net.IP, uint16, error) {
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

func readHostname(r *bufio.Reader, ip net.IP) (string, error) {
	//	bytes.HasPrefix()
	if slices.Equal(ip[:3], []byte{0, 0, 0}) {
		if ip[3] == 0 {
			// The address is empty
			return "", nil
		}

		// The address is a hostname
		addr, err := r.ReadString(0)
		if err != nil {
			return "", err
		}
		return addr[:len(addr)-1], nil // Remove the null terminator
	}

	// The address is an IPv4 address
	return ip.String(), nil
}

func readPort(r *bufio.Reader) (uint16, error) {
	var port [2]byte
	if _, err := io.ReadFull(r, port[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(port[:]), nil
}

func encodeAddr(a *addr.Addr) (addr []byte, hostname []byte, err error) {
	port := binary.BigEndian.AppendUint16(nil, a.Port)
	if a.Host != "" {
		if ip := net.ParseIP(a.Host); ip != nil {
			ip4 := ip.To4()
			if ip4 == nil {
				// IPv6 address
				return nil, nil, fmt.Errorf("not an IPv4 address: %v", ip)
			}
			// IPv4 address
			return append(port, ip4...), nil, nil
		}
		// Hostname
		return append(port, 0, 0, 0, 1), []byte(a.Host), nil
	}
	// No address was specified
	return append(port, 0, 0, 0, 0), nil, nil
}
