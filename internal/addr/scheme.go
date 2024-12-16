package addr

const (
	// Direct is a pseudo-scheme that requests a direct connection to a server without an intermediate proxy.
	Direct = "direct"

	// SOCKS4 requests the SOCKS4 protocol.
	SOCKS4 = "socks4"

	// HTTP requests the HTTP protocol.
	HTTP = "http"
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
