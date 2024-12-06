package addr

import (
	"cmp"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Addr struct {
	Scheme   Scheme
	Hostname string
	Port     uint16
}

func (a *Addr) Host() string {
	port := ""
	if a.Port != 0 {
		port = ":" + strconv.FormatUint(uint64(a.Port), 10)
	}
	return a.Hostname + port
}

func (a *Addr) String() string {
	scheme := a.Scheme
	if scheme != "" {
		scheme += "://"
	}
	return scheme.String() + a.Host()
}

func ParseAddr(addr string) (*Addr, error) {
	raddr, err := parseRawAddr(addr)
	if err != nil {
		return nil, err
	}

	var scheme Scheme
	if raddr.scheme != "" {
		scheme, err = ParseScheme(raddr.scheme)
		if err != nil {
			return nil, err
		}
	}

	port := scheme.Port()
	if raddr.port != "" {
		port, err = ParsePort(raddr.port)
		if err != nil {
			return nil, fmt.Errorf("port number %q: %w", raddr.port, err)
		}
	}

	return &Addr{
		Scheme:   scheme,
		Hostname: cmp.Or(strings.ToLower(raddr.hostname), "localhost"),
		Port:     port,
	}, nil
}

type rawAddr struct {
	scheme   string
	hostname string
	port     string
}

var rgxStr = fmt.Sprintf(`\A(((?<SCHEME>%[1]v)://(?<HOSTNAME>%[1]v)(:(?<PORT>%[1]v))?)|((?<STR1>%[1]v)(:(?<STR2>%[1]v))?))\z`, `[^:]+`)
var rgx = regexp.MustCompile(rgxStr)

func parseRawAddr(addr string) (*rawAddr, error) {
	matches := rgx.FindStringSubmatch(addr)
	if matches == nil {
		return nil, fmt.Errorf("invalid network address %q", addr)
	}

	raddr := &rawAddr{
		scheme:   matches[rgx.SubexpIndex("SCHEME")],
		hostname: matches[rgx.SubexpIndex("HOSTNAME")],
		port:     matches[rgx.SubexpIndex("PORT")],
	}

	// if address a regular URL
	if raddr.scheme != "" {
		return raddr, nil
	}

	str2 := matches[rgx.SubexpIndex("STR2")]
	str1 := matches[rgx.SubexpIndex("STR1")]

	if str2 != "" {
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
	} else {
		switch {
		case IsValidScheme(str1):
			raddr.scheme = str1
		case IsValidPort(str1):
			raddr.port = str1
		default:
			raddr.hostname = str1
		}
	}
	return raddr, nil
}
