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

	dstIP4, dstPort, err := readAddr(r)
	if err != nil {
		return nil, fmt.Errorf("decode destination address: %w", err)
	}

	username, err := r.ReadString('\x00')
	if err != nil {
		return nil, fmt.Errorf("decode username: %w", err)
	}
	username = username[:len(username)-1] // Remove the null terminator

	hostname := ""
	if slices.Equal(dstIP4[:3], []byte{0, 0, 0}) {
		// The destination address is a hostname
		hostname, err = r.ReadString('\x00')
		if err != nil {
			return nil, fmt.Errorf("decode destination hostname: %w", err)
		}
		hostname = hostname[:len(hostname)-1] // Remove the null terminator
	} else {
		hostname = dstIP4.String()
	}

	return &Request{
		Command(command),
		*addr.NewHost(hostname, dstPort),
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
		dstIP       addr.IPv4
		dstHostname string
	)

	// Append destination IPv4 address and port
	bytes = binary.BigEndian.AppendUint16(bytes, r.DstAddr.Port)
	if ip4, ok := r.DstAddr.ToIPv4(); ok {
		dstIP = ip4
	} else {
		dstHostname = r.DstAddr.Hostname
	}
	bytes = append(bytes, dstIP[:]...)

	// Append the username
	bytes = append(bytes, []byte(r.Username)...)
	bytes = append(bytes, 0)

	// Append the destination hostname, if any
	if len(dstHostname) > 0 {
		bytes = append(bytes, []byte(dstHostname)...)
		bytes = append(bytes, 0)
	}

	_, err := w.Write(bytes)
	return err
}
