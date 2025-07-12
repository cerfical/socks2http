package socks_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/socks"
	"github.com/stretchr/testify/suite"
)

func TestVersion(t *testing.T) {
	suite.Run(t, new(VersionTest))
}

type VersionTest struct {
	suite.Suite
}

func (t *VersionTest) TestString() {
	tests := map[string]struct {
		ver  socks.Version
		want string
	}{
		"prints valid versions as text": {
			ver:  socks.V4,
			want: "SOCKS4",
		},

		"prints invalid versions as numeric code": {
			ver:  0x17,
			want: "0x17",
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			t.Equal(test.want, test.ver.String())
		})
	}
}
