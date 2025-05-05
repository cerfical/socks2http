package socks5

import "fmt"

const (
	AuthNone AuthMethod = 0x00

	AuthNotAcceptable AuthMethod = 0xff
)

var authMethods = map[AuthMethod]string{
	AuthNone: "No Authentication",

	AuthNotAcceptable: "No Acceptable Authentication Methods",
}

type AuthMethod byte

func (m AuthMethod) String() string {
	if str, ok := authMethods[m]; ok {
		return fmt.Sprintf("%v (%v)", str, hexByte(m))
	}
	return fmt.Sprintf("Invalid Authentication Method (%v)", hexByte(m))
}

func (m AuthMethod) MarshalText() ([]byte, error) {
	return []byte(m.String()), nil
}
