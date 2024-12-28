package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/stretchr/testify/assert"
)

func TestLookupIPv4(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      addr.IPv4Addr
		assertErr assert.ErrorAssertionFunc
	}{
		{"localhost", "localhost", addr.IPv4Addr{127, 0, 0, 1}, assert.NoError},
		{"empty_hostname", "", addr.IPv4Addr{127, 0, 0, 1}, assert.NoError},
		{"ipv4_addr", "1.1.1.1", addr.IPv4Addr{1, 1, 1, 1}, assert.NoError},
		{"no_ipv6_addr", "[0::0]", addr.IPv4Addr{}, assert.Error},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addr.LookupIPv4(tt.input)

			assert.Equal(t, tt.want, got)
			tt.assertErr(t, err)
		})
	}
}

func TestIPv4Addr_String(t *testing.T) {
	want := "1.2.3.4"
	input := addr.IPv4Addr{1, 2, 3, 4}

	assert.Equal(t, want, input.String())
}
