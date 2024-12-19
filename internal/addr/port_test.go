package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/stretchr/testify/assert"
)

func TestParsePort(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  uint16
		ok    bool
	}{
		{"min", "0", 0, true},
		{"max", "65535", 65535, true},
		{"no_empty", "", 0, false},
		{"no_out_of_range", "65536", 0, false},
		{"no_negative", "-1", 0, false},
		{"no_float", "1.0", 0, false},
		{"no_letters", "txt", 1, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := addr.ParsePort(test.input)
			if test.ok {
				if assert.NoErrorf(t, err, "") {
					assert.Equalf(t, test.want, got, "")
				}
			} else {
				assert.Errorf(t, err, "")
			}
		})
	}
}
