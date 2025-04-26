package socks4_test

import (
	"testing"

	"github.com/cerfical/socks2http/socks4"
	"github.com/stretchr/testify/assert"
)

func TestStatus_String(t *testing.T) {
	tests := map[string]struct {
		input socks4.Status
		want  string
	}{
		"prints valid statuses as short description followed by status code in hex": {
			socks4.StatusGranted, "Request Granted (0x5a)",
		},

		"prints invalid statuses as an empty string": {
			0x17, "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := test.input.String()
			assert.Equal(t, test.want, got)
		})
	}
}
