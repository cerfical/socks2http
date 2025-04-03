package socks

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/cerfical/socks2http/addr"
)

func NewRequest(v Version, c Command, h *addr.Host) *Request {
	return &Request{
		Version: v,
		Command: c,
		Host:    *h,
	}
}

func ReadRequest(r *bufio.Reader) (*Request, error) {
	version, err := r.Peek(1)
	if err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}

	v, ok := makeVersion(version[0])
	if !ok {
		return nil, fmt.Errorf("%w %v", ErrUnsupportedVersion, v)
	}

	var h requestHeader
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}

	c, ok := makeCommand(h.Command)
	if !ok {
		return nil, fmt.Errorf("%w %v", ErrUnsupportedCommand, c)
	}

	host := addr.NewHost(addr.IPv4(h.DstIP).String(), h.DstPort)
	req := Request{
		Version: v,
		Command: c,
		Host:    *host,
	}
	return &req, nil
}

type Request struct {
	Version Version
	Command Command
	Host    addr.Host
}

func (r *Request) Write(w io.Writer) error {
	ipv4, err := addr.LookupIPv4(r.Host.Hostname)
	if err != nil {
		return fmt.Errorf("resolve host %v: %w", &r.Host, err)
	}

	h := requestHeader{
		Version: byte(r.Version),
		Command: byte(r.Command),
		DstPort: r.Host.Port,
		DstIP:   ipv4,
	}

	// +1 is the NULL character
	bytes := make([]byte, unsafe.Sizeof(h)+1)
	if _, err := binary.Encode(bytes, binary.BigEndian, &h); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}

	_, err = w.Write(bytes)
	return err
}

type requestHeader struct {
	Version byte
	Command byte
	DstPort uint16
	DstIP   [4]byte
}
