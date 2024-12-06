package addr

import (
	"fmt"
	"strings"
)

type Scheme string

const (
	Direct Scheme = "direct"
	SOCKS4 Scheme = "socks4"
	HTTP   Scheme = "http"
)

func (s Scheme) Port() uint16 {
	switch s {
	case SOCKS4:
		return 1080
	case HTTP:
		return 8080
	default:
		return 0
	}
}

func (s Scheme) String() string {
	return string(s)
}

func ParseScheme(scheme string) (Scheme, error) {
	switch scheme := Scheme(strings.ToLower(scheme)); scheme {
	case Direct, SOCKS4, HTTP, "":
		return scheme, nil
	}
	return "", fmt.Errorf("unsupported protocol scheme %q", scheme)
}

func IsValidScheme(scheme string) bool {
	_, err := ParseScheme(scheme)
	return err == nil
}
