package socks

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

const replyVersion = 0
const (
	Granted    Reply = 0x5a
	Rejected   Reply = 0x5b
	NoAuth     Reply = 0x5c
	AuthFailed Reply = 0x5d
)

var replies = map[Reply]string{
	Granted:    "request granted",
	Rejected:   "request rejected or failed",
	NoAuth:     "request rejected because SOCKS server cannot connect to identd on the client",
	AuthFailed: "request rejected because the client program and identd report different user-ids",
}

func ReadReply(r *bufio.Reader) (Reply, error) {
	version, err := r.Peek(1)
	if err != nil {
		return 0, fmt.Errorf("decode version: %w", err)
	}
	if v := version[0]; v != replyVersion {
		return 0, fmt.Errorf("invalid version %v", Version(v))
	}

	var h replyHeader
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return 0, fmt.Errorf("decode header: %w", err)
	}

	reply, ok := makeReply(h.Code)
	if !ok {
		return 0, fmt.Errorf("%w %v", ErrInvalidReply, reply)
	}
	return reply, nil
}

func makeReply(b byte) (r Reply, isValid bool) {
	r = Reply(b)
	if _, ok := replies[r]; ok {
		return r, true
	}
	return r, false
}

type Reply byte

func (r Reply) String() string {
	code := fmt.Sprintf("(%#02x)", byte(r))
	if s, ok := replies[r]; ok {
		return fmt.Sprintf("%v %v", s, code)
	}
	return code
}

func (r Reply) Write(w io.Writer) error {
	h := replyHeader{
		Version: replyVersion,
		Code:    byte(r),
	}
	if err := binary.Write(w, binary.BigEndian, &h); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}
	return nil
}

type replyHeader struct {
	Version byte
	Code    byte
	_       uint16
	_       [4]byte
}
