package socks4

import "fmt"

type hexByte byte

func (b hexByte) String() string {
	return fmt.Sprintf("%#02x", byte(b))
}
