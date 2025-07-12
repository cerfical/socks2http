package server_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/server"
	"github.com/cerfical/socks2http/internal/test/mocks"
	"github.com/cerfical/socks2http/internal/test/stubs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestHTTPServer(t *testing.T) {
	suite.Run(t, new(HTTPServerTest))
}

type HTTPServerTest struct {
	suite.Suite
}

func (t *HTTPServerTest) TestServeHTTP() {
	t.Run("performs graceful shutdown on context cancellation", func() {
		listener := stubs.NewIdleListener(1000, 50*time.Millisecond)

		server := server.HTTPServer{
			Tunneler: proxy.DefaultTunneler,
			Dialer:   proxy.DirectDialer,
			Log:      proxy.DiscardLogger,
		}

		serveCtx, serveStop := context.WithCancel(context.Background())
		serveErr := make(chan error)
		go func() {
			serveErr <- server.ServeHTTP(serveCtx, listener)
		}()

		serveStop()
		t.Require().NoError(<-serveErr)
		t.Equal(0, listener.OpenConns())
	})

	t.Run("non-CONNECT requests are forwarded to destination", func() {
		dstHost := addr.NewAddr("localhost", 1111)
		dstServerConn, dstProxyConn := net.Pipe()

		dial := mocks.NewDialer(t.T())
		dial.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(dstProxyConn, nil)

		proxyConn := t.openProxyConn(nil, dial)

		// Simulate a client HTTP GET request through the proxy
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://%v", dstHost), nil)
		t.Require().NoError(req.WriteProxy(proxyConn))

		// Ensure the request is received by the destination
		_, err := http.ReadRequest(bufio.NewReader(dstServerConn))
		t.Require().NoError(err)

		// Simulate a response from the destination
		dstResp := http.Response{
			StatusCode: http.StatusOK,
			ProtoMajor: 1,
			ProtoMinor: 1,
		}
		t.Require().NoError(dstResp.Write(dstServerConn))

		// Ensure the proxy returns the response to the client
		proxyResp, err := http.ReadResponse(bufio.NewReader(proxyConn), nil)
		t.Require().NoError(err)
		t.Equal(http.StatusOK, proxyResp.StatusCode)
	})

	t.Run("replies to non-CONNECT requests with 502-Bad-Gateway if destination is unreachable", func() {
		dstHost := addr.NewAddr("unreachable-host", 1111)

		dial := mocks.NewDialer(t.T())
		dial.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(nil, dial)

		// Simulate a client HTTP GET request through the proxy
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://%v", dstHost), nil)
		t.Require().NoError(req.WriteProxy(proxyConn))

		// Ensure the proxy returns 502 Bad Gateway to the client
		resp, err := http.ReadResponse(bufio.NewReader(proxyConn), nil)
		t.Require().NoError(err)
		t.Equal(http.StatusBadGateway, resp.StatusCode)
	})

	t.Run("CONNECT opens a tunnel to destination", func() {
		dstHost := addr.NewAddr("localhost", 1111)
		dstConn := stubs.NewDummyConn()

		dial := mocks.NewDialer(t.T())
		dial.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(dstConn, nil)

		tun := mocks.NewTunneler(t.T())
		tun.EXPECT().
			Tunnel(mock.Anything, mock.Anything, dstConn).
			Return(nil)

		proxyConn := t.openProxyConn(tun, dial)

		req := httptest.NewRequest(http.MethodConnect, dstHost.String(), nil)
		t.Require().NoError(req.WriteProxy(proxyConn))

		resp, err := http.ReadResponse(bufio.NewReader(proxyConn), nil)
		t.Require().NoError(err)

		t.Equal(http.StatusOK, resp.StatusCode)
	})

	t.Run("replies to CONNECT with 502-Bad-Gateway if destination is unreachable", func() {
		dstHost := addr.NewAddr("unreachable-host", 1111)

		dial := mocks.NewDialer(t.T())
		dial.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(nil, dial)

		req := httptest.NewRequest(http.MethodConnect, dstHost.String(), nil)
		t.Require().NoError(req.WriteProxy(proxyConn))

		resp, err := http.ReadResponse(bufio.NewReader(proxyConn), nil)
		t.Require().NoError(err)

		t.Equal(http.StatusBadGateway, resp.StatusCode)
	})
}

func (t *HTTPServerTest) openProxyConn(tun proxy.Tunneler, dial proxy.Dialer) net.Conn {
	l, err := net.Listen("tcp", "localhost:0")
	t.Require().NoError(err)

	serveErr := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		server := server.HTTPServer{
			Tunneler: tun,
			Dialer:   dial,
			Log:      proxy.DiscardLogger,
		}

		serveErr <- server.ServeHTTP(ctx, l)
	}()
	t.T().Cleanup(func() {
		cancel()
		t.Require().NoError(<-serveErr)
	})

	conn, err := net.Dial("tcp", l.Addr().String())
	t.Require().NoError(err)
	t.T().Cleanup(func() { conn.Close() })

	return conn
}
