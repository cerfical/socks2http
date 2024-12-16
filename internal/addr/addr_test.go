package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	httpAddr := addr.New(addr.HTTP, "", 0)
	tests := []struct {
		input string
		want  *addr.Addr
	}{
		{"http://localhost:8080", httpAddr},
		{"http://localhost", httpAddr},
		{"http:8080", httpAddr},
		{"http:localhost", httpAddr},
		{"localhost:8080", httpAddr},
		{"localhost", httpAddr},
		{"LOCALHOST", httpAddr},
		{"8080", httpAddr},
		{"http", httpAddr},
		{"HTTP", httpAddr},
		{"socks4", addr.New(addr.SOCKS4, "", 0)},
		{"direct", addr.New(addr.Direct, "", 0)},
		{"http://localhost:65535", addr.New(addr.HTTP, "", 65535)},
		{"http://localhost:65536", nil},
		{"invalidscheme://localhost:8080", nil},
		{"http:localhost:8080", nil},
	}

	for _, test := range tests {
		gotAddr, gotErr := addr.Parse(test.input)
		if test.want != nil {
			assert.Equalf(t, test.want, gotAddr, "Want %q to be parsed into %v", test.input, test.want)
			assert.NoErrorf(t, gotErr, "Want parsing of %q to not fail", test.input)
		} else {
			assert.Errorf(t, gotErr, "Want parsing of %q to fail", test.input)
		}
	}
}
