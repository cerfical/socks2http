package socks5_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/socks5"
	"github.com/stretchr/testify/suite"
)

func TestAuthMethod(t *testing.T) {
	suite.Run(t, new(AuthMethodTest))
}

type AuthMethodTest struct {
	suite.Suite
}

func (t *AuthMethodTest) TestString() {
	tests := map[string]struct {
		input socks5.AuthMethod
		want  string
	}{
		"prints valid auth methods as method name": {
			input: socks5.AuthNone,
			want:  "None",
		},

		"prints invalid auth methods as numeric code in hex": {
			input: 0x17,
			want:  "0x17",
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			got := test.input.String()
			t.Equal(test.want, got)
		})
	}
}
