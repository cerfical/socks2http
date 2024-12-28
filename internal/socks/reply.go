package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	RequestGranted            = 90
	RequestRejectedOrFailed   = 91
	RequestRejectedNoAuth     = 92
	RequestRejectedAuthFailed = 93
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
	case RequestGranted:
		return nil
	case RequestRejectedOrFailed:
		msg = "request rejected or failed"
	case RequestRejectedNoAuth:
		msg = "request rejected: failed to connect to authentication service"
	case RequestRejectedAuthFailed:
		msg = "request rejected: authentication failure"
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
	if err := binary.Write(w, binary.BigEndian, r); err != nil {
		return err
	}
	return nil
}
