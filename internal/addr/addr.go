package addr

import (
	"cmp"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Addr struct {
	Hostname string
	Scheme   ProtoScheme
	Port     uint16
}

func (a *Addr) Host() string {
	var suffix string
	if a.Port != 0 {
		suffix = ":" + strconv.FormatUint(uint64(a.Port), 10)
	}
	return a.Hostname + suffix
}

func (a *Addr) String() string {
	prefix := a.Scheme
	if prefix != "" {
		prefix += "://"
	}
	return fmt.Sprintf("%v%v", prefix, a.Host())
}

func ParseAddr(addr string) (*Addr, error) {
	raddr, err := parseRawAddr(addr)
	if err != nil {
		return nil, err
	}

	var scheme ProtoScheme
	if raddr.scheme != "" {
		scheme, err = ParseScheme(raddr.scheme)
		if err != nil {
			return nil, err
		}
	}

	portNum := scheme.Port()
	if raddr.port != "" {
		portNum, err = ParsePort(raddr.port)
		if err != nil {
			return nil, fmt.Errorf("port number %q: %w", raddr.port, err)
		}
	}

	return &Addr{
		Scheme:   scheme,
		Hostname: cmp.Or(raddr.hostname, "localhost"),
		Port:     portNum,
	}, nil
}

type rawAddr struct {
	scheme   string
	hostname string
	port     string
}

var addrRgx = regexp.MustCompile(`\A(?:(?<SCHEME>[^:]+):)?(?://)?(?<HOSTNAME>[^:]+)?(?::(?<PORT>[^:]+))?\z`)

func parseRawAddr(addr string) (raddr rawAddr, err error) {
	matches := addrRgx.FindStringSubmatch(addr)
	if matches == nil {
		err = fmt.Errorf("invalid network address %q", addr)
	} else {
		raddr = makeRawAddr(
			matches[addrRgx.SubexpIndex("SCHEME")],
			matches[addrRgx.SubexpIndex("HOSTNAME")],
			matches[addrRgx.SubexpIndex("PORT")],
		)
	}
	return
}

func makeRawAddr(scheme, hostname, port string) (raddr rawAddr) {
	// normalize all names to lowercase
	scheme = strings.ToLower(scheme)
	hostname = strings.ToLower(hostname)
	port = strings.ToLower(port)

	if hostname != "" {
		if scheme != "" {
			if port == "" {
				return makeRawAddr2(scheme, hostname)
			}
		} else if port != "" {
			return makeRawAddr2(hostname, port)
		} else {
			return makeRawAddr1(hostname)
		}
	}

	raddr.scheme = scheme
	raddr.hostname = hostname
	raddr.port = port
	return
}

func makeRawAddr1(str string) (raddr rawAddr) {
	switch {
	case IsValidScheme(str):
		raddr.scheme = str
	case IsValidPort(str):
		raddr.port = str
	default:
		raddr.hostname = str
	}
	return
}

func makeRawAddr2(str1, str2 string) (raddr rawAddr) {
	if IsValidScheme(str1) {
		raddr.scheme = str1
		if IsValidPort(str2) {
			raddr.port = str2
		} else {
			raddr.hostname = str2
		}
	} else {
		raddr.hostname = str1
		raddr.port = str2
	}
	return
}
