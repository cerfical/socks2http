package proxcli_test

import (
	"context"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy/proxcli"
	"github.com/cerfical/socks2http/proxy/proxtest"
	"github.com/stretchr/testify/suite"
)

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTest))
}

type ClientTest struct {
	suite.Suite
}

func (t *ClientTest) TestNew() {
	tests := map[string]struct {
		options proxcli.Option
		want    func(*proxcli.Client)
		err     func(error)
	}{
		"uses http-localhost-8080 as the default proxy address": {
			want: func(c *proxcli.Client) {
				t.Equal(addr.New(addr.HTTP, "localhost", 8080), c.ProxyAddr())
			},
		},

		"uses a non-default proxy address if one is provided": {
			options: proxcli.WithProxyAddr(addr.New(addr.HTTP, "example.com", 8181)),
			want: func(c *proxcli.Client) {
				t.Equal(addr.New(addr.HTTP, "example.com", 8181), c.ProxyAddr())
			},
		},

		"rejects unsupported protocol schemes": {
			options: proxcli.WithProxyAddr(addr.New("SOCKS9", "", 0)),
			err: func(err error) {
				t.ErrorContains(err, "SOCKS9")
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			ops := []proxcli.Option{}
			if test.options != nil {
				ops = append(ops, test.options)
			}

			client, err := proxcli.New(ops...)
			if test.err != nil {
				test.err(err)
			} else {
				t.Require().NoError(err)
				test.want(client)
			}
		})
	}
}

func (t *ClientTest) TestDial() {
	tests := map[string]struct {
		setup func() proxcli.Option
	}{
		"establishes a direct connection to a server if Direct is used": {
			setup: func() proxcli.Option {
				return proxcli.WithProxyAddr(addr.New(addr.Direct, "", 0))
			},
		},

		"connects to a server via an HTTP proxy": {
			setup: func() proxcli.Option {
				httpProxyHost := proxtest.StartProxyServer(t.T(), addr.HTTP)
				return proxcli.WithProxyAddr(&addr.Addr{
					Scheme: addr.HTTP,
					Host:   *httpProxyHost,
				})
			},
		},

		"connects to a server via a SOCKS proxy": {
			setup: func() proxcli.Option {
				socks4ProxyHost := proxtest.StartProxyServer(t.T(), addr.SOCKS4)
				return proxcli.WithProxyAddr(&addr.Addr{
					Scheme: addr.SOCKS4,
					Host:   *socks4ProxyHost,
				})
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			client, err := proxcli.New(test.setup())
			t.Require().NoError(err)

			serverHost := proxtest.StartHTTPEchoServer(t.T())
			t.assertHostIsReachable(serverHost, client)
		})
	}
}

func (t *ClientTest) assertHostIsReachable(h *addr.Host, c *proxcli.Client) {
	t.T().Helper()

	conn, err := c.Dial(context.Background(), h)
	t.Require().NoError(err)
	t.T().Cleanup(func() { conn.Close() })
}
