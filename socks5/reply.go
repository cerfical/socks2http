package socks5

import (
	"bufio"
	"fmt"
	"io"

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
	// Encode reply header
	bytes := []byte{VersionCode, byte(r.Status), 0x00}

	addr, err := encodeAddr(&r.BindAddr)
	if err != nil {
		return fmt.Errorf("encode bind address: %w ", err)
	}
	bytes = append(bytes, addr...)

	_, err = w.Write(bytes)
	return err
}
