package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/stretchr/testify/assert"
)

func TestIPv4_String(t *testing.T) {
	t.Run("prints the address in dot-decimal notation", func(t *testing.T) {
		want := "1.2.3.4"
		ip4 := addr.IPv4{1, 2, 3, 4}

		assert.Equal(t, want, ip4.String())
	})
}

func TestIPv4_IsEmpty(t *testing.T) {
	t.Run("designates 0.0.0.0 as an empty address", func(t *testing.T) {
		ip4 := addr.IPv4{0, 0, 0, 0}

		assert.True(t, ip4.IsEmpty())
	})

	t.Run("designates addresses other than 0.0.0.0 as non-empty", func(t *testing.T) {
		ip4 := addr.IPv4{127, 0, 0, 1}

		assert.False(t, ip4.IsEmpty())
	})
}
