package addr

import (
	"regexp"
	"strings"
)

type rawAddr struct {
	scheme   string
	hostname string
	port     string
}

var rgx = regexp.MustCompile(`\A((?<SCHEME>[^:]+)://(?<HOSTNAME>[^:]+)(:(?<PORT>[^:]+))?|(?<STR1>[^:]+)(:(?<STR2>[^:]+))?)\z`)

func parseRaw(addr string) (*rawAddr, bool) {
	matches := rgx.FindStringSubmatch(addr)
	if matches == nil {
		return nil, false
	}

	raddr := &rawAddr{
		scheme:   strings.ToLower(matches[rgx.SubexpIndex("SCHEME")]),
		hostname: strings.ToLower(matches[rgx.SubexpIndex("HOSTNAME")]),
		port:     strings.ToLower(matches[rgx.SubexpIndex("PORT")]),
	}

	// if the address is a regular URL
	if raddr.scheme != "" {
		return raddr, true
	}

	str2 := strings.ToLower(matches[rgx.SubexpIndex("STR2")])
	str1 := strings.ToLower(matches[rgx.SubexpIndex("STR1")])

	if str2 != "" {
		if isValidScheme(str1) {
			raddr.scheme = str1
			if isValidPort(str2) {
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
		case isValidScheme(str1):
			raddr.scheme = str1
		case isValidPort(str1):
			raddr.port = str1
		default:
			raddr.hostname = str1
		}
	}
	return raddr, true
}
