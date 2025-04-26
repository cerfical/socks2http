package socks4_test

import (
	"testing"

	"github.com/cerfical/socks2http/socks4"
	"github.com/stretchr/testify/assert"
)

func TestVersion_String(t *testing.T) {
	tests := map[string]struct {
		input socks4.Version
		want  string
	}{
		"prints valid version codes as version name followed by version code in hex": {
			input: socks4.V4,
			want:  "SOCKS4 (0x04)",
		},

		"prints invalid version codes as version code in hex": {
			input: 0x17,
			want:  "(0x17)",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.input.String()
			assert.Equal(t, test.want, got)
		})
	}
}
