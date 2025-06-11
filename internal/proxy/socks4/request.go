package socks4

import (
	"bufio"
	"fmt"
	"io"

	"github.com/cerfical/socks2http/internal/proxy/addr"
)

func ReadRequest(r *bufio.Reader) (*Request, error) {
	if err := checkVersion(r, VersionCode); err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}

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

	dstHost, err := readHostname(r, dstIP)
	if err != nil {
		return nil, fmt.Errorf("decode destination hostname: %w", err)
	}

	return &Request{
		Command(command),
		*addr.NewAddr(dstHost, dstPort),
		username,
	}, nil
}

type Request struct {
	Command  Command
	DstAddr  addr.Addr
	Username string
}

func (r *Request) Write(w io.Writer) error {
	bytes := []byte{VersionCode, byte(r.Command)}

	dstAddr, dstHostname, err := encodeAddr(&r.DstAddr)
	if err != nil {
		return fmt.Errorf("encode destination address: %w", err)
	}
	bytes = append(bytes, dstAddr...)

	// Append username
	bytes = append(bytes, []byte(r.Username)...)
	bytes = append(bytes, 0)

	// Append destination hostname, if any
	if len(dstHostname) > 0 {
		bytes = append(bytes, []byte(dstHostname)...)
		bytes = append(bytes, 0)
	}

	_, err = w.Write(bytes)
	return err
}
