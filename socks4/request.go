package socks4

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
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
	if slices.Equal(dstIP[:3], []byte{0, 0, 0}) {
		if dstIP[3] != 0 {
			// The destination address is a hostname
			dstAddr, err = r.ReadString('\x00')
			if err != nil {
				return nil, fmt.Errorf("decode destination hostname: %w", err)
			}
			dstAddr = dstAddr[:len(dstAddr)-1] // Remove the null terminator
		} else {
			dstAddr = ""
		}
	} else {
		dstAddr = dstIP.String()
	}

	return &Request{
		Command(command),
		*addr.New(dstAddr, dstPort),
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
	var (
		dstIP   [4]byte
		dstAddr string
	)

	// Append destination IPv4 address and port
	bytes = binary.BigEndian.AppendUint16(bytes, r.DstAddr.Port)

	if r.DstAddr.Host != "" {
		if ip := net.ParseIP(r.DstAddr.Host); ip != nil {
			ip4 := ip.To4()
			if ip4 == nil {
				// Destination is an IPv6 address
				return fmt.Errorf("encode destination address: not an IPv4 address: %v", ip)
			}
			// Destination is an IPv4 address
			dstIP = [4]byte(ip4)
		} else {
			// Destination is a hostname
			dstAddr = r.DstAddr.Host
			dstIP = [4]byte{0, 0, 0, 1}
		}
	} else {
		// No destination address was specified
		dstIP = [4]byte{0, 0, 0, 0}
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
