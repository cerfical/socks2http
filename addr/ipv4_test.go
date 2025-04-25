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
