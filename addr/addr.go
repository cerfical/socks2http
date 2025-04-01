package addr

import (
	"cmp"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	Direct = "direct"
	SOCKS4 = "socks4"
	HTTP   = "http"
)

func New(scheme, hostname string, port uint16) *Addr {
	return &Addr{Scheme: scheme, Hostname: hostname, Port: port}
}

type Addr struct {
	Scheme   string
	Hostname string
	Port     uint16
}

// Host presents [Addr] as a <hostname>:<port> string.
func (a *Addr) Host() string {
	p := ""
	if a.Port != 0 {
		p = strconv.Itoa(int(a.Port))
	}
	return a.Hostname + ":" + p
}

// UnmarshalText parses [Addr] from its text representation.
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
func (a *Addr) UnmarshalText(text []byte) error {
	matches := rgx.FindSubmatch(text)
	if matches == nil {
		return errors.New("malformed network address")
	}

	// group named captures by name
	submatches := make(map[string]string)
	for i, n := range rgx.SubexpNames() {
		submatches[n] = cmp.Or(submatches[n], string(matches[i]))
	}

	// provide some reasonable non-empty default values
	scheme := strings.ToLower(cmp.Or(submatches["SCHEME"], HTTP))
	hostname := strings.ToLower(cmp.Or(submatches["HOSTNAME"], defSchemeHostname(scheme)))
	port := strings.ToLower(cmp.Or(submatches["PORT"], defSchemePort(scheme)))

	if port != "" {
		p, err := ParsePort(port)
		if err != nil {
			return fmt.Errorf("parsing port number %v: %w", port, err)
		}
		a.Port = p
	} else {
		a.Port = 0
	}

	a.Scheme = scheme
	a.Hostname = hostname

	return nil
}

var (
	s   = `(?<SCHEME>[^:]+)`
	h   = `(?<HOSTNAME>[^:]+)`
	p   = `(?<PORT>[^:]+)`
	rgx = regexp.MustCompile(fmt.Sprintf(`\A(?:(?:%[1]v:)?(?://%[2]v)?(?::%[3]v)?|%[2]v:%[3]v|%[1]v)\z`, s, h, p))
)

func defSchemeHostname(scheme string) string {
	switch scheme {
	case Direct:
		return ""
	default:
		return "localhost"
	}
}

func defSchemePort(scheme string) string {
	switch scheme {
	case SOCKS4:
		return "1080"
	case HTTP:
		return "8080"
	default:
		return ""
	}
}

func (a *Addr) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Addr) String() string {
	s, h, p := a.Scheme, a.Hostname, a.Port
	switch {
	case s != "" && h == "" && p == 0:
		return s
	case p != 0 && h != "" && s == "":
		return a.Host()
	default:
		s := s
		if s != "" {
			s += ":"
		}
		if h != "" {
			s += "//"
		}
		if p != 0 {
			s += a.Host()
		} else {
			s += a.Hostname
		}
		return s
	}
}
