package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/stretchr/testify/assert"
)

func TestAddr_UnmarshalText(t *testing.T) {
	httpAddr := &addr.Addr{Scheme: addr.HTTP, Hostname: "localhost", Port: 8080}
	tests := []struct {
		name  string
		input string
		want  *addr.Addr
	}{
		{"scheme_hostname_port_url", "http://localhost:8080", httpAddr},
		{"scheme_hostname_url", "http://localhost", httpAddr},
		{"scheme_port", "http:8080", httpAddr},
		{"scheme_hostname", "http:localhost", httpAddr},
		{"hostname_port", "localhost:8080", httpAddr},
		{"hostname", "localhost", httpAddr},
		{"upper_hostname", "LOCALHOST", httpAddr},
		{"port", "8080", httpAddr},
		{"http_scheme", "http", httpAddr},
		{"upper_scheme", "HTTP", httpAddr},
		{"socks4_scheme", "socks4", &addr.Addr{Scheme: addr.SOCKS4, Hostname: "localhost", Port: 1080}},
		{"direct_scheme", "direct", &addr.Addr{Scheme: addr.Direct, Hostname: "localhost"}},
		{"max_port", "65535", &addr.Addr{Scheme: addr.HTTP, Hostname: "localhost", Port: 65535}},
		{"min_port", "0", &addr.Addr{Scheme: addr.HTTP, Hostname: "localhost"}},
		{"no_out_of_range_port", "http://localhost:65536", nil},
		{"no_invalid_scheme", "invalidscheme://localhost:8080", nil},
		{"no_invalid", "http:localhost:0", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := &addr.Addr{}
			err := got.UnmarshalText([]byte(test.input))

			if test.want != nil {
				if assert.NoErrorf(t, err, "") {
					assert.Equalf(t, test.want, got, "")
				}
			} else {
				assert.Errorf(t, err, "")
			}
		})
	}
}

func TestAddr_MarshalText(t *testing.T) {
	tests := []struct {
		name  string
		input *addr.Addr
		want  string
	}{
		{"zero_value", &addr.Addr{}, "0"},
		{"port", &addr.Addr{Port: 80}, "80"},
		{"hostname", &addr.Addr{Hostname: "localhost"}, "localhost:0"},
		{"hostname_port", &addr.Addr{Hostname: "localhost", Port: 80}, "localhost:80"},
		{"scheme", &addr.Addr{Scheme: "http"}, "http:0"},
		{"scheme_port", &addr.Addr{Scheme: "http", Port: 80}, "http:80"},
		{"scheme_hostname", &addr.Addr{Scheme: "http", Hostname: "localhost"}, "http://localhost:0"},
		{"scheme_hostname_port", &addr.Addr{Scheme: "http", Hostname: "localhost", Port: 80}, "http://localhost:80"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := test.input.MarshalText()
			if assert.NoErrorf(t, err, "") {
				assert.Equalf(t, []byte(test.want), got, "")
			}
		})
	}
}

func TestAddr_Host(t *testing.T) {
	tests := []struct {
		name  string
		input *addr.Addr
		want  string
	}{
		{"zero_value", &addr.Addr{}, ":0"},
		{"port", &addr.Addr{Port: 80}, ":80"},
		{"hostname", &addr.Addr{Hostname: "localhost"}, "localhost:0"},
		{"hostname_port", &addr.Addr{Hostname: "localhost", Port: 80}, "localhost:80"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.Host()
			assert.Equalf(t, test.want, got, "")
		})
	}
}

func TestAddr_String(t *testing.T) {
	tests := []struct {
		name  string
		input *addr.Addr
		want  string
	}{
		{"zero_value", &addr.Addr{}, "://:0"},
		{"port", &addr.Addr{Port: 80}, "://:80"},
		{"hostname", &addr.Addr{Hostname: "localhost"}, "://localhost:0"},
		{"hostname_port", &addr.Addr{Hostname: "localhost", Port: 80}, "://localhost:80"},
		{"scheme", &addr.Addr{Scheme: "http"}, "http://:0"},
		{"scheme_port", &addr.Addr{Scheme: "http", Port: 80}, "http://:80"},
		{"scheme_hostname", &addr.Addr{Scheme: "http", Hostname: "localhost"}, "http://localhost:0"},
		{"scheme_hostname_port", &addr.Addr{Scheme: "http", Hostname: "localhost", Port: 80}, "http://localhost:80"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.input.String()
			assert.Equalf(t, test.want, got, "")
		})
	}
}
