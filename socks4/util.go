package socks4

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

const VersionCode = 0x04

type hexByte byte

func (b hexByte) String() string {
	return fmt.Sprintf("%#02x", byte(b))
}

func checkVersion(r *bufio.Reader, version byte) error {
	versionBytes, err := r.Peek(1)
	if err != nil {
		return err
	}
	if v := versionBytes[0]; v != version {
		return fmt.Errorf("%w (%v)", ErrInvalidVersion, hexByte(v))
	}
	return nil
}

func readAddr(r *bufio.Reader) (net.IP, uint16, error) {
	port, err := readPort(r)
	if err != nil {
		return nil, 0, fmt.Errorf("decode port: %w", err)
	}

	var ip4 [4]byte
	if _, err := io.ReadFull(r, ip4[:]); err != nil {
		return nil, 0, fmt.Errorf("decode IPv4 address: %w", err)
	}
	return net.IP(ip4[:]), port, nil
}

func readPort(r *bufio.Reader) (uint16, error) {
	var port [2]byte
	if _, err := io.ReadFull(r, port[:]); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(port[:]), nil
}
