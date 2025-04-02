package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseHost(t *testing.T) {
	okTests := map[string]struct {
		input string
		want  *addr.Host
	}{
		"parses hostname-port":                  {"localhost:80", addr.NewHost("localhost", 80)},
		"parses only port if hostname is empty": {":80", addr.NewHost("", 80)},
	}

	for name, test := range okTests {
		t.Run(name, func(t *testing.T) {
			h, err := addr.ParseHost(test.input)
			require.NoError(t, err)

			assert.Equal(t, test.want, h)
		})
	}

	failTests := map[string]struct {
		input string
	}{
		"rejects empty input": {""},
		"rejects_empty_port":  {"localhost:"},
	}

	for name, test := range failTests {
		t.Run(name, func(t *testing.T) {
			_, err := addr.ParseHost(test.input)
			assert.Error(t, err)
		})
	}
}

func TestHost_String(t *testing.T) {
	tests := map[string]struct {
		input *addr.Host
		want  string
	}{
		"prints zero value as zero port":        {addr.NewHost("", 0), ":0"},
		"prints only port if hostname is empty": {addr.NewHost("", 80), ":80"},
		"prints hostname-port if non-zero":      {addr.NewHost("localhost", 80), "localhost:80"},
		"prints IPv4-address-port if non-zero":  {addr.NewHost("127.0.0.1", 80), "127.0.0.1:80"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.input.String()
			assert.Equal(t, test.want, got)
		})
	}
}
