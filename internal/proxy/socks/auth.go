package socks

import "fmt"

const (
	AuthNone Auth = 0x00

	AuthNotAcceptable Auth = 0xff
)

var authText = map[Auth]string{
	AuthNone: "None",

	AuthNotAcceptable: "No Acceptable Authentication",
}

type Auth byte

func (a Auth) String() string {
	if str, ok := authText[a]; ok {
		return fmt.Sprintf("%v", str)
	}
	return fmt.Sprintf("%#02x", byte(a))
}

func (a Auth) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}
