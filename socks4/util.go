package socks4

import "fmt"

const VersionCode = 0x04

type hexByte byte

func (b hexByte) String() string {
	return fmt.Sprintf("%#02x", byte(b))
}
