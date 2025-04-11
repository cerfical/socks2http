package socks

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/cerfical/socks2http/addr"
)

func NewRequest(v Version, c Command, dstAddr *addr.Host) *Request {
	r := Request{
		Version: v,
		Command: c,
	}
	if dstAddr != nil {
		r.DstAddr = *dstAddr
	}
	return &r
}

func ReadRequest(r *bufio.Reader) (*Request, error) {
	version, err := r.Peek(1)
	if err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}

	v, ok := makeVersion(version[0])
	if !ok {
		return nil, fmt.Errorf("%w %v", ErrInvalidVersion, v)
	}

	var h requestHeader
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}

	c, ok := makeCommand(h.Command)
	if !ok {
		return nil, fmt.Errorf("%w %v", ErrInvalidCommand, c)
	}

	dstAddr := addr.NewHost(addr.IPv4(h.DstIP).String(), h.DstPort)
	req := Request{
		Version: v,
		Command: c,
		DstAddr: *dstAddr,
	}
	return &req, nil
}

type Request struct {
	Version Version
	Command Command
	DstAddr addr.Host
}

func (r *Request) Write(w io.Writer) error {
	ip4, ok := r.DstAddr.ToIPv4()
	if !ok {
		return fmt.Errorf("not an IPv4 address: %v", r.DstAddr.Hostname)
	}

	h := requestHeader{
		Version: byte(r.Version),
		Command: byte(r.Command),
		DstPort: r.DstAddr.Port,
		DstIP:   ip4,
	}

	// +1 is the NULL character
	bytes := make([]byte, unsafe.Sizeof(h)+1)
	if _, err := binary.Encode(bytes, binary.BigEndian, &h); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}

	_, err := w.Write(bytes)
	return err
}

type requestHeader struct {
	Version byte
	Command byte
	DstPort uint16
	DstIP   [4]byte
}
