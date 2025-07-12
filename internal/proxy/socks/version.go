package socks

import "fmt"

const (
	V4 Version = 0x04
	V5 Version = 0x05
)

var versions = map[Version]string{
	V4: "SOCKS4",
	V5: "SOCKS5",
}

type Version byte

func (v Version) String() string {
	if str, ok := versions[v]; ok {
		return fmt.Sprintf("%v", str)
	}
	return fmt.Sprintf("%#02x", byte(v))
}

func (v Version) MarshalText() ([]byte, error) {
	return []byte(v.String()), nil
}
