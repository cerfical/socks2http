package socks4

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/cerfical/socks2http/addr"
)

const replyVersion = 0

func ReadReply(r *bufio.Reader) (*Reply, error) {
	if err := checkVersion(r, replyVersion); err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}
	r.Discard(1)

	status, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode status: %w", err)
	}

	bindIP4, bindPort, err := readAddr(r)
	if err != nil {
		return nil, fmt.Errorf("decode bind address: %w", err)
	}

	return &Reply{
		Status(status),
		*addr.NewHost(bindIP4.String(), bindPort),
	}, nil
}

type Reply struct {
	Status   Status
	BindAddr addr.Host
}

func (r *Reply) Write(w io.Writer) error {
	bytes := []byte{replyVersion, byte(r.Status)}

	// Append bind IPv4 address and port
	bytes = binary.BigEndian.AppendUint16(bytes, r.BindAddr.Port)

	var bindIP addr.IPv4
	if r.BindAddr.Hostname != "" {
		ip, ok := r.BindAddr.ToIPv4()
		if !ok {
			return fmt.Errorf("not an IPv4 address: %v", r.BindAddr.Hostname)
		}
		bindIP = ip
	}
	bytes = append(bytes, bindIP[:]...)

	_, err := w.Write(bytes)
	return err
}
