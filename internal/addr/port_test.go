package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/stretchr/testify/assert"
)

func TestIsValidPort(t *testing.T) {
	tests := []struct {
		input  string
		wantOk bool
	}{
		{"0", true},
		{"65535", true},
		{"65536", false},
		{"-1", false},
		{"1.0", false},
		{"txt", false},
		{"", false},
	}

	for _, test := range tests {
		gotOk := addr.IsValidPort(test.input)
		if test.wantOk {
			assert.Truef(t, gotOk, "want %q to be considered a valid port number", test.input)
		} else {
			assert.Falsef(t, gotOk, "want %q to be considered an invalid port number", test.input)
		}
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		input   string
		wantNum uint16
		wantOk  bool
	}{
		{"0", 0, true},
		{"65535", 65535, true},
		{"65536", 0, false},
		{"-1", 0, false},
		{"1.0", 0, false},
		{"txt", 1, false},
		{"", 0, false},
	}

	for _, test := range tests {
		gotNum, gotErr := addr.ParsePort(test.input)
		if test.wantOk {
			assert.Equalf(t, test.wantNum, gotNum, "want %q to be parsed to %v", test.input, test.wantNum)
			assert.NoErrorf(t, gotErr, "want parsing of %q to not fail", test.input)
		} else {
			assert.Errorf(t, gotErr, "want parsing of %q to fail", test.input)
		}
	}
}