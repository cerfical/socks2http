package addr

import (
	"cmp"
	"errors"
	"regexp"
	"strconv"
	"strings"
)

const (
	// Direct is a pseudo-scheme that requests a direct connection to a server without an intermediate proxy.
	Direct = "direct"

	// SOCKS4 requests the SOCKS4 protocol.
	SOCKS4 = "socks4"

	// HTTP requests the HTTP protocol.
	HTTP = "http"
)

// ParsePort converts a string to a valid port number.
func ParsePort(port string) (uint16, error) {
	portNum, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(portNum), nil
}

// Addr represents a reduced set of [net/url.URL] network addresses.
type Addr struct {
	// Scheme represents the scheme [Addr] component.
	Scheme string

	// Hostname represents the hostname [Addr] component.
	Hostname string

	// Port represents the port [Addr] component.
	Port uint16
}

// Host presents [Addr] as a "<hostname>:<port>" string.
func (a *Addr) Host() string {
	return a.Hostname + ":" + strconv.Itoa(int(a.Port))
}

// UnmarshalText creates a new [Addr] from a text representation.
// The syntax is similar to [net/url.URL], but simplified for ease of use.
//
// For example, the address http://localhost:8080 can be represented as:
//   - http://localhost:8080
//   - http://localhost
//   - http::8080
//   - //localhost:8080
//   - //localhost
//   - :8080
//   - http:
//
// By default, assumes HTTP protocol scheme and localhost for the hostname.
// If no port is specified, it is inferred from the scheme.
func (a *Addr) UnmarshalText(text []byte) error {
	matches := rgx.FindStringSubmatch(string(text))
	if matches == nil {
		return errors.New("malformed network address")
	}

	scheme := strings.ToLower(string(matches[rgx.SubexpIndex("SCHEME")]))
	hostname := strings.ToLower(string(matches[rgx.SubexpIndex("HOSTNAME")]))
	port := string(matches[rgx.SubexpIndex("PORT")])

	// provide some reasonable non-empty default values
	scheme = cmp.Or(scheme, HTTP)
	hostname = cmp.Or(hostname, "localhost")

	portNum := defaultProxyPort(scheme)
	if port != "" {
		p, err := ParsePort(port)
		if err != nil {
			return err
		}
		portNum = p
	}

	a.Scheme = scheme
	a.Hostname = hostname
	a.Port = portNum
	return nil
}

var rgx = regexp.MustCompile(`\A(((?<SCHEME>[^:]+):)?(//(?<HOSTNAME>[^:/]+))?(:(?<PORT>[^:]+))?)\z`)

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

// MarshalText converts [Addr] into a compact text representation.
func (a *Addr) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// String presents [Addr] as a "<scheme>://<hostname>:<port>" string.
func (a *Addr) String() string {
	s := a.Scheme
	if s != "" {
		s = s + ":"
	}

	h := a.Host()
	if strings.HasPrefix(h, ":") {
		return s + h
	}
	return s + "//" + h
}
