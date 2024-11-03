package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
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

type tcp4Addr struct {
	Port uint16
	Ip   [4]byte
}

func Dial(socksAddr string, destAddr string) (net.Conn, error) {
	tcpAddr, err := resolveAddr(destAddr)
	if err != nil {
		return nil, err
	}

	socksConn, err := net.Dial("tcp", socksAddr)
	if err != nil {
		return nil, err
	}

	if err = connect(socksConn, tcpAddr); err != nil {
		if clsErr := socksConn.Close(); clsErr != nil {
			return nil, fmt.Errorf(err.Error()+": %w", clsErr)
		}
		return nil, err
	}
	return socksConn, nil
}

func resolveAddr(addr string) (tcp4Addr, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return tcp4Addr{}, err
	}

	ip, err := resolveIp4Addr(host)
	if err != nil {
		return tcp4Addr{}, err
	}

	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return tcp4Addr{}, err
	}
	return tcp4Addr{Port: uint16(portNum), Ip: [4]byte(ip)}, nil
}

func resolveIp4Addr(host string) (net.IP, error) {
	ip, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, err
	}

	if ip4 := ip.IP.To4(); ip4 == nil {
		return nil, errors.New("only IPv4 addresses are supported")
	}
	return ip.IP, nil
}

func connect(socksConn net.Conn, destAddr tcp4Addr) error {
	req := newRequest(connectRequest, destAddr)

	// write a SOCK4 request header, followed by an empty null-terminated userId
	buf := make([]byte, unsafe.Sizeof(req)+1)
	if _, err := binary.Encode(buf, binary.BigEndian, req); err != nil {
		return err
	}
	if _, err := socksConn.Write(buf); err != nil {
		return err
	}

	reply := socksRequest{}
	if err := binary.Read(socksConn, binary.BigEndian, &reply); err != nil {
		return err
	}
	if reply.Version != 0 || reply.DestAddr.Ip != [4]byte{0} || reply.DestAddr.Port != 0 {
		return errors.New("invalid SOCKS reply")
	}
	return checkReplyCode(reply.Command)
}

func newRequest(command byte, destAddr tcp4Addr) socksRequest {
	return socksRequest{
		Version:  socksVersion,
		Command:  command,
		DestAddr: destAddr,
	}
}

type socksRequest struct {
	Version  byte
	Command  byte
	DestAddr tcp4Addr
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
