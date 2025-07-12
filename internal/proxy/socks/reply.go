package socks

import (
	"bufio"
	"fmt"
	"io"

	"github.com/cerfical/socks2http/internal/proxy/addr"
)

func ReadReply(r *bufio.Reader) (*Reply, error) {
	ver, err := checkVersion(r, 0, V5)
	if err != nil {
		return nil, err
	}

	statusByte, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("decode status: %w", err)
	}

	rep := Reply{
		Version: ver,
		Status:  decodeStatus(statusByte),
	}

	switch rep.Version {
	case 0:
		rep.Version = V4

		bindIP, bindPort, err := v4ReadAddr(r)
		if err != nil {
			return nil, fmt.Errorf("decode bind address: %w", err)
		}

		bindHost, err := v4ReadHostname(r, bindIP)
		if err != nil {
			return nil, fmt.Errorf("decode bind hostname: %w", err)
		}
		rep.BindAddr = *addr.NewAddr(bindHost, bindPort)
	case V5:
		if err := v5CheckReserved(r); err != nil {
			return nil, err
		}

		bindAddr, err := v5ReadAddr(r)
		if err != nil {
			return nil, fmt.Errorf("decode bind address: %w", err)
		}
		rep.BindAddr = *bindAddr
	}

	return &rep, nil
}

type Reply struct {
	Version  Version
	Status   Status
	BindAddr addr.Addr
}

func (r *Reply) Write(w io.Writer) error {
	bytes := []byte{byte(r.Version), encodeStatus(r.Status, r.Version)}

	switch r.Version {
	case V4:
		// SOCKS4 reply version is 0
		bytes[0] = 0

		bindAddr, bindHostname, err := v4EncodeAddr(&r.BindAddr)
		if err != nil {
			return fmt.Errorf("encode bind address: %w", err)
		}
		bytes = append(bytes, bindAddr...)

		// Append bind hostname, if any
		if len(bindHostname) > 0 {
			bytes = append(bytes, []byte(bindHostname)...)
			bytes = append(bytes, 0)
		}
	case V5:
		// Reserved field
		bytes = append(bytes, 0)

		addr, err := v5EncodeAddr(&r.BindAddr)
		if err != nil {
			return fmt.Errorf("encode bind address: %w ", err)
		}
		bytes = append(bytes, addr...)
	default:
		return badVersion(r.Version)
	}

	_, err := w.Write(bytes)
	return err
}
