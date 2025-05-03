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

func ReadRequest(r *bufio.Reader) (*Request, error) {
	if err := checkVersion(r); err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}

	cmd, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode command: %w", err)
	}

	if err := checkReserved(r); err != nil {
		return nil, fmt.Errorf("decode reserved field: %w", err)
	}

	dstAddr, err := readAddr(r)
	if err != nil {
		return nil, fmt.Errorf("decode destination address: %w", err)
	}

	return &Request{Command(cmd), *dstAddr}, nil
}

type Request struct {
	Command Command
	DstAddr addr.Addr
}

func (r *Request) Write(w io.Writer) error {
	// Encode the request header
	bytes := []byte{VersionCode, byte(r.Command), 0x00}

	if ip := net.ParseIP(r.DstAddr.Host); ip != nil {
		var addrType byte
		if ip4 := ip.To4(); ip4 != nil {
			// Destination is an IPv4 address
			addrType = ip4AddrType
			ip = ip4
		} else {
			// Destination is an IPv6 address
		}
		bytes = append(bytes, addrType)
		bytes = append(bytes, ip[:]...)
	} else {
		// Destination is a hostname
		n := len(r.DstAddr.Host)
		if n > math.MaxUint8 {
			return fmt.Errorf("encode destination address: %w (%v)", ErrHostnameTooLong, n)
		}
		bytes = append(bytes, hostnameAddrType)
		bytes = append(bytes, byte(n))
		bytes = append(bytes, []byte(r.DstAddr.Host)...)
	}

	// Append destination port
	bytes = binary.BigEndian.AppendUint16(bytes, r.DstAddr.Port)

	_, err := w.Write(bytes)
	return err
}
