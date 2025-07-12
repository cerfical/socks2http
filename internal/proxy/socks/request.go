package socks

import (
	"bufio"
	"fmt"
	"io"

	"github.com/cerfical/socks2http/internal/proxy/addr"
)

func ReadRequest(r *bufio.Reader) (*Request, error) {
	ver, err := checkVersion(r, V4, V5)
	if err != nil {
		return nil, err
	}

	commandByte, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode command: %w", err)
	}

	req := Request{
		Version: ver,
		Command: Command(commandByte),
	}

	switch req.Version {
	case V4:
		dstIP, dstPort, err := v4ReadAddr(r)
		if err != nil {
			return nil, fmt.Errorf("decode destination address: %w", err)
		}

		username, err := r.ReadString('\x00')
		if err != nil {
			return nil, fmt.Errorf("decode username: %w", err)
		}
		req.Username = username[:len(username)-1] // Remove the null terminator

		dstHost, err := v4ReadHostname(r, dstIP)
		if err != nil {
			return nil, fmt.Errorf("decode destination hostname: %w", err)
		}
		req.DstAddr = *addr.NewAddr(dstHost, dstPort)
	case V5:
		if err := v5CheckReserved(r); err != nil {
			return nil, err
		}

		dstAddr, err := v5ReadAddr(r)
		if err != nil {
			return nil, fmt.Errorf("decode destination address: %w", err)
		}
		req.DstAddr = *dstAddr
	}

	return &req, nil
}

type Request struct {
	Version  Version
	Command  Command
	DstAddr  addr.Addr
	Username string
}

func (r *Request) Write(w io.Writer) error {
	bytes := []byte{byte(r.Version), byte(r.Command)}

	switch r.Version {
	case V4:
		dstAddr, dstHostname, err := v4EncodeAddr(&r.DstAddr)
		if err != nil {
			return fmt.Errorf("encode destination address: %w", err)
		}
		bytes = append(bytes, dstAddr...)

		bytes = append(bytes, []byte(r.Username)...)
		bytes = append(bytes, 0)

		if len(dstHostname) > 0 {
			bytes = append(bytes, []byte(dstHostname)...)
			bytes = append(bytes, 0)
		}
	case V5:
		// Append reserved field
		bytes = append(bytes, 0)

		addr, err := v5EncodeAddr(&r.DstAddr)
		if err != nil {
			return fmt.Errorf("encode destination address: %w ", err)
		}
		bytes = append(bytes, addr...)
	default:
		return badVersion(r.Version)
	}

	_, err := w.Write(bytes)
	return err
}
