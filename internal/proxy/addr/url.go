package addr

import (
	"cmp"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	urlRgx = regexp.MustCompile(fmt.Sprintf(
		`\A(?:%[1]v:(?://%[2]v)?(?::%[3]v)?|(?:%[2]v)?(?::%[3]v)?)\z`,
		`(?<SCHEME>[^:]+)`,   // Protocol scheme.
		`(?<HOSTNAME>[^:]+)`, // Hostname.
		`(?<PORT>[^:]+)`,     // Port number.
	))
	urlDefProto = ProtoHTTP
)

func NewURL(proto Proto, host string, port uint16) *URL {
	return &URL{
		Proto: proto,
		Host:  host,
		Port:  port,
	}
}

func ParseURL(url string, defProto Proto) (*URL, error) {
	if url == "" {
		return NewURL(0, "", 0), nil
	}

	rawURL, err := parseRawURL(url)
	if err != nil {
		return nil, err
	}

	scheme := strings.ToLower(cmp.Or(rawURL.Scheme, defProto.String()))
	host := strings.ToLower(rawURL.Host)

	proto, err := ParseProto(scheme)
	if err != nil {
		return nil, fmt.Errorf("parse scheme '%v': %w", scheme, err)
	}

	port := defaultPortForProto(proto)
	if rawURL.Port != "" {
		p, err := ParsePort(rawURL.Port)
		if err != nil {
			return nil, fmt.Errorf("parse port '%v': %w", rawURL.Port, err)
		}
		port = p
	}

	return NewURL(proto, host, port), nil
}

func parseRawURL(url string) (*rawURL, error) {
	matches := urlRgx.FindStringSubmatch(url)
	if matches == nil {
		return nil, errors.New("invalid syntax")
	}

	// Group named captures by name.
	submatches := make(map[string]string)
	for i, n := range urlRgx.SubexpNames() {
		submatches[n] = cmp.Or(submatches[n], matches[i])
	}

	return &rawURL{
		Scheme: submatches["SCHEME"],
		Host:   submatches["HOSTNAME"],
		Port:   submatches["PORT"],
	}, nil
}

func defaultPortForProto(p Proto) uint16 {
	switch p {
	case ProtoSOCKS, ProtoSOCKS4, ProtoSOCKS4a, ProtoSOCKS5, ProtoSOCKS5h:
		return 1080
	case ProtoHTTP:
		return 80
	default:
		return 0
	}
}

type URL struct {
	Proto Proto

	Host string
	Port uint16
}

func (u *URL) Addr() *Addr {
	return NewAddr(u.Host, u.Port)
}

func (u *URL) IsZero() bool {
	return u.Proto == 0 && u.Host == "" && u.Port == 0
}

func (u *URL) String() string {
	if u.IsZero() {
		return ""
	}

	scheme := fmt.Sprintf("%v:", u.Proto)
	host := u.Host
	port := fmt.Sprintf(":%v", u.Port)

	if host != "" {
		host = fmt.Sprintf("//%v", host)
	}

	return strings.ToLower(scheme) + strings.ToLower(host) + port
}

func (u *URL) MarshalText() ([]byte, error) {
	return []byte(u.String()), nil
}

func (u *URL) UnmarshalText(text []byte) error {
	url, err := ParseURL(string(text), urlDefProto)
	if err != nil {
		return err
	}
	*u = *url
	return nil
}

type rawURL struct {
	Scheme string
	Host   string
	Port   string
}
