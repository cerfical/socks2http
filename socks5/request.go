package socks5

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"

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
	DstAddr addr.Host
}

func (r *Request) Write(w io.Writer) error {
	// Encode the request header
	bytes := []byte{VersionCode, byte(r.Command), 0x00}

	if ip4, ok := r.DstAddr.ToIPv4(); ok {
		// Append destination IPv4 address
		bytes = append(bytes, ip4AddrType)
		bytes = append(bytes, ip4[:]...)
	} else {
		n := len(r.DstAddr.Hostname)
		if n > math.MaxUint8 {
			return fmt.Errorf("encode destination address: %w (%v)", ErrHostnameTooLong, n)
		}

		// Append destination hostname
		bytes = append(bytes, hostnameAddrType)
		bytes = append(bytes, byte(n))
		bytes = append(bytes, []byte(r.DstAddr.Hostname)...)
	}

	// Append destination port
	bytes = binary.BigEndian.AppendUint16(bytes, r.DstAddr.Port)

	_, err := w.Write(bytes)
	return err
}
