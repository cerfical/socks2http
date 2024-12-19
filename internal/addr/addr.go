package addr

import (
	"errors"
	"fmt"
	"strconv"
)

// Addr represents a reduced set of [net/url.URL] network addresses.
type Addr struct {
	scheme   string
	hostname string
	port     uint16
}

// New creates a new [Addr] from the specified [net/url.URL] components.
// Does not perform validation of the supplied arguments.
func New(scheme, hostname string, port uint16) *Addr {
	return &Addr{
		scheme:   scheme,
		hostname: hostname,
		port:     port,
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
// By default, assumes HTTP protocol scheme and localhost for the hostname.
// If no port was specified, it is inferred from the scheme.
func Parse(addr string) (*Addr, error) {
	raddr, ok := parseRaw(addr)
	if !ok {
		return nil, errors.New("malformed network address")
	}

	// make sure the scheme has some reasonable non-empty value
	if raddr.scheme == "" {
		raddr.scheme = HTTP
	} else if !isValidScheme(raddr.scheme) {
		return nil, fmt.Errorf("unsupported protocol scheme %q", raddr.scheme)
	}

	if raddr.hostname == "" {
		raddr.hostname = "localhost"
	}

	portNum := defaultProxyPort(raddr.scheme)
	if raddr.port != "" {
		p, err := ParsePort(raddr.port)
		if err != nil {
			return nil, err
		}
		portNum = p
	}

	return New(raddr.scheme, raddr.hostname, portNum), nil
}

// Scheme returns the scheme component of [Addr].
func (a *Addr) Scheme() string {
	return a.scheme
}

// Hostname returns the hostname component of [Addr].
func (a *Addr) Hostname() string {
	return a.hostname
}

// Port returns the port component of [Addr].
func (a *Addr) Port() uint16 {
	return a.port
}

// Host presents [Addr] as a string "<hostname>:<port>".
func (a *Addr) Host() string {
	return a.Hostname() + ":" + strconv.Itoa(int(a.Port()))
}

// String presents [Addr] as a string "<scheme>://<hostname>:<port>".
func (a *Addr) String() string {
	return a.Scheme() + "://" + a.Host()
}
