package addr

import (
	"fmt"
	"strings"
)

type ProtoScheme string

const (
	Direct ProtoScheme = "direct"
	SOCKS4 ProtoScheme = "socks4"
	HTTP   ProtoScheme = "http"
)

func (s ProtoScheme) Port() uint16 {
	switch s {
	case SOCKS4:
		return 1080
	case HTTP:
		return 8080
	default:
		return 0
	}
}

func (s ProtoScheme) String() string {
	return string(s)
}

func ParseScheme(scheme string) (ProtoScheme, error) {
	switch ProtoScheme(scheme) {
	case Direct, SOCKS4, HTTP, "":
		return ProtoScheme(strings.ToLower(scheme)), nil
	}
	return "", fmt.Errorf("unsupported protocol scheme %q", scheme)
}

func IsValidScheme(scheme string) bool {
	_, err := ParseScheme(scheme)
	return err == nil
}
