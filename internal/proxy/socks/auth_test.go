package socks_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/socks"
	"github.com/stretchr/testify/suite"
)

func TestAuth(t *testing.T) {
	suite.Run(t, new(AuthTest))
}

type AuthTest struct {
	suite.Suite
}

func (t *AuthTest) TestString() {
	tests := map[string]struct {
		auth socks.Auth
		want string
	}{
		"prints valid auth methods as text": {
			auth: socks.AuthNone,
			want: "None",
		},

		"prints invalid auth methods as numeric code": {
			auth: 0x17,
			want: "0x17",
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			t.Equal(test.want, test.auth.String())
		})
	}
}
