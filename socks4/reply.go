package socks4

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"

	"github.com/cerfical/socks2http/addr"
)

const replyVersion = 0

func NewReply(c ReplyCode, bindAddr *addr.Host) *Reply {
	r := Reply{Code: c}
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

	code, ok := makeReplyCode(h.Code)
	if !ok {
		return nil, fmt.Errorf("invalid reply code (%v)", hexByte(h.Code))
	}

	// Check if an empty bind address was specified
	if h.BindIP == [4]byte{0, 0, 0, 0} && h.BindPort == 0 {
		return NewReply(code, nil), nil
	}

	bindAddr := addr.NewHost(addr.IPv4(h.BindIP).String(), h.BindPort)
	return NewReply(code, bindAddr), nil
}

type Reply struct {
	Code     ReplyCode
	BindAddr addr.Host
}

func (r *Reply) Write(w io.Writer) error {
	h := replyHeader{
		Version:  replyVersion,
		Code:     byte(r.Code),
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
	Code     byte
	BindPort uint16
	BindIP   [4]byte
}
