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

	v, ok := decodeVersion(version[0])
	if !ok {
		return nil, fmt.Errorf("invalid version code (%v)", hexByte(version[0]))
	}

	var h requestHeader
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}

	c, ok := decodeCommand(h.Command)
	if !ok {
		return nil, fmt.Errorf("invalid command code (%v)", hexByte(h.Command))
	}

	if h.DstIP.IsEmpty() {
		return nil, fmt.Errorf("empty destination address")
	}

	un, err := readNullString(r)
	if err != nil {
		return nil, fmt.Errorf("read username: %w", err)
	}

	dstAddr := addr.NewHost(h.DstIP.String(), h.DstPort)
	req := Request{
		Version:  v,
		Command:  c,
		DstAddr:  *dstAddr,
		Username: un,
	}
	return &req, nil
}

type Request struct {
	Version  Version
	Command  Command
	DstAddr  addr.Host
	Username string
}

func (r *Request) Write(w io.Writer) error {
	ip4, ok := r.DstAddr.ToIPv4()
	if !ok {
		return fmt.Errorf("not an IPv4 address: %v", r.DstAddr.Hostname)
	}

	v, ok := encodeVersion(r.Version)
	if !ok {
		return fmt.Errorf("invalid version")
	}

	c, ok := encodeCommand(r.Command)
	if !ok {
		return fmt.Errorf("invalid command")
	}

	h := requestHeader{
		Version: v,
		Command: c,
		DstPort: r.DstAddr.Port,
		DstIP:   ip4,
	}

	bytes := make([]byte, unsafe.Sizeof(h))
	if _, err := binary.Encode(bytes, binary.BigEndian, &h); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}

	// Append the username bytes
	bytes = append(bytes, []byte(r.Username)...)
	bytes = append(bytes, 0)

	_, err := w.Write(bytes)
	return err
}

type requestHeader struct {
	Version byte
	Command byte
	DstPort uint16
	DstIP   addr.IPv4
}

func readNullString(r *bufio.Reader) (string, error) {
	var buf []byte
	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", fmt.Errorf("read byte: %w", err)
		}
		if b == 0 {
			break
		}
		buf = append(buf, b)
	}
	return string(buf), nil
}
