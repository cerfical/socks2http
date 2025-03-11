package config_test

import (
	"fmt"
	"testing"

	"github.com/cerfical/socks2http/config"
	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := map[string]struct {
		arg  string
		want func(*config.Config) any
	}{
		"serve": {
			arg: "http://localhost:8090",
			want: func(c *config.Config) any {
				return c.ServeAddr.String()
			},
		},

		"proxy": {
			arg: "http://localhost:8090",
			want: func(c *config.Config) any {
				return c.ProxyAddr.String()
			},
		},

		"timeout": {
			arg: "12s",
			want: func(c *config.Config) any {
				return c.Timeout.String()
			},
		},

		"log": {
			arg: "info",
			want: func(c *config.Config) any {
				return c.LogLevel.String()
			},
		},
	}

	for name, test := range tests {
		t.Run(fmt.Sprintf("supports %s flag", name), func(t *testing.T) {
			config := config.Load([]string{"", fmt.Sprintf("--%s", name), test.arg})
			assert.Equal(t, test.arg, test.want(config))
		})
	}
}
