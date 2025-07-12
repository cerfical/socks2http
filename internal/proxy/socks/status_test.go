package socks_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/socks"
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
		sts  socks.Status
		want string
	}{
		"prints valid statuses as text": {
			sts:  socks.StatusGeneralFailure,
			want: "General Failure",
		},

		"prints invalid statuses as numeric code": {
			sts:  0x17,
			want: "0x17",
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			t.Equal(test.want, test.sts.String())
		})
	}
}
