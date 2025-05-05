package socks5_test

import (
	"testing"

	"github.com/cerfical/socks2http/proxy/socks5"
	"github.com/stretchr/testify/suite"
)

func TestCommand(t *testing.T) {
	suite.Run(t, new(CommandTest))
}

type CommandTest struct {
	suite.Suite
}

func (t *CommandTest) TestString() {
	tests := map[string]struct {
		input socks5.Command
		want  string
	}{
		"prints valid commands as command name followed by command code in hex": {
			input: socks5.CommandConnect,
			want:  "CONNECT (0x01)",
		},

		"prints invalid commands as error message followed by command code in hex": {
			input: 0x17,
			want:  "Invalid Command (0x17)",
		},
	}
	for name, test := range tests {
		t.Run(name, func() {
			got := test.input.String()
			t.Equal(test.want, got)
		})
	}
}
