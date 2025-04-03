package socks

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"

	"github.com/cerfical/socks2http/addr"
)

const (
	V4 = 0x04
)

const (
	Connect = 0x01
)

var ErrUnsupportedVersion = errors.New("unsupported version")
var ErrUnsupportedCommand = errors.New("unsupported command")

func NewRequest(version, cmd byte, h *addr.Host) *Request {
	return &Request{
		Version: version,
		Command: cmd,
		Host:    *h,
	}
}

func ReadRequest(r *bufio.Reader) (*Request, error) {
	version, err := r.Peek(1)
	if err != nil {
		return nil, fmt.Errorf("decode version: %w", err)
	}

	if v := version[0]; v != V4 {
		return nil, fmt.Errorf("%w %v", ErrUnsupportedVersion, v)
	}

	var h header
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}

	if c := h.Command; c != Connect {
		return nil, fmt.Errorf("%w %v", ErrUnsupportedCommand, c)
	}

	req := Request{
		Version: h.Version,
		Command: h.Command,
		Host:    *addr.NewHost(h.DstIP.String(), h.DstPort),
	}
	return &req, nil
}

type Request struct {
	Version byte
	Command byte
	Host    addr.Host
}

type header struct {
	Version byte
	Command byte
	DstPort uint16
	DstIP   addr.IPv4
}

func (r *Request) Write(w io.Writer) error {
	ipv4, err := addr.LookupIPv4(r.Host.Hostname)
	if err != nil {
		return fmt.Errorf("resolve host %v: %w", &r.Host, err)
	}

	h := header{
		Version: r.Version,
		Command: r.Command,
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
