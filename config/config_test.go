package config_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/config"
	"github.com/cerfical/socks2http/log"
	"github.com/stretchr/testify/suite"
)

func TestConfig(t *testing.T) {
	suite.Run(t, new(ConfigTest))
}

type ConfigTest struct {
	suite.Suite
}

func (t *ConfigTest) TestLoad() {
	flagTests := map[string]struct {
		arg  string
		want func(*config.Config)
	}{
		"serve": {
			arg: "http://localhost:8090",
			want: func(c *config.Config) {
				want := addr.New(addr.HTTP, "localhost", 8090)
				t.Equal(want, &c.ServeAddr)
			},
		},

		"proxy": {
			arg: "http://localhost:8090",
			want: func(c *config.Config) {
				want := addr.New(addr.HTTP, "localhost", 8090)
				t.Equal(want, &c.ProxyAddr)
			},
		},

		"timeout": {
			arg: "12s",
			want: func(c *config.Config) {
				t.Equal(time.Second*12, c.Timeout)
			},
		},

		"log": {
			arg: "info",
			want: func(c *config.Config) {
				t.Equal(log.LevelInfo, c.LogLevel)
			},
		},
	}

	for flagName, test := range flagTests {
		t.Run(fmt.Sprintf("supports %s flag", flagName), func() {
			config := config.Load([]string{"", fmt.Sprintf("--%s", flagName), test.arg})
			test.want(config)
		})
	}
}
