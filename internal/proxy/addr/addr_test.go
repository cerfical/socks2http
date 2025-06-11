package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/stretchr/testify/suite"
)

func TestAddr(t *testing.T) {
	suite.Run(t, new(AddrTest))
}

type AddrTest struct {
	suite.Suite
}

func (t *AddrTest) TestParse() {
	tests := map[string]struct {
		input string
		want  *addr.Addr
	}{
		"parses a host-port pair": {
			input: "localhost:80",
			want:  addr.NewAddr("localhost", 80),
		},

		"parses an empty host": {
			input: ":80",
			want:  addr.NewAddr("", 80),
		},

		"parses an IPv4 address": {
			input: "127.0.0.1:80",
			want:  addr.NewAddr("127.0.0.1", 80),
		},

		"parses an IPv6 address": {
			input: "[1::1]:80",
			want:  addr.NewAddr("1::1", 80),
		},

		"parses an empty input to a default value": {
			input: "",
			want:  &addr.Addr{},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			addr, err := addr.ParseAddr(test.input)
			t.Require().NoError(err)

			t.Equal(test.want, addr)
		})
	}

	t.Run("rejects an empty port", func() {
		_, err := addr.ParseAddr("localhost:")
		t.Error(err)
	})
}

func (t *AddrTest) TestString() {
	tests := map[string]struct {
		addr *addr.Addr
		want string
	}{
		"prints a host-port pair": {
			addr: addr.NewAddr("localhost", 80),
			want: "localhost:80",
		},

		"prints an empty host": {
			addr: addr.NewAddr("", 80),
			want: ":80",
		},

		"prints a default value as an empty string": {
			addr: &addr.Addr{},
			want: "",
		},

		"prints an IPv4-address": {
			addr: addr.NewAddr("127.0.0.1", 80),
			want: "127.0.0.1:80",
		},

		"prints an IPv6-address": {
			addr: addr.NewAddr("1::1", 80),
			want: "[1::1]:80",
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			got := test.addr.String()
			t.Equal(test.want, got)
		})
	}
}
