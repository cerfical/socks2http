package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAddr(t *testing.T) {
	httpAddr := addr.New(addr.HTTP, "localhost", 8080)
	okTests := map[string]struct {
		input string
		want  *addr.Addr
	}{
		"parses scheme-hostname-port": {"http://localhost:8080", httpAddr},
		"parses scheme-hostname":      {"http://localhost", httpAddr},
		"parses scheme-port":          {"http::8080", httpAddr},
		"parses hostname-port":        {"localhost:8080", httpAddr},
		"parses hostname":             {"//localhost", httpAddr},
		"parses port":                 {":8080", httpAddr},

		"parses hostname case-insensitively": {"//LOCALHOST", httpAddr},
		"parses scheme case-insensitively":   {"HTTP", httpAddr},
	}
	for name, test := range okTests {
		t.Run(name, func(t *testing.T) {
			got, err := addr.Parse(test.input)
			require.NoError(t, err)

			assert.Equal(t, test.want, got)
		})
	}

	failTests := map[string]struct {
		input string
	}{
		"rejects malformed input": {"http:localhost:0"},
	}
	for name, test := range failTests {
		t.Run(name, func(t *testing.T) {
			_, err := addr.Parse(test.input)
			assert.Error(t, err)
		})
	}
}

func TestAddr_String(t *testing.T) {
	tests := map[string]struct {
		input *addr.Addr
		want  string
	}{
		"prints zero value as empty string":       {addr.New("", "", 0), ""},
		"prints scheme if non-zero":               {addr.New("http", "", 0), "http"},
		"prints hostname if non-zero":             {addr.New("", "localhost", 0), "//localhost"},
		"prints port if non-zero":                 {addr.New("", "", 80), ":80"},
		"prints hostname-port if non-zero":        {addr.New("", "localhost", 80), "localhost:80"},
		"prints scheme-port if non-zero":          {addr.New("http", "", 80), "http::80"},
		"prints scheme-hostname if non-zero":      {addr.New("http", "localhost", 0), "http://localhost"},
		"prints scheme-hostname-port if non-zero": {addr.New("http", "localhost", 80), "http://localhost:80"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.input.String()
			assert.Equal(t, test.want, got)
		})
	}
}
