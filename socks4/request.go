package socks4

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"slices"

	"github.com/cerfical/socks2http/addr"
)

func ReadRequest(r *bufio.Reader) (*Request, error) {
	if err := checkVersion(r, VersionCode); err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}
	r.Discard(1)

	command, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode command: %w", err)
	}

	dstIP, dstPort, err := readAddr(r)
	if err != nil {
		return nil, fmt.Errorf("decode destination address: %w", err)
	}

	username, err := r.ReadString('\x00')
	if err != nil {
		return nil, fmt.Errorf("decode username: %w", err)
	}
	username = username[:len(username)-1] // Remove the null terminator

	dstAddr := ""
	if slices.Equal(dstIP[:3], []byte{0, 0, 0}) && dstIP[3] != 0 {
		// The destination address is a hostname
		dstAddr, err = r.ReadString('\x00')
		if err != nil {
			return nil, fmt.Errorf("decode destination hostname: %w", err)
		}
		dstAddr = dstAddr[:len(dstAddr)-1] // Remove the null terminator
	} else {
		dstAddr = dstIP.String()
	}

	return &Request{
		Command(command),
		*addr.NewHost(dstAddr, dstPort),
		username,
	}, nil
}

type Request struct {
	Command  Command
	DstAddr  addr.Host
	Username string
}

func (r *Request) Write(w io.Writer) error {
	bytes := []byte{VersionCode, byte(r.Command)}
	var (
		dstIP   addr.IPv4
		dstAddr string
	)

	// Append destination IPv4 address and port
	bytes = binary.BigEndian.AppendUint16(bytes, r.DstAddr.Port)
	if ip4, ok := r.DstAddr.ToIPv4(); ok {
		dstIP = ip4
	} else {
		dstIP[3] = 1
		dstAddr = r.DstAddr.Hostname
	}
	bytes = append(bytes, dstIP[:]...)

	// Append the username
	bytes = append(bytes, []byte(r.Username)...)
	bytes = append(bytes, 0)

	// Append the destination hostname, if any
	if len(dstAddr) > 0 {
		bytes = append(bytes, []byte(dstAddr)...)
		bytes = append(bytes, 0)
	}

	_, err := w.Write(bytes)
	return err
}
