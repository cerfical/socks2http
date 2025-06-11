package config_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cerfical/socks2http/internal/config"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/proxy/addr"
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
		"server": {
			arg: "http://localhost:81",
			want: func(c *config.Config) {
				want := addr.NewURL(addr.ProtoHTTP, "localhost", 81)
				t.Equal(want, &c.Server)
			},
		},

		"proxy": {
			arg: "http://localhost:81",
			want: func(c *config.Config) {
				want := addr.NewURL(addr.ProtoHTTP, "localhost", 81)
				t.Equal(want, &c.Proxy)
			},
		},

		"timeout": {
			arg: "12s",
			want: func(c *config.Config) {
				t.Equal(time.Second*12, c.Timeout)
			},
		},

		"log-level": {
			arg: "info",
			want: func(c *config.Config) {
				t.Equal(log.LevelInfo, c.Log.Level)
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
