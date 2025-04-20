package socks_test

import (
	"testing"

	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/assert"
)

func TestStatus_String(t *testing.T) {
	tests := map[string]struct {
		status socks.Status
		want   string
	}{
		"prints valid reply codes as reply message followed by reply code in hex": {
			socks.Granted, "Request Granted (0x5a)",
		},

		"prints invalid reply codes as an empty string": {
			0x17, "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.status.String()
			assert.Equal(t, test.want, got)
		})
	}
}
