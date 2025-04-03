package addr

import (
	"cmp"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	Direct = "direct"
	SOCKS4 = "socks4"
	HTTP   = "http"
)

var (
	scheme = `(?<SCHEME>[^:]+)`
	host   = `(?<HOSTNAME>[^:]+)`
	port   = `(?<PORT>[^:]+)`

	addrRgx = regexp.MustCompile(fmt.Sprintf(
		`\A(?:(?:%[1]v:)?(?://%[2]v)?(?::%[3]v)?|%[2]v:%[3]v|%[1]v)\z`,
		scheme, host, port,
	))
)

func New(scheme, hostname string, port uint16) *Addr {
	return &Addr{Scheme: scheme, Host: *NewHost(hostname, port)}
}

// Parse parses [Addr] from its text representation.
// The syntax is similar to [net/url.URL], but simplified for ease of use.
//
// For example, the address http://localhost:8080 can be represented as:
//   - http://localhost:8080
//   - http://localhost
//   - http::8080
//   - localhost:8080
//   - //localhost
//   - http
//   - :8080
//
// By default, assumes HTTP protocol scheme and localhost for the hostname.
// If no port is specified, it is inferred from the scheme.
func Parse(addr string) (*Addr, error) {
	matches := addrRgx.FindStringSubmatch(addr)
	if matches == nil {
		return nil, errors.New("malformed network address")
	}

	// Group named captures by name
	submatches := make(map[string]string)
	for i, n := range addrRgx.SubexpNames() {
		submatches[n] = cmp.Or(submatches[n], string(matches[i]))
	}

	// Provide some reasonable non-empty default values
	scheme := strings.ToLower(cmp.Or(submatches["SCHEME"], HTTP))
	hostname := strings.ToLower(cmp.Or(submatches["HOSTNAME"], defaultHostnameFromScheme(scheme)))
	port := strings.ToLower(cmp.Or(submatches["PORT"], defaultPortFromScheme(scheme)))

	var a Addr
	if port != "" {
		p, err := ParsePort(port)
		if err != nil {
			return nil, fmt.Errorf("parse port number: %w", err)
		}
		a.Host.Port = p
	} else {
		a.Host.Port = 0
	}

	a.Scheme = scheme
	a.Host.Hostname = hostname

	return &a, nil
}

func defaultHostnameFromScheme(scheme string) string {
	switch scheme {
	case Direct:
		return ""
	default:
		return "localhost"
	}
}

func defaultPortFromScheme(scheme string) string {
	switch scheme {
	case SOCKS4:
		return "1080"
	case HTTP:
		return "8080"
	default:
		return ""
	}
}

type Addr struct {
	Scheme string
	Host   Host
}

func (a *Addr) String() string {
	s, h, p := a.Scheme, a.Host.Hostname, a.Host.Port
	switch {
	case s != "" && h == "" && p == 0:
		return s
	case p != 0 && h != "" && s == "":
		return a.Host.String()
	default:
		s := s
		if s != "" {
			s += ":"
		}
		if h != "" {
			s += "//"
		}
		if p != 0 {
			s += a.Host.String()
		} else {
			s += a.Host.Hostname
		}
		return s
	}
}

func (a *Addr) UnmarshalText(text []byte) error {
	addr, err := Parse(string(text))
	if err != nil {
		return err
	}
	*a = *addr
	return nil
}

func (a *Addr) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}
