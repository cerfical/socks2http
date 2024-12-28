package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

func ReadReply(r io.Reader) error {
	rep := reply{}
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

type reply struct {
	Version byte
	Code    byte
	_       uint16
	_       [4]byte
}

func checkReplyCode(code byte) error {
	const (
		accessGranted   = 90
		accessRejected  = 91
		noIdentdService = 92
		authFailed      = 93
	)

	msg := ""
	switch code {
	case accessGranted:
		return nil
	case accessRejected:
		msg = "access rejected"
	case noIdentdService:
		msg = "failed to connect to identd service"
	case authFailed:
		msg = "identd authentication failed"
	default:
		msg = fmt.Sprintf("unexpected reply code %v", code)
	}
	return errors.New(msg)
}
