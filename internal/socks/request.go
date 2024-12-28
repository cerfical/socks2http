package socks

import (
	"encoding/binary"
	"io"

	"github.com/cerfical/socks2http/internal/addr"
)

const (
	V4             = 4
	ConnectCommand = 1
)

type Request struct {
	Version  byte
	Command  byte
	DestPort uint16
	DestIP   addr.IPv4Addr
	User     string
}

func (r *Request) Write(w io.Writer) error {
	// preallocate buffer large enough to hold the serialized request
	bytes := make([]byte, 0, 2+2+4+len(r.User)+1)

	bytes = append(bytes, r.Version, r.Command)              // 2 bytes
	bytes = binary.BigEndian.AppendUint16(bytes, r.DestPort) // 2 bytes
	bytes = append(bytes, r.DestIP[:]...)                    // 4 bytes
	bytes = append(bytes, []byte(r.User)...)                 // len(User) bytes
	bytes = append(bytes, 0)                                 // 1 byte

	if _, err := w.Write(bytes); err != nil {
		return err
	}
	return nil
}
