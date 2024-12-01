package socks

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"socks2http/internal/addr"
	"socks2http/internal/util"
	"time"
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
	IP   [4]byte
}

func Connect(socksHost, destHost addr.Host) (net.Conn, error) {
	return ConnectTimeout(socksHost, destHost, 0)
}

func ConnectTimeout(socksHost, destHost addr.Host, timeout time.Duration) (net.Conn, error) {
	tcpAddr, err := resolveHost(destHost)
	if err != nil {
		return nil, fmt.Errorf("address resolution failed: %w", err)
	}

	socksConn, err := net.DialTimeout("tcp", socksHost.String(), timeout)
	if err != nil {
		return nil, fmt.Errorf("connection to SOCKS proxy failed: %w", err)
	}

	if err = sendConnectRequest(socksConn, tcpAddr); err != nil {
		util.Must(socksConn.Close())
		return nil, fmt.Errorf("connection to destination server failed: %w", err)
	}
	return socksConn, nil
}

func resolveHost(host addr.Host) (tcp4Addr, error) {
	ip, err := net.ResolveIPAddr("ip4", host.Hostname)
	if err != nil {
		return tcp4Addr{}, err
	}
	return tcp4Addr{Port: host.Port, IP: [4]byte(ip.IP)}, nil
}

func sendConnectRequest(socksConn net.Conn, destAddr tcp4Addr) error {
	req := socksRequest{
		Version:  socksVersion,
		Command:  connectRequest,
		DestAddr: destAddr,
	}

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
		return fmt.Errorf("failed to receive a client reply: %w", err)
	}
	if reply.Version != 0 || reply.DestAddr.IP != [4]byte{0} || reply.DestAddr.Port != 0 {
		return errors.New("invalid SOCKS reply")
	}
	return checkReplyCode(reply.Command)
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
