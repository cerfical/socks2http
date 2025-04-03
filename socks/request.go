package socks

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/cerfical/socks2http/addr"
)

const (
	V4             = 4
	RequestConnect = 1
)

func ReadRequest(r *bufio.Reader) (*Request, error) {
	req := Request{}
	if err := binary.Read(r, binary.BigEndian, &req.Header); err != nil {
		return nil, err
	}

	if req.Version != V4 {
		return nil, fmt.Errorf("invalid version number %v", req.Version)
	}

	if req.Command != RequestConnect {
		return nil, fmt.Errorf("invalid command code %v", req.Command)
	}

	// read a null-terminated username string
	user := []byte{}
	for {
		b, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		if b == 0 {
			break
		}
		user = append(user, b)
	}
	req.User = string(user)

	return &req, nil
}

type Request struct {
	Header
	User string
}

type Header struct {
	Version  byte
	Command  byte
	DestPort uint16
	DestIP   addr.IPv4
}

func (r *Request) Write(w io.Writer) error {
	bytes := make([]byte, unsafe.Sizeof(r.Header))
	if _, err := binary.Encode(bytes, binary.BigEndian, &r.Header); err != nil {
		return err
	}
	bytes = append(bytes, append([]byte(r.User), 0)...)

	if _, err := w.Write(bytes); err != nil {
		return err
	}
	return nil
}
