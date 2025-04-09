package proxy_test

import (
	"bufio"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/suite"
)

func TestSOCKSProxy(t *testing.T) {
	suite.Run(t, new(SOCKSProxyTest))
}

type SOCKSProxyTest struct {
	ProxyTest
}

func (t *SOCKSProxyTest) TestServe() {
	serverHost := t.startHTTPEchoServer()
	tests := map[string]struct {
		dstHost   *addr.Host
		wantReply socks.Reply
	}{
		"supports tunneling of HTTP requests": {
			dstHost:   serverHost,
			wantReply: socks.Granted,
		},

		"responds with a Request Rejected if the server is unreachable": {
			dstHost:   addr.NewHost("0.0.0.0", 0),
			wantReply: socks.Rejected,
		},
	}

	for name, test := range tests {
		t.Run(name, func() {
			// TODO: Ensure the proxy is closed
			proxyConn, _ := t.proxyServe(addr.New(addr.SOCKS4, "localhost", 1080))

			// Write a SOCKS CONNECT request
			req := socks.NewRequest(socks.V4, socks.Connect, test.dstHost)
			t.Require().NoError(req.Write(proxyConn))

			reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
			t.Require().NoError(err)

			t.Equal(test.wantReply, reply)
			if reply == socks.Granted {
				// Make a bunch of HTTP requests
				for range numInRowRequests {
					msg := makeString(maxPayloadSize)

					req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(msg))
					t.Require().NoError(req.Write(proxyConn))
					resp := t.readHTTPResponse(proxyConn)

					t.Equal(http.StatusOK, resp.StatusCode)
					t.Equal(msg, t.readString(resp.Body))
				}
			}
		})
	}
}

func (t *SOCKSProxyTest) TestConnect() {
	t.Run("makes a SOCKS CONNECT request to the proxy", func() {
		// Use a raw IPv4 address, as SOCKS4 doesn't support domain names
		dstHost := addr.NewHost("127.0.0.1", 9999)
		proxyServerConn := t.proxyConnect(
			addr.New(addr.SOCKS4, "localhost", 1080),
			dstHost,
		)

		// Assert that the proxy server receives the correct request
		req, err := socks.ReadRequest(bufio.NewReader(proxyServerConn))
		t.Require().NoError(err)

		t.Equal(socks.V4, req.Version)
		t.Equal(dstHost, &req.Host)
		t.Equal(socks.Connect, req.Command)

		// Write a SOCKS Granted reply so the client can proceed
		t.Require().NoError(socks.Granted.Write(proxyServerConn))
	})
}
