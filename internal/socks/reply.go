package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"unsafe"
)

const (
	AccessGranted = 90
	AccessDenied  = 91
	NoIdentd      = 92
	AuthFailed    = 93
)

func ReadReply(r io.Reader) error {
	rep := Reply{}
	if err := binary.Read(r, binary.BigEndian, &rep); err != nil {
		return err
	}

	if rep.Version != 0 {
		return fmt.Errorf("unexpected version number %v", rep.Version)
	}
	if err := checkReplyCode(rep.Code); err != nil {
		return err
	}
	return nil
}

func checkReplyCode(code byte) error {
	msg := ""
	switch code {
	case AccessGranted:
		return nil
	case AccessDenied:
		msg = "access denied"
	case NoIdentd:
		msg = "failed to connect to identd service"
	case AuthFailed:
		msg = "identd authentication failed"
	default:
		msg = fmt.Sprintf("unexpected reply code %v", code)
	}
	return errors.New(msg)
}

type Reply struct {
	Version byte
	Code    byte
	_       uint16
	_       [4]byte
}

func (r *Reply) Write(w io.Writer) error {
	bytes := make([]byte, unsafe.Sizeof(*r))
	if _, err := binary.Encode(bytes, binary.BigEndian, r); err != nil {
		return err
	}

	if _, err := w.Write(bytes); err != nil {
		return err
	}
	return nil
}
