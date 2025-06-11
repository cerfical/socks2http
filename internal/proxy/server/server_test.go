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
	"github.com/cerfical/socks2http/internal/proxy/socks4"
	"github.com/cerfical/socks2http/internal/proxy/socks5"
	"github.com/cerfical/socks2http/internal/test/mocks"
	"github.com/cerfical/socks2http/internal/test/stubs"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestServer(t *testing.T) {
	suite.Run(t, new(ServerTest))
}

type ServerTest struct {
	suite.Suite
}

func (t *ServerTest) TestServe() {
	t.Run("performs graceful shutdown on context cancellation", func() {
		listener := stubs.NewIdleListener(1000, 50*time.Millisecond)

		server, err := server.New()
		t.Require().NoError(err)

		serveCtx, serveStop := context.WithCancel(context.Background())
		serveErr := make(chan error)
		go func() {
			serveErr <- server.Serve(serveCtx, listener)
		}()

		// Stop the server and check that all previously open connections are now closed
		serveStop()
		t.Require().NoError(<-serveErr)
		t.Equal(0, listener.OpenConns())
	})
}

func (t *ServerTest) TestServe_HTTP() {
	t.Run("non-CONNECT requests are forwarded to destination", func() {
		dstHost := addr.NewAddr("localhost", 1111)

		p := mocks.NewProxy(t.T())
		p.EXPECT().
			ForwardHTTP(mock.Anything, mock.Anything, dstHost).
			Return(&http.Response{}, nil)

		proxyConn := t.openProxyConn(addr.ProtoHTTP, p)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://%v", dstHost), nil)
		t.Require().NoError(req.WriteProxy(proxyConn))
	})

	t.Run("replies to non-CONNECT requests with 502-Bad-Gateway if destination is unreachable", func() {
		dstHost := addr.NewAddr("unreachable-host", 1111)

		p := mocks.NewProxy(t.T())
		p.EXPECT().
			ForwardHTTP(mock.Anything, mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(addr.ProtoHTTP, p)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://%v", dstHost), nil)
		t.Require().NoError(req.WriteProxy(proxyConn))

		resp, err := http.ReadResponse(bufio.NewReader(proxyConn), nil)
		t.Require().NoError(err)

		t.Equal(http.StatusBadGateway, resp.StatusCode)
	})

	t.Run("CONNECT opens a tunnel to destination", func() {
		dstHost := addr.NewAddr("localhost", 1111)

		p := mocks.NewProxy(t.T())
		p.EXPECT().
			OpenTunnel(mock.Anything, mock.Anything, dstHost).
			Return(dummyChan(), nil)

		proxyConn := t.openProxyConn(addr.ProtoHTTP, p)

		req := httptest.NewRequest(http.MethodConnect, dstHost.String(), nil)
		t.Require().NoError(req.WriteProxy(proxyConn))

		resp, err := http.ReadResponse(bufio.NewReader(proxyConn), nil)
		t.Require().NoError(err)

		t.Equal(http.StatusOK, resp.StatusCode)
	})

	t.Run("replies to CONNECT with 502-Bad-Gateway if destination is unreachable", func() {
		dstHost := addr.NewAddr("unreachable-host", 1111)

		p := mocks.NewProxy(t.T())
		p.EXPECT().
			OpenTunnel(mock.Anything, mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(addr.ProtoHTTP, p)

		req := httptest.NewRequest(http.MethodConnect, dstHost.String(), nil)
		t.Require().NoError(req.WriteProxy(proxyConn))

		resp, err := http.ReadResponse(bufio.NewReader(proxyConn), nil)
		t.Require().NoError(err)

		t.Equal(http.StatusBadGateway, resp.StatusCode)
	})
}

func (t *ServerTest) TestServe_SOCKS4() {
	t.Run("CONNECT opens a tunnel to destination", func() {
		dstHost := addr.NewAddr("127.0.0.1", 1111)

		p := mocks.NewProxy(t.T())
		p.EXPECT().
			OpenTunnel(mock.Anything, mock.Anything, dstHost).
			Return(dummyChan(), nil)

		proxyConn := t.openProxyConn(addr.ProtoSOCKS4, p)

		req := socks4.Request{
			Command: socks4.CommandConnect,
			DstAddr: *dstHost,
		}
		t.Require().NoError(req.Write(proxyConn))

		reply, err := socks4.ReadReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks4.StatusGranted, reply.Status)
	})

	t.Run("replies to CONNECT with Request-Rejected if destination is unreachable", func() {
		dstHost := addr.NewAddr("127.0.0.1", 1080)

		p := mocks.NewProxy(t.T())
		p.EXPECT().
			OpenTunnel(mock.Anything, mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(addr.ProtoSOCKS4, p)

		req := socks4.Request{
			Command: socks4.CommandConnect,
			DstAddr: *dstHost,
		}
		t.Require().NoError(req.Write(proxyConn))

		reply, err := socks4.ReadReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks4.StatusRejectedOrFailed, reply.Status)
	})
}

func (t *ServerTest) TestServe_SOCKS5() {
	t.Run("CONNECT opens a tunnel to destination", func() {
		dstHost := addr.NewAddr("localhost", 1111)

		p := mocks.NewProxy(t.T())
		p.EXPECT().
			OpenTunnel(mock.Anything, mock.Anything, dstHost).
			Return(dummyChan(), nil)

		proxyConn := t.openProxyConn(addr.ProtoSOCKS5, p)
		t.socks5Authenticate(proxyConn)

		req := socks5.Request{
			Command: socks5.CommandConnect,
			DstAddr: *dstHost,
		}
		t.Require().NoError(req.Write(proxyConn))

		reply, err := socks5.ReadReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks5.StatusOK, reply.Status)
	})

	t.Run("replies to non-CONNECT requests with Command-Not-Supported", func() {
		dstHost := addr.NewAddr("localhost", 1111)

		proxyConn := t.openProxyConn(addr.ProtoSOCKS5, nil)
		t.socks5Authenticate(proxyConn)

		req := socks5.Request{
			Command: socks5.CommandBind,
			DstAddr: *dstHost,
		}
		t.Require().NoError(req.Write(proxyConn))

		reply, err := socks5.ReadReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks5.StatusCommandNotSupported, reply.Status)
	})

	t.Run("replies to unsupported auth methods with Not-Acceptable", func() {
		proxyConn := t.openProxyConn(addr.ProtoSOCKS5, nil)

		greet := socks5.Greeting{AuthMethods: []socks5.AuthMethod{0xf0}}
		t.Require().NoError(greet.Write(proxyConn))

		greetReply, err := socks5.ReadGreetingReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks5.AuthNotAcceptable, greetReply.AuthMethod)
	})

	t.Run("replies to CONNECT with Host-Unreachable if destination is unreachable", func() {
		dstHost := addr.NewAddr("localhost", 1111)

		p := mocks.NewProxy(t.T())
		p.EXPECT().
			OpenTunnel(mock.Anything, mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(addr.ProtoSOCKS5, p)
		t.socks5Authenticate(proxyConn)

		req := socks5.Request{
			Command: socks5.CommandConnect,
			DstAddr: *dstHost,
		}
		t.Require().NoError(req.Write(proxyConn))

		reply, err := socks5.ReadReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks5.StatusHostUnreachable, reply.Status)
	})
}

func (t *ServerTest) openProxyConn(proto addr.Proto, proxy proxy.Proxy) net.Conn {
	server, err := server.New(
		server.WithServeProto(proto),
		server.WithProxy(proxy),
	)
	t.Require().NoError(err)

	l, err := net.Listen("tcp", "localhost:0")
	t.Require().NoError(err)

	serveErr := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		serveErr <- server.Serve(ctx, l)
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

func (t *ServerTest) socks5Authenticate(c net.Conn) {
	greet := socks5.Greeting{
		AuthMethods: []socks5.AuthMethod{socks5.AuthNone},
	}
	t.Require().NoError(greet.Write(c))

	greetReply, err := socks5.ReadGreetingReply(bufio.NewReader(c))
	t.Require().NoError(err)

	t.Equal(socks5.AuthNone, greetReply.AuthMethod)
}

func dummyChan() chan error {
	ch := make(chan error)
	close(ch)
	return ch
}
