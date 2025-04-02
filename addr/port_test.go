package addr_test

import (
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePort(t *testing.T) {
	okTests := map[string]struct {
		input string
		want  int
	}{
		"parses minimum port number": {"0", 0},
		"parses maximum port number": {"65535", 65535},
	}
	for name, test := range okTests {
		t.Run(name, func(t *testing.T) {
			got, err := addr.ParsePort(test.input)
			require.NoError(t, err)

			assert.Equal(t, test.want, got)
		})
	}

	failTests := map[string]struct {
		input string
	}{
		"rejects empty input":               {""},
		"rejects out-of-range port numbers": {"65536"},
		"rejects negative port numbers":     {"-1"},
	}
	for name, test := range failTests {
		t.Run(name, func(t *testing.T) {
			_, err := addr.ParsePort(test.input)
			assert.Error(t, err)
		})
	}
}
