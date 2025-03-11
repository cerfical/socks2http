package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/stretchr/testify/assert"
)

func TestAddr_UnmarshalText(t *testing.T) {
	httpAddr := &addr.Addr{Scheme: addr.HTTP, Hostname: "localhost", Port: 8080}
	tests := []struct {
		name  string
		input string
		want  *addr.Addr
	}{
		{"scheme_hostname_port", "http://localhost:8080", httpAddr},
		{"scheme_hostname", "http://localhost", httpAddr},
		{"scheme_port", "http::8080", httpAddr},
		{"hostname_port", "localhost:8080", httpAddr},
		{"hostname", "//localhost", httpAddr},
		{"upper_hostname", "//LOCALHOST", httpAddr},
		{"port", ":8080", httpAddr},
		{"http_scheme", "http", httpAddr},
		{"upper_scheme", "HTTP", httpAddr},
		{"socks4_scheme", "socks4:", &addr.Addr{Scheme: addr.SOCKS4, Hostname: "localhost", Port: 1080}},
		{"direct_scheme", "direct:", &addr.Addr{Scheme: addr.Direct}},
		{"max_port", ":65535", &addr.Addr{Scheme: addr.HTTP, Hostname: "localhost", Port: 65535}},
		{"min_port", ":0", &addr.Addr{Scheme: addr.HTTP, Hostname: "localhost", Port: 0}},
		{"no_out_of_range_port", ":65536", nil},
		{"no_malformed", "http:localhost:0", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &addr.Addr{}
			err := got.UnmarshalText([]byte(tt.input))

			if tt.want != nil {
				if assert.NoErrorf(t, err, "") {
					assert.Equalf(t, tt.want, got, "")
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
		{"zero_value", &addr.Addr{}, ""},
		{"port", &addr.Addr{Port: 80}, ":80"},
		{"hostname", &addr.Addr{Hostname: "localhost"}, "//localhost"},
		{"hostname_port", &addr.Addr{Hostname: "localhost", Port: 80}, "localhost:80"},
		{"scheme", &addr.Addr{Scheme: "http"}, "http"},
		{"scheme_port", &addr.Addr{Scheme: "http", Port: 80}, "http::80"},
		{"scheme_hostname", &addr.Addr{Scheme: "http", Hostname: "localhost"}, "http://localhost"},
		{"scheme_hostname_port", &addr.Addr{Scheme: "http", Hostname: "localhost", Port: 80}, "http://localhost:80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.MarshalText()
			if assert.NoErrorf(t, err, "") {
				assert.Equalf(t, tt.want, string(got), "")
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
		{"zero_value", &addr.Addr{}, ":"},
		{"port", &addr.Addr{Port: 80}, ":80"},
		{"hostname", &addr.Addr{Hostname: "localhost"}, "localhost:"},
		{"hostname_port", &addr.Addr{Hostname: "localhost", Port: 80}, "localhost:80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Host()
			assert.Equalf(t, tt.want, got, "")
		})
	}
}

func TestAddr_String(t *testing.T) {
	tests := []struct {
		name  string
		input *addr.Addr
		want  string
	}{
		{"zero_value", &addr.Addr{}, ""},
		{"port", &addr.Addr{Port: 80}, ":80"},
		{"hostname", &addr.Addr{Hostname: "localhost"}, "//localhost"},
		{"hostname_port", &addr.Addr{Hostname: "localhost", Port: 80}, "localhost:80"},
		{"scheme", &addr.Addr{Scheme: "http"}, "http"},
		{"scheme_port", &addr.Addr{Scheme: "http", Port: 80}, "http::80"},
		{"scheme_hostname", &addr.Addr{Scheme: "http", Hostname: "localhost"}, "http://localhost"},
		{"scheme_hostname_port", &addr.Addr{Scheme: "http", Hostname: "localhost", Port: 80}, "http://localhost:80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.String()
			assert.Equalf(t, tt.want, got, "")
		})
	}
}
