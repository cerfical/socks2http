package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/stretchr/testify/suite"
)

func TestURL(t *testing.T) {
	suite.Run(t, new(URLTest))
}

type URLTest struct {
	suite.Suite
}

func (t *URLTest) TestParse() {
	tests := map[string]struct {
		input string
		want  func(*addr.URL)
	}{
		"parses empty input to empty URL": {
			input: "",
			want: func(u *addr.URL) {
				t.Zero(*u)
			},
		},

		"parses URL with only host specified": {
			input: "example.com",
			want: func(u *addr.URL) {
				t.Equal("example.com", u.Host)
			},
		},

		"parses URL with only port number specified": {
			input: ":1080",
			want: func(u *addr.URL) {
				t.Equal(uint16(1080), u.Port)
			},
		},

		"parses URL with only protocol scheme specified": {
			input: "socks5:",
			want: func(u *addr.URL) {
				t.Equal(addr.ProtoSOCKS5, u.Proto)
			},
		},

		"parses host-port pairs": {
			input: "example.com:80",
			want: func(u *addr.URL) {
				t.Equal("example.com", u.Host)
				t.Equal(uint16(80), u.Port)
			},
		},

		"parses scheme-port pairs": {
			input: "http::8080",
			want: func(u *addr.URL) {
				t.Equal(addr.ProtoHTTP, u.Proto)
				t.Equal(uint16(8080), u.Port)
			},
		},

		"parses scheme-host pairs": {
			input: "socks4://example.com",
			want: func(u *addr.URL) {
				t.Equal(addr.ProtoSOCKS4, u.Proto)
				t.Equal("example.com", u.Host)
			},
		},

		"parses URL with scheme, host and port number": {
			input: "socks4://example.com:1081",
			want: func(u *addr.URL) {
				t.Equal(addr.ProtoSOCKS4, u.Proto)
				t.Equal("example.com", u.Host)
				t.Equal(uint16(1081), u.Port)
			},
		},

		"derives port number from protocol scheme if not specified": {
			input: "http://example.com",
			want: func(u *addr.URL) {
				t.Equal(uint16(80), u.Port)
			},
		},

		"uses HTTP as default protocol scheme": {
			input: "example.com",
			want: func(u *addr.URL) {
				t.Equal(addr.ProtoHTTP, u.Proto)
			},
		},

		"ignores case when parsing protocol scheme": {
			input: "HTTP://example.com",
			want: func(u *addr.URL) {
				t.Equal(addr.ProtoHTTP, u.Proto)
			},
		},

		"ignores case when parsing host": {
			input: "EXAMPLE.COM",
			want: func(u *addr.URL) {
				t.Equal("example.com", u.Host)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			url, err := addr.ParseURL(test.input)
			t.Require().NoError(err)

			test.want(url)
		})
	}

	errors := map[string]string{
		"rejects malformed URL":            "http:example.com:80",
		"rejects invalid port number":      "example.com:abc",
		"rejects out-of-range port number": "example.com:70000",
		"rejects invalid protocol scheme":  "badproto://example.com",
	}

	for name, test := range errors {
		t.Run(name, func() {
			_, err := addr.ParseURL(test)
			t.Require().Error(err)
		})
	}
}

func (t *URLTest) TestString() {
	tests := map[string]struct {
		url  *addr.URL
		want string
	}{
		"prints empty URL as empty string": {
			url:  addr.NewURL(0, "", 0),
			want: "",
		},

		"prints URL with only host specified": {
			url:  addr.NewURL(0, "example.com", 0),
			want: "example.com",
		},

		"prints URL with only protocol scheme specified": {
			url:  addr.NewURL(addr.ProtoSOCKS4, "", 0),
			want: "socks4:",
		},

		"prints URL with only port number specified": {
			url:  addr.NewURL(0, "", 80),
			want: ":80",
		},

		"prints scheme-port pairs": {
			url:  addr.NewURL(addr.ProtoSOCKS4, "", 80),
			want: "socks4::80",
		},

		"prints host-port pairs": {
			url:  addr.NewURL(0, "example.com", 80),
			want: "example.com:80",
		},

		"prints scheme-host pairs": {
			url:  addr.NewURL(addr.ProtoSOCKS4, "example.com", 0),
			want: "socks4://example.com",
		},

		"prints URL with scheme, host and port number": {
			url:  addr.NewURL(addr.ProtoSOCKS4, "example.com", 1081),
			want: "socks4://example.com:1081",
		},

		"ignores case when printing host": {
			url:  addr.NewURL(0, "EXAMPLE.COM", 0),
			want: "example.com",
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			t.Equal(test.want, test.url.String())
		})
	}
}

func (t *URLTest) TestTextMarshalUnmarshal() {
	t.Run("marshalling converts URL to string", func() {
		url := addr.NewURL(addr.ProtoSOCKS4, "example.com", 1081)

		data, err := url.MarshalText()
		t.Require().NoError(err)

		t.Equal("socks4://example.com:1081", string(data))
	})

	t.Run("unmarshalling parses URL from string", func() {
		input := "socks4://example.com:1081"

		var url addr.URL
		err := url.UnmarshalText([]byte(input))
		t.Require().NoError(err)

		want := addr.NewURL(addr.ProtoSOCKS4, "example.com", 1081)
		t.Equal(want, &url)
	})
}

func (t *URLTest) TestAddr() {
	t.Run("returns host-port part of URL", func() {
		url := addr.NewURL(addr.ProtoSOCKS4, "example.com", 1081)

		want := addr.New("example.com", 1081)
		t.Equal(want, url.Addr())
	})
}
