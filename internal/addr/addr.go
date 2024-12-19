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

// Host presents [Addr] as a "<hostname>:<port>" string.
func (a *Addr) Host() string {
	return a.Hostname() + ":" + a.portString()
}

// String presents [Addr] as a "<scheme>://<hostname>:<port>" string.
func (a *Addr) String() string {
	return a.Scheme() + "://" + a.Host()
}

// UnmarshalText creates a new [Addr] from a compact text representation.
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
func (a *Addr) UnmarshalText(text []byte) error {
	raddr, ok := parseRaw(string(text))
	if !ok {
		return errors.New("malformed network address")
	}

	// make sure the scheme has some reasonable non-empty value
	if raddr.scheme == "" {
		raddr.scheme = HTTP
	} else if !isValidScheme(raddr.scheme) {
		return fmt.Errorf("unsupported protocol scheme %q", raddr.scheme)
	}

	if raddr.hostname == "" {
		raddr.hostname = "localhost"
	}

	portNum := defaultProxyPort(raddr.scheme)
	if raddr.port != "" {
		p, err := ParsePort(raddr.port)
		if err != nil {
			return err
		}
		portNum = p
	}

	a.scheme = raddr.scheme
	a.hostname = raddr.hostname
	a.port = portNum
	return nil
}

// MarshalText converts [Addr] into a compact text representation.
func (a *Addr) MarshalText() ([]byte, error) {
	var str string
	switch s, hn := a.Scheme(), a.Hostname(); {
	case s != "" && hn != "":
		str = s + "://" + hn
	case s != "":
		str = s
	default:
		str = hn
	}

	if str != "" {
		str += ":"
	}
	return []byte(str + a.portString()), nil
}

func (a *Addr) portString() string {
	return strconv.Itoa(int(a.Port()))
}
