package socks_test

import (
	"testing"

	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/assert"
)

func TestCommand_String(t *testing.T) {
	tests := map[string]struct {
		cmd  socks.Command
		want string
	}{
		"prints valid command codes as command name followed by command code in hex": {
			socks.Connect, "CONNECT (0x01)",
		},

		"prints invalid command codes as an empty string": {
			0x17, "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.cmd.String()
			assert.Equal(t, test.want, got)
		})
	}
}
