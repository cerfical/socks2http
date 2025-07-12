package socks_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/socks"
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
		cmd  socks.Command
		want string
	}{
		"prints valid commands as text": {
			cmd:  socks.CommandConnect,
			want: "CONNECT",
		},

		"prints invalid commands as numeric code": {
			cmd:  0x17,
			want: "0x17",
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			t.Equal(test.want, test.cmd.String())
		})
	}
}
