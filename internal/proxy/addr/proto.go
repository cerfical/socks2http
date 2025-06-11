package addr

import (
	"errors"
	"slices"
	"strings"
)

const (
	ProtoSOCKS Proto = iota + 1
	ProtoSOCKS4
	ProtoSOCKS4a
	ProtoSOCKS5
	ProtoSOCKS5h

	ProtoHTTP
)

const (
	protoMin = ProtoSOCKS
	protoMax = ProtoHTTP
)

var protos = []string{
	ProtoSOCKS:   "SOCKS",
	ProtoSOCKS4:  "SOCKS4",
	ProtoSOCKS4a: "SOCKS4a",
	ProtoSOCKS5:  "SOCKS5",
	ProtoSOCKS5h: "SOCKS5h",

	ProtoHTTP: "HTTP",
}

func ParseProto(proto string) (Proto, error) {
	i := slices.IndexFunc(protos, func(s string) bool {
		return strings.EqualFold(s, proto)
	})
	if i == -1 {
		return 0, errors.New("unknown protocol")
	}
	return Proto(i), nil
}

type Proto int

func (p Proto) String() string {
	if p >= protoMin && p <= protoMax {
		return protos[p]
	}
	panic("unknown protocol")
}

func (p Proto) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

func (p *Proto) UnmarshalText(text []byte) error {
	proto, err := ParseProto(string(text))
	if err != nil {
		return err
	}
	*p = proto
	return nil
}
