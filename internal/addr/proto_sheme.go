package addr

import "fmt"

type ProtoScheme uint16

const (
	Direct ProtoScheme = 0
	SOCKS4 ProtoScheme = 1080
	HTTP   ProtoScheme = 8080
)

func (s ProtoScheme) Port() uint16 {
	return uint16(s)
}

func (s ProtoScheme) String() string {
	switch s {
	case SOCKS4:
		return "socks4"
	case HTTP:
		return "http"
	default:
		panic(fmt.Sprintf("invalid protocol scheme \"%d\"", s))
	}
}

func ParseScheme(scheme string) (ProtoScheme, error) {
	switch scheme {
	case "socks4":
		return SOCKS4, nil
	case "http":
		return HTTP, nil
	case "direct":
		return Direct, nil
	default:
		return 0, fmt.Errorf("unknown protocol scheme %q", scheme)
	}
}

func IsValidScheme(scheme string) bool {
	_, err := ParseScheme(scheme)
	return err == nil
}
