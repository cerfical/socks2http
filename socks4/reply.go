package socks4

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"slices"

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

	bindIP, bindPort, err := readAddr(r)
	if err != nil {
		return nil, fmt.Errorf("decode bind address: %w", err)
	}

	bindAddr := ""
	if slices.Equal(bindIP[:3], []byte{0, 0, 0}) && bindIP[3] != 0 {
		// The bind address is a hostname
		bindAddr, err = r.ReadString('\x00')
		if err != nil {
			return nil, fmt.Errorf("decode bind hostname: %w", err)
		}
		bindAddr = bindAddr[:len(bindAddr)-1] // Remove the null terminator
	} else {
		bindAddr = bindIP.String()
	}

	return &Reply{
		Status(status),
		*addr.NewHost(bindAddr, bindPort),
	}, nil
}

type Reply struct {
	Status   Status
	BindAddr addr.Host
}

func (r *Reply) Write(w io.Writer) error {
	bytes := []byte{replyVersion, byte(r.Status)}
	var (
		bindIP   addr.IPv4
		bindAddr string
	)

	// Append bind IPv4 address and port
	bytes = binary.BigEndian.AppendUint16(bytes, r.BindAddr.Port)
	if ip4, ok := r.BindAddr.ToIPv4(); ok {
		bindIP = ip4
	} else {
		bindIP[3] = 1
		bindAddr = r.BindAddr.Hostname
	}
	bytes = append(bytes, bindIP[:]...)

	// Append the bind hostname, if any
	if len(bindAddr) > 0 {
		bytes = append(bytes, []byte(bindAddr)...)
		bytes = append(bytes, 0)
	}

	_, err := w.Write(bytes)
	return err
}
