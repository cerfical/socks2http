package addr

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Addr represents a reduced set of [net/url.URL] network addresses.
type Addr struct {
	scheme   string
	hostname string
	port     uint16
}

// New creates a new [Addr] from the specified [net/url.URL] components.
// By default, assumes HTTP protocol scheme and localhost.
// If port is zero, it is inferred from the scheme.
func New(scheme, hostname string, port uint16) *Addr {
	s := strings.ToLower(cmp.Or(scheme, HTTP))
	h := strings.ToLower(cmp.Or(hostname, "localhost"))
	p := cmp.Or(port, defaultProxyPort(s))

	return &Addr{
		scheme:   s,
		hostname: h,
		port:     p,
	}
}

// Parse constructs a new [Addr] from a string.
// The syntax is similar to the one of [net/url.URL], but is greatly simplified for ease of use.
//
// For example, the address http://localhost:8080 can be represented as:
//   - http://localhost:8080
//   - http:localhost
//   - http:8080
//   - localhost:8080
//   - http
//   - localhost
//   - 8080
//
// Empty components will be set according to [New].
func Parse(addr string) (*Addr, error) {
	raddr, ok := parseRaw(addr)
	if !ok {
		return nil, errors.New("malformed network address")
	}

	port, err := ParsePort(raddr.port)
	if err != nil {
		return nil, err
	}

	if raddr.scheme != "" && !isValidScheme(raddr.scheme) {
		return nil, fmt.Errorf("unsupported protocol scheme %q", raddr.scheme)
	}

	return New(raddr.scheme, raddr.hostname, port), nil
}

func (a *Addr) Scheme() string {
	return a.scheme
}

func (a *Addr) Hostname() string {
	return a.hostname
}

func (a *Addr) Port() uint16 {
	return a.port
}

func (a *Addr) Host() string {
	return a.Hostname() + ":" + strconv.Itoa(int(a.Port()))
}

func (a *Addr) String() string {
	return a.Scheme() + "://" + a.Host()
}
