package socks

import "fmt"

const (
	V4 Version = 0x04
)

var versions = map[Version]string{
	V4: "SOCKS4",
}

func makeVersion(b byte) (v Version, isValid bool) {
	v = Version(b)
	if _, ok := versions[v]; ok {
		return v, true
	}
	return v, false
}

type Version byte

func (v Version) String() string {
	if str, ok := versions[v]; ok {
		return fmt.Sprintf("%v (%v)", str, hexByte(v))
	}
	return ""
}
