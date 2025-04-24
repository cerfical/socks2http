package socks5_test

import (
	"testing"

	"github.com/cerfical/socks2http/socks5"
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
		"prints valid statuses as short description followed by status code in hex": {
			input: socks5.StatusGeneralFailure,
			want:  "General Failure (0x01)",
		},

		"prints invalid statuses as error message followed by status code in hex": {
			input: 0x17,
			want:  "Invalid Status (0x17)",
		},
	}
	for name, test := range tests {
		t.Run(name, func() {
			got := test.input.String()
			t.Equal(test.want, got)
		})
	}
}
