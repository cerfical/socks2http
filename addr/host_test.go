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
		host *addr.Host
		want string
	}{
		"prints zero value as zero port":        {addr.NewHost("", 0), ":0"},
		"prints only port if hostname is empty": {addr.NewHost("", 80), ":80"},
		"prints hostname-port if non-zero":      {addr.NewHost("localhost", 80), "localhost:80"},
		"prints IPv4-address-port if non-zero":  {addr.NewHost("127.0.0.1", 80), "127.0.0.1:80"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.host.String()
			assert.Equal(t, test.want, got)
		})
	}
}

func TestHost_ResolveToIPv4(t *testing.T) {
	okTests := map[string]struct {
		host *addr.Host
		want addr.IPv4
	}{
		"resolves localhost to 127-0-0-1":         {addr.NewHost("localhost", 0), addr.IPv4{127, 0, 0, 1}},
		"resolves an empty hostname to 127-0-0-1": {addr.NewHost("", 0), addr.IPv4{127, 0, 0, 1}},
		"resolves an IPv4 address to itself":      {addr.NewHost("1.1.1.1", 0), addr.IPv4{1, 1, 1, 1}},
	}
	for name, test := range okTests {
		t.Run(name, func(t *testing.T) {
			got, err := test.host.ResolveToIPv4()
			require.NoError(t, err)

			assert.Equal(t, test.want, got)
		})
	}

	failTests := map[string]struct {
		host *addr.Host
	}{
		"rejects IPv6 addresses": {addr.NewHost("[0::0]", 0)},
	}
	for name, test := range failTests {
		t.Run(name, func(t *testing.T) {
			_, err := test.host.ResolveToIPv4()
			require.Error(t, err)
		})
	}
}

func TestHost_ToIPv4(t *testing.T) {
	tests := map[string]struct {
		hostname string
		want     addr.IPv4
		fail     bool
	}{
		"parses IPv4 hostnames": {
			hostname: "127.0.0.1",
			want:     addr.IPv4{127, 0, 0, 1},
		},

		"rejects symbolic hostnames": {
			hostname: "localhost",
			fail:     true,
		},

		"rejects IPv6 addresses": {
			hostname: "0:0:0:0:0:0:0:1",
			fail:     true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			host := addr.NewHost(test.hostname, 0)

			ip, ok := host.ToIPv4()
			if test.fail {
				require.False(t, ok)
			} else {
				require.True(t, ok)
				assert.Equal(t, test.want, ip)
			}
		})
	}
}
