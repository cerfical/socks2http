package proxy

import (
	"errors"
	"slices"
	"strings"
)

const (
	ProtoDirect Proto = iota + 1

	ProtoSOCKS
	ProtoSOCKS4
	ProtoSOCKS4a
	ProtoSOCKS5
	ProtoSOCKS5h

	ProtoHTTP
)

const (
	protoMin = ProtoDirect
	protoMax = ProtoHTTP
)

var protos = []string{
	ProtoDirect: "direct",

	ProtoSOCKS:   "socks",
	ProtoSOCKS4:  "socks4",
	ProtoSOCKS4a: "socks4a",
	ProtoSOCKS5:  "socks5",
	ProtoSOCKS5h: "socks5h",

	ProtoHTTP: "http",
}

type Proto int

func (p Proto) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Proto) UnmarshalText(text []byte) error {
	s := strings.ToLower(string(text))
	i := slices.Index(protos, s)
	if i == -1 {
		return errors.New("unknown protocol")
	}
	*p = Proto(i)
	return nil
}

func (p Proto) String() string {
	if p >= protoMin && p <= protoMax {
		return protos[p]
	}
	panic("unknown protocol")
}
