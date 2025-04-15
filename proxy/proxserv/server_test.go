package proxserv_test

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/proxy/proxserv"
	"github.com/cerfical/socks2http/socks"
	"github.com/cerfical/socks2http/test/mocks"
	"github.com/cerfical/socks2http/test/stubs"
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

		server, err := proxserv.New()
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
	t.Run("non-CONNECT requests are forwarded to the destination", func() {
		dstHost := addr.NewHost("localhost", 1111)

		proxy := mocks.NewProxy(t.T())
		proxy.EXPECT().
			ForwardHTTP(mock.Anything, mock.Anything, dstHost).
			Return(&http.Response{}, nil)

		proxyConn := t.openProxyConn(addr.HTTP, proxy)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://%v", dstHost), nil)
		t.writeHTTPRequest(req, proxyConn)
	})

	t.Run("responds to non-CONNECT requests with 502-Bad-Gateway if the destination is unreachable", func() {
		dstHost := addr.NewHost("unreachable-host", 1111)

		proxy := mocks.NewProxy(t.T())
		proxy.EXPECT().
			ForwardHTTP(mock.Anything, mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(addr.HTTP, proxy)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("http://%v", dstHost), nil)
		t.writeHTTPRequest(req, proxyConn)

		resp := t.readHTTPResponse(proxyConn)
		t.Equal(http.StatusBadGateway, resp.StatusCode)
	})

	t.Run("CONNECT opens a tunnel to the destination", func() {
		dstHost := addr.NewHost("localhost", 1111)

		proxy := mocks.NewProxy(t.T())
		proxy.EXPECT().
			OpenTunnel(mock.Anything, mock.Anything, dstHost).
			Return(dummyChan(), nil)

		proxyConn := t.openProxyConn(addr.HTTP, proxy)

		req := httptest.NewRequest(http.MethodConnect, dstHost.String(), nil)
		t.writeHTTPRequest(req, proxyConn)

		resp := t.readHTTPResponse(proxyConn)
		t.Equal(http.StatusOK, resp.StatusCode)
	})

	t.Run("responds to CONNECT with 502-Bad-Gateway if the destination is unreachable", func() {
		dstHost := addr.NewHost("unreachable-host", 1111)

		proxy := mocks.NewProxy(t.T())
		proxy.EXPECT().
			OpenTunnel(mock.Anything, mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(addr.HTTP, proxy)

		req := httptest.NewRequest(http.MethodConnect, dstHost.String(), nil)
		t.writeHTTPRequest(req, proxyConn)

		resp := t.readHTTPResponse(proxyConn)
		t.Equal(http.StatusBadGateway, resp.StatusCode)
	})
}

func (t *ServerTest) TestServe_SOCKS() {
	t.Run("CONNECT opens a tunnel to the destination", func() {
		dstHost := addr.NewHost("127.0.0.1", 1111)

		proxy := mocks.NewProxy(t.T())
		proxy.EXPECT().
			OpenTunnel(mock.Anything, mock.Anything, dstHost).
			Return(dummyChan(), nil)

		proxyConn := t.openProxyConn(addr.SOCKS4, proxy)

		req := socks.NewRequest(socks.V4, socks.Connect, dstHost)
		t.writeSOCKSRequest(req, proxyConn)

		reply := t.readSOCKSReply(proxyConn)
		t.Equal(socks.Granted, reply.Status)
	})

	t.Run("responds to CONNECT with Request-Rejected if the destination is unreachable", func() {
		dstHost := addr.NewHost("0.0.0.0", 0)

		proxy := mocks.NewProxy(t.T())
		proxy.EXPECT().
			OpenTunnel(mock.Anything, mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(addr.SOCKS4, proxy)

		req := socks.NewRequest(socks.V4, socks.Connect, dstHost)
		t.writeSOCKSRequest(req, proxyConn)

		reply := t.readSOCKSReply(proxyConn)
		t.Equal(socks.Rejected, reply.Status)
	})
}

func (t *ServerTest) openProxyConn(proto string, p proxy.Proxy) net.Conn {
	server, err := proxserv.New(
		proxserv.WithProto(proto),
		proxserv.WithProxy(p),
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

func (t *ServerTest) readHTTPResponse(r io.Reader) *http.Response {
	t.T().Helper()

	resp, err := http.ReadResponse(bufio.NewReader(r), nil)
	t.Require().NoError(err)
	t.T().Cleanup(func() { resp.Body.Close() })

	return resp
}

func (t *ServerTest) writeHTTPRequest(r *http.Request, w io.Writer) {
	t.T().Helper()

	t.Require().NoError(r.WriteProxy(w))
}

func (t *ServerTest) writeSOCKSRequest(r *socks.Request, w io.Writer) {
	t.T().Helper()

	t.Require().NoError(r.Write(w))
}

func (t *ServerTest) readSOCKSReply(r io.Reader) *socks.Reply {
	t.T().Helper()

	reply, err := socks.ReadReply(bufio.NewReader(r))
	t.Require().NoError(err)

	return reply
}

func dummyChan() chan error {
	ch := make(chan error)
	close(ch)
	return ch
}
