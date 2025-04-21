package socks

import (
	"fmt"
	"slices"
)

const (
	V4 Version = iota + 1
)

var versions = []versionInfo{
	{0x04, V4, "SOCKS4"},
}

func decodeVersion(b byte) (Version, bool) {
	i := slices.IndexFunc(versions, func(vi versionInfo) bool {
		return vi.byte == b
	})
	if i != -1 {
		return versions[i].Version, true
	}
	return 0, false
}

func encodeVersion(v Version) (byte, bool) {
	i := slices.IndexFunc(versions, func(vi versionInfo) bool {
		return vi.Version == v
	})
	if i != -1 {
		return versions[i].byte, true
	}
	return 0, false
}

type Version byte

func (v Version) String() string {
	i := slices.IndexFunc(versions, func(vi versionInfo) bool {
		return vi.Version == v
	})
	if i != -1 {
		return fmt.Sprintf("%v (%v)", versions[i].string, hexByte(versions[i].byte))
	}
	return ""
}

type versionInfo struct {
	byte
	Version
	string
}
