package socks4_test

import (
	"testing"

	"github.com/cerfical/socks2http/socks4"
	"github.com/stretchr/testify/assert"
)

func TestCommand_String(t *testing.T) {
	tests := map[string]struct {
		input socks4.Command
		want  string
	}{
		"prints valid command codes as command name followed by command code in hex": {
			input: socks4.CommandConnect,
			want:  "CONNECT (0x01)",
		},

		"prints invalid command codes as command code in hex": {
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
