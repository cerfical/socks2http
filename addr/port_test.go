package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/addr"
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := addr.ParsePort(tt.input)
			if tt.ok {
				if assert.NoErrorf(t, err, "") {
					assert.Equalf(t, tt.want, got, "")
				}
			} else {
				assert.Errorf(t, err, "")
			}
		})
	}
}
