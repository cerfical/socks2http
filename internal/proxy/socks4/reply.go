package socks4

import (
	"bufio"
	"fmt"
	"io"

	"github.com/cerfical/socks2http/internal/proxy/addr"
)

const replyVersion = 0

func ReadReply(r *bufio.Reader) (*Reply, error) {
	if err := checkVersion(r, replyVersion); err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}

	status, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode status: %w", err)
	}

	bindIP, bindPort, err := readAddr(r)
	if err != nil {
		return nil, fmt.Errorf("decode bind address: %w", err)
	}

	bindHost, err := readHostname(r, bindIP)
	if err != nil {
		return nil, fmt.Errorf("decode bind hostname: %w", err)
	}

	return &Reply{
		Status(status),
		*addr.NewAddr(bindHost, bindPort),
	}, nil
}

type Reply struct {
	Status   Status
	BindAddr addr.Addr
}

func (r *Reply) Write(w io.Writer) error {
	bytes := []byte{replyVersion, byte(r.Status)}

	bindAddr, bindHostname, err := encodeAddr(&r.BindAddr)
	if err != nil {
		return fmt.Errorf("encode bind address: %w", err)
	}
	bytes = append(bytes, bindAddr...)

	// Append bind hostname, if any
	if len(bindHostname) > 0 {
		bytes = append(bytes, []byte(bindHostname)...)
		bytes = append(bytes, 0)
	}

	_, err = w.Write(bytes)
	return err
}
