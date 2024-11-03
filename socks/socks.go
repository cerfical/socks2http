package socks

import (
	"encoding/binary"
	"errors"
	"net"
	"unsafe"
)

const (
	socksVersion        = 4
	connectRequest      = 1
	accessGranted       = 90
	accessRejected      = 91
	accessIdentRequired = 92
	accessIdentFailed   = 93
)

type ip4Addr struct {
	Data [4]byte
}

func Connect(conn net.Conn, addr net.IP, port uint16) error {
	if ip4 := addr.To4(); ip4 == nil {
		return errors.New("only IPv4 addresses are supported")
	}
	req := newRequest(connectRequest, ip4Addr{Data: [4]byte(addr)}, port)

	// write a SOCK4 request header, followed by an empty null-terminated userId
	buf := make([]byte, unsafe.Sizeof(req)+1)
	if _, err := binary.Encode(buf, binary.BigEndian, req); err != nil {
		return err
	}
	if _, err := conn.Write(buf); err != nil {
		return err
	}

	reply := socksReply{}
	if err := binary.Read(conn, binary.BigEndian, &reply); err != nil {
		return err
	}
	if reply.Version != 0 || reply.DestAddr.Data != [4]byte{0} || reply.DestPort != 0 {
		return errors.New("invalid SOCKS4 reply")
	}
	return checkReplyCode(reply.Code)
}

func newRequest(command byte, destAddr ip4Addr, destPort uint16) socksRequest {
	return socksRequest{
		Version:  socksVersion,
		Command:  command,
		DestPort: destPort,
		DestAddr: destAddr,
	}
}

type socksRequest struct {
	Version  byte
	Command  byte
	DestPort uint16
	DestAddr ip4Addr
}

type socksReply struct {
	Version  byte
	Code     byte
	DestPort uint16
	DestAddr ip4Addr
}

func checkReplyCode(code byte) error {
	if code == accessGranted {
		return nil
	}

	msg := ""
	switch code {
	case accessRejected:
		msg = "access rejected or failed"
	case accessIdentRequired:
		msg = "access rejected: no Ident service"
	case accessIdentFailed:
		msg = "access rejected: Ident auth failed"
	default:
		msg = "uknown reply code"
	}
	return errors.New(msg)
}
