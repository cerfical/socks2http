package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/addr"
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
			want:  addr.New("localhost", 80),
		},

		"parses an empty host": {
			input: ":80",
			want:  addr.New("", 80),
		},

		"parses an empty input to a default value": {
			input: "",
			want:  &addr.Addr{},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			addr, err := addr.Parse(test.input)
			t.Require().NoError(err)

			t.Equal(test.want, addr)
		})
	}

	t.Run("rejects an empty port", func() {
		_, err := addr.Parse("localhost:")
		t.Error(err)
	})
}

func (t *AddrTest) TestString() {
	tests := map[string]struct {
		addr *addr.Addr
		want string
	}{
		"prints a host-port pair": {
			addr: addr.New("localhost", 80),
			want: "localhost:80",
		},

		"prints an empty host": {
			addr: addr.New("", 80),
			want: ":80",
		},

		"prints a default value as an empty string": {
			addr: &addr.Addr{},
			want: "",
		},

		"prints an IPv4-address": {
			addr: addr.New("127.0.0.1", 80),
			want: "127.0.0.1:80",
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			got := test.addr.String()
			t.Equal(test.want, got)
		})
	}
}

func (t *AddrTest) TestResolveToIPv4() {
	tests := map[string]struct {
		addr *addr.Addr
		want addr.IPv4
	}{
		"resolves localhost to 127-0-0-1": {
			addr: addr.New("localhost", 0),
			want: addr.IPv4{127, 0, 0, 1},
		},

		"resolves an empty host to 127-0-0-1": {
			addr: addr.New("", 0),
			want: addr.IPv4{127, 0, 0, 1},
		},

		"resolves an IPv4 address to itself": {
			addr: addr.New("1.1.1.1", 0),
			want: addr.IPv4{1, 1, 1, 1},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			got, err := test.addr.ResolveToIPv4()
			t.Require().NoError(err)

			t.Equal(test.want, got)
		})
	}

	t.Run("rejects IPv6 addresses", func() {
		a := addr.New("[0::0]", 0)

		_, err := a.ResolveToIPv4()
		t.Require().Error(err)
	})
}

func (t *AddrTest) TestToIPv4() {
	t.Run("parses an IPv4 host", func() {
		a := addr.New("127.0.0.1", 0)

		ip, ok := a.ToIPv4()
		t.Require().True(ok)

		want := addr.IPv4{127, 0, 0, 1}
		t.Equal(want, ip)
	})

	t.Run("rejects named hosts", func() {
		a := addr.New("localhost", 0)

		_, ok := a.ToIPv4()
		t.Require().False(ok)
	})

	t.Run("rejects IPv6 addresses", func() {
		a := addr.New("0:0:0:0:0:0:0:1", 0)

		_, ok := a.ToIPv4()
		t.Require().False(ok)
	})
}
