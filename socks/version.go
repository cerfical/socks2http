package socks

import "fmt"

const (
	V4 Version = 0x04
)

var versions = map[Version]string{
	V4: "SOCKS4",
}

type Version byte

func (v Version) String() string {
	if str, ok := versions[v]; ok {
		return fmt.Sprintf("%v (%v)", str, hexByte(v))
	}
	return fmt.Sprintf("(%v)", hexByte(v))
}

func isValidVersion(v Version) bool {
	_, ok := versions[v]
	return ok
}
