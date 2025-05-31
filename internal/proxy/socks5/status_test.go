package socks5_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/socks5"
	"github.com/stretchr/testify/suite"
)

func TestStatus(t *testing.T) {
	suite.Run(t, new(StatusTest))
}

type StatusTest struct {
	suite.Suite
}

func (t *StatusTest) TestString() {
	tests := map[string]struct {
		input socks5.Status
		want  string
	}{
		"prints valid statuses as short description": {
			input: socks5.StatusGeneralFailure,
			want:  "General Failure",
		},

		"prints invalid statuses as numeric code in hex": {
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
