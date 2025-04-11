package socks

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const replyVersion = 0

func NewReply(s Status) *Reply {
	return &Reply{
		Status: s,
	}
}

func ReadReply(r *bufio.Reader) (*Reply, error) {
	version, err := r.Peek(1)
	if err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}
	if v := version[0]; v != replyVersion {
		return nil, fmt.Errorf("invalid reply version %v", printByte(v))
	}

	var h replyHeader
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}

	status, ok := makeStatus(h.Status)
	if !ok {
		return nil, fmt.Errorf("%w %v", ErrInvalidReply, status)
	}
	return NewReply(status), nil
}

type Reply struct {
	Status Status
}

func (r *Reply) Write(w io.Writer) error {
	h := replyHeader{
		Version: replyVersion,
		Status:  byte(r.Status),
	}
	if err := binary.Write(w, binary.BigEndian, &h); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}
	return nil
}

type replyHeader struct {
	Version byte
	Status  byte
	_       uint16
	_       [4]byte
}
