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
		"prints supported statuses as description followed by status code in hex": {
			socks.Granted, "request granted (0x5a)",
		},

		"prints unsupported statuses as status code in hex": {
			0x17, "(0x17)",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.status.String()
			assert.Equal(t, test.want, got)
		})
	}
}
