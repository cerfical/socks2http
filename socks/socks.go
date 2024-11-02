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

type Ip4Addr struct {
	Data [4]byte
}

func MakeIp4Addr(addr net.IP) Ip4Addr {
	ipv4 := addr.To4()
	if ipv4 == nil {
		return Ip4Addr{}
	}
	return Ip4Addr{Data: [4]byte(addr)}
}

func Connect(conn net.Conn, addr Ip4Addr, port uint16) error {
	req := newRequest(connectRequest, addr, port)

	buf := make([]byte, unsafe.Sizeof(req)+1)
	_, err := binary.Encode(buf, binary.BigEndian, req)
	if err != nil {
		return err
	}

	// use empty usedId
	buf = append(buf, byte('\x00'))
	_, err = conn.Write(buf)
	return err
}

func newRequest(command byte, destAddr Ip4Addr, destPort uint16) socksRequest {
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
	DestAddr Ip4Addr
}

func GetReply(conn net.Conn) error {
	reply := socksReply{}
	err := binary.Read(conn, binary.BigEndian, &reply)
	if err != nil {
		return err
	}
	return errorFromReplyCode(reply.Code)
}

type socksReply struct {
	Version  byte
	Code     byte
	DestPort uint16
	DestAddr Ip4Addr
}

func errorFromReplyCode(code byte) error {
	switch code {
	case accessGranted:
		return nil
	case accessRejected:
		return errors.New("socks: access rejected or failed")
	case accessIdentRequired:
		return errors.New("socks: access rejected: no Ident service")
	case accessIdentFailed:
		return errors.New("socks: access rejected: Ident auth failed")
	default:
		return errors.New("socks: uknown reply code")
	}
}
