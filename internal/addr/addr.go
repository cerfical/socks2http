package addr

import (
	"cmp"
	"errors"
	"fmt"
	"net"
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

// Addr represents a reduced set of [net/url.URL] network addresses.
type Addr struct {
	// Scheme represents the scheme [Addr] component.
	Scheme string

	// Hostname represents the hostname [Addr] component.
	Hostname string

	// Port represents the port [Addr] component.
	Port uint16
}

// Host presents [Addr] as a <hostname>:<port> string.
func (a *Addr) Host() string {
	p := ""
	if a.Port != 0 {
		p = strconv.Itoa(int(a.Port))
	}
	return net.JoinHostPort(a.Hostname, p)
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
	matches := rgx.FindSubmatch(text)
	if matches == nil {
		return errors.New("malformed network address")
	}

	scheme := string(matches[rgx.SubexpIndex("SCHEME")])
	hostname := string(matches[rgx.SubexpIndex("HOSTNAME")])
	port := string(matches[rgx.SubexpIndex("PORT")])

	// provide some reasonable non-empty default values
	scheme = strings.ToLower(cmp.Or(scheme, HTTP))
	hostname = strings.ToLower(cmp.Or(hostname, defaultHostnameForScheme(scheme)))
	port = strings.ToLower(cmp.Or(port, defaultProxyPortForScheme(scheme)))

	if port != "" {
		p, err := ParsePort(port)
		if err != nil {
			return fmt.Errorf("parse port number %v: %w", port, err)
		}
		a.Port = p
	} else {
		a.Port = 0
	}

	a.Scheme = scheme
	a.Hostname = hostname

	return nil
}

var rgx = regexp.MustCompile(`\A(((?<SCHEME>[^:]+):)?(//(?<HOSTNAME>[^:/]+))?(:(?<PORT>[^:]+))?)\z`)

func defaultHostnameForScheme(scheme string) string {
	switch scheme {
	case Direct:
		return ""
	default:
		return "localhost"
	}
}

func defaultProxyPortForScheme(scheme string) string {
	switch scheme {
	case SOCKS4:
		return "1080"
	case HTTP:
		return "8080"
	default:
		return ""
	}
}

// MarshalText converts [Addr] into a compact text representation.
func (a *Addr) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// String presents [Addr] as a <scheme>://<hostname>:<port> string.
func (a *Addr) String() string {
	s := a.Scheme
	if s != "" {
		s = s + ":"
	}

	switch {
	case a.Hostname != "" && a.Port != 0:
		return s + "//" + a.Host()
	case a.Hostname != "":
		return s + "//" + a.Hostname
	case a.Port != 0:
		return s + a.Host()
	default:
		return s
	}
}
