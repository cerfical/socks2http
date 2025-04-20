package socks

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/cerfical/socks2http/addr"
)

const replyVersion = 0

func NewReply(s Status, bindAddr *addr.Host) *Reply {
	r := Reply{Status: s}
	if bindAddr != nil {
		r.BindAddr = *bindAddr
	}
	return &r
}

func ReadReply(r *bufio.Reader) (*Reply, error) {
	version, err := r.Peek(1)
	if err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}
	if v := version[0]; v != replyVersion {
		return nil, fmt.Errorf("invalid version code (%v)", hexByte(v))
	}

	var h replyHeader
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}

	status, ok := makeStatus(h.Status)
	if !ok {
		return nil, fmt.Errorf("invalid reply code (%v)", hexByte(h.Status))
	}

	// Check if an empty bind address was specified
	if h.BindIP == [4]byte{0, 0, 0, 0} && h.BindPort == 0 {
		return NewReply(status, nil), nil
	}

	bindAddr := addr.NewHost(addr.IPv4(h.BindIP).String(), h.BindPort)
	return NewReply(status, bindAddr), nil
}

type Reply struct {
	Status   Status
	BindAddr addr.Host
}

func (r *Reply) Write(w io.Writer) error {
	h := replyHeader{
		Version:  replyVersion,
		Status:   byte(r.Status),
		BindPort: r.BindAddr.Port,
	}

	// Check if a non-empty bind address was specified
	if r.BindAddr.Hostname != "" {
		ip, ok := r.BindAddr.ToIPv4()
		if !ok {
			return fmt.Errorf("not an IPv4 address: %v", r.BindAddr.Hostname)
		}
		h.BindIP = ip
	}

	if err := binary.Write(w, binary.BigEndian, &h); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}
	return nil
}

type replyHeader struct {
	Version  byte
	Status   byte
	BindPort uint16
	BindIP   [4]byte
}
