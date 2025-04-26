package socks4

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"unsafe"

	"github.com/cerfical/socks2http/addr"
)

func NewRequest(c Command, dstAddr *addr.Host) *Request {
	r := Request{
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

	if v := version[0]; v != VersionCode {
		return nil, fmt.Errorf("invalid version (%v)", hexByte(v))
	}

	var h requestHeader
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}

	un, err := readNullString(r)
	if err != nil {
		return nil, fmt.Errorf("decode username: %w", err)
	}

	dstAddr := addr.NewHost(h.DstIP.String(), h.DstPort)
	if strings.HasPrefix(dstAddr.Hostname, "0.0.0") {
		// The destination address is a hostname
		hn, err := readNullString(r)
		if err != nil {
			return nil, fmt.Errorf("decode destination hostname: %w", err)
		}
		dstAddr.Hostname = hn
	}

	req := Request{
		Command:  Command(h.Command),
		DstAddr:  *dstAddr,
		Username: un,
	}
	return &req, nil
}

type Request struct {
	Command  Command
	DstAddr  addr.Host
	Username string
}

func (r *Request) Write(w io.Writer) error {
	var (
		dstIP       addr.IPv4
		dstHostname string
	)

	if ip4, ok := r.DstAddr.ToIPv4(); ok {
		dstIP = ip4
	} else {
		dstHostname = r.DstAddr.Hostname
	}

	h := requestHeader{
		Version: VersionCode,
		Command: byte(r.Command),
		DstPort: r.DstAddr.Port,
		DstIP:   dstIP,
	}

	bytes := make([]byte, unsafe.Sizeof(h))
	if _, err := binary.Encode(bytes, binary.BigEndian, &h); err != nil {
		return fmt.Errorf("encode header: %w", err)
	}

	// Append the username bytes
	bytes = append(bytes, []byte(r.Username)...)
	bytes = append(bytes, 0)

	// Append the destination hostname, if any
	if len(dstHostname) > 0 {
		bytes = append(bytes, []byte(dstHostname)...)
		bytes = append(bytes, 0)
	}

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
