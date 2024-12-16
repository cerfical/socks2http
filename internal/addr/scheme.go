package addr

const (
	Direct = "direct"
	SOCKS4 = "socks4"
	HTTP   = "http"
)

func isValidScheme(scheme string) bool {
	switch scheme {
	case Direct, SOCKS4, HTTP:
		return true
	}
	return false
}

func defaultProxyPort(scheme string) uint16 {
	switch scheme {
	case SOCKS4:
		return 1080
	case HTTP:
		return 8080
	default:
		return 0
	}
}
