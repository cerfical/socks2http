package addr

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

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

type Addr struct {
	scheme   string
	hostname string
	port     uint16
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
