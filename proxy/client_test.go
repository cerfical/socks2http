package proxy_test

import (
	"context"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy"
	"github.com/stretchr/testify/suite"
)

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTest))
}

type ClientTest struct {
	ProxyTest
}

func (t *ClientTest) TestNew() {
	tests := map[string]struct {
		options proxy.ClientOption
		want    func(*proxy.Client)
		err     func(error)
	}{
		"uses http://localhost:8080 as the default proxy address": {
			want: func(c *proxy.Client) {
				t.Equal(addr.New(addr.HTTP, "localhost", 8080), c.ProxyAddr())
			},
		},

		"uses a non-default proxy address if one is provided": {
			options: proxy.WithProxyAddr(addr.New(addr.HTTP, "example.com", 8181)),
			want: func(c *proxy.Client) {
				t.Equal(addr.New(addr.HTTP, "example.com", 8181), c.ProxyAddr())
			},
		},

		"rejects unsupported protocol schemes": {
			options: proxy.WithProxyAddr(addr.New("SOCKS9", "", 0)),
			err: func(err error) {
				t.ErrorContains(err, "SOCKS9")
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			ops := []proxy.ClientOption{}
			if test.options != nil {
				ops = append(ops, test.options)
			}

			client, err := proxy.NewClient(ops...)
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
		setup func() proxy.ClientOption
	}{
		"establishes a direct connection to a server if Direct is used": {
			setup: func() proxy.ClientOption {
				return proxy.WithProxyAddr(addr.New(addr.Direct, "", 0))
			},
		},

		"connects to a server via an HTTP proxy": {
			setup: func() proxy.ClientOption {
				httpProxy := t.startProxyServer(addr.HTTP).ListenAddr()
				return proxy.WithProxyAddr(httpProxy)
			},
		},

		"connects to a server via a SOCKS4 proxy": {
			setup: func() proxy.ClientOption {
				socks4Proxy := t.startProxyServer(addr.SOCKS4).ListenAddr()
				return proxy.WithProxyAddr(socks4Proxy)
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			client, err := proxy.NewClient(test.setup())
			t.Require().NoError(err)

			serverHost := t.startHTTPEchoServer()
			t.assertHostIsReachable(serverHost, client)
		})
	}
}

func (t *ClientTest) assertHostIsReachable(h *addr.Host, c *proxy.Client) {
	t.T().Helper()

	conn, err := c.Dial(context.Background(), h)
	t.Require().NoError(err)
	t.T().Cleanup(func() { conn.Close() })
}
