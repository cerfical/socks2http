package socks5

import (
	"bufio"
	"fmt"
	"io"

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
	// Encode request header
	bytes := []byte{VersionCode, byte(r.Command), 0x00}

	addr, err := encodeAddr(&r.DstAddr)
	if err != nil {
		return fmt.Errorf("encode destination address: %w ", err)
	}
	bytes = append(bytes, addr...)

	_, err = w.Write(bytes)
	return err
}
