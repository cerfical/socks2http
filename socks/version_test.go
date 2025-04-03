package socks_test

import (
	"testing"

	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/assert"
)

func TestVersion_String(t *testing.T) {
	tests := map[string]struct {
		ver  socks.Version
		want string
	}{
		"prints supported versions as version name followed by version code in hex": {socks.V4, "SOCKS4 (0x04)"},
		"prints unsupported versions as version code in hex":                        {0x17, "(0x17)"},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.ver.String()
			assert.Equal(t, test.want, got)
		})
	}
}
