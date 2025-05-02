package proxy_test

import (
	"bufio"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/test/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

const TunnelCloseTimeout = 2 * time.Second

func TestProxy(t *testing.T) {
	suite.Run(t, new(ProxyTest))
}

type ProxyTest struct {
	suite.Suite
}

func (t *ProxyTest) TestProxy_ForwardHTTP() {
	t.Run("forwards an HTTP request to the destination and reads the response", func() {
		dstServerConn, dstProxyConn := net.Pipe()

		dstHost := addr.New("localhost", 1111)

		dialer := mocks.NewDialer(t.T())
		dialer.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(dstProxyConn, nil)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		respChan, errChan := t.forwardHTTP(req, dstHost, dialer)

		_ = t.readHTTPRequest(dstServerConn)
		t.writeHTTPStatus(http.StatusOK, dstServerConn)

		t.Require().NoError(<-errChan)
		t.Equal(http.StatusOK, (<-respChan).StatusCode)
	})
}

func (t *ProxyTest) TestProxy_OpenTunnel() {
	t.Run("closing the destination closes the tunnel", func() {
		tunnelDone, _, dstConn := t.openTunnel()
		dstConn.Close()

		t.assertTunnelClosed(tunnelDone)
	})

	t.Run("closing the source closes the tunnel", func() {
		tunnelDone, srcConn, _ := t.openTunnel()
		srcConn.Close()

		t.assertTunnelClosed(tunnelDone)
	})

	t.Run("data sent from the destination is received at the source", func() {
		tunnelDone, srcConn, dstConn := t.openTunnel()
		defer func() {
			srcConn.Close()
			dstConn.Close()
			<-tunnelDone
		}()

		want := "abcd"

		t.writeString(dstConn, want)
		got := t.readString(srcConn, len(want))

		t.Equal(want, got)
	})

	t.Run("data sent from the source is received at the destination", func() {
		tunnelDone, srcConn, dstConn := t.openTunnel()
		defer func() {
			srcConn.Close()
			dstConn.Close()
			<-tunnelDone
		}()

		want := "abcd"

		t.writeString(srcConn, want)
		got := t.readString(dstConn, len(want))

		t.Equal(want, got)
	})
}

func (t *ProxyTest) openTunnel() (done <-chan error, srcConn, dstConn net.Conn) {
	srcClientConn, srcProxyConn := net.Pipe()
	dstServerConn, dstProxyConn := net.Pipe()

	// Make sure the tunnel is properly closed and leaks no goroutines
	t.T().Cleanup(func() {
		srcClientConn.Close()
		srcProxyConn.Close()
		dstServerConn.Close()
		dstProxyConn.Close()

		goleak.VerifyNone(t.T())
	})

	dstHost := addr.New("localhost", 1111)

	dialer := mocks.NewDialer(t.T())
	dialer.EXPECT().
		Dial(mock.Anything, dstHost).
		Return(dstProxyConn, nil)

	proxy := proxy.New(dialer)

	tunnelDone, err := proxy.OpenTunnel(context.Background(), srcProxyConn, dstHost)
	t.Require().NoError(err)

	return tunnelDone, srcClientConn, dstServerConn
}

func (t *ProxyTest) assertTunnelClosed(tunnelDone <-chan error) {
	select {
	case err := <-tunnelDone:
		t.NoError(err)
	case <-time.After(TunnelCloseTimeout):
		t.Fail("tunnel wasn't closed", "tunnel was expected to close in %v, but it didn't", TunnelCloseTimeout)
	}
}

func (t *ProxyTest) forwardHTTP(r *http.Request, dstHost *addr.Addr, d proxy.Dialer) (<-chan *http.Response, <-chan error) {
	respChan := make(chan *http.Response, 1)
	errChan := make(chan error, 1)

	go func() {
		proxy := proxy.New(d)
		resp, err := proxy.ForwardHTTP(context.Background(), r, dstHost)

		respChan <- resp
		errChan <- err
	}()

	return respChan, errChan
}

func (t *ProxyTest) readHTTPRequest(r io.Reader) *http.Request {
	t.T().Helper()

	req, err := http.ReadRequest(bufio.NewReader(r))
	t.Require().NoError(err)

	return req
}

func (t *ProxyTest) writeHTTPStatus(status int, w io.Writer) {
	t.T().Helper()

	resp := http.Response{ProtoMajor: 1, ProtoMinor: 1, StatusCode: status}
	t.Require().NoError(resp.Write(w))
}

func (t *ProxyTest) readString(r io.Reader, n int) string {
	t.T().Helper()

	buf := make([]byte, n)
	_, err := r.Read(buf)
	t.Require().NoError(err)

	return string(buf)
}

func (t *ProxyTest) writeString(w io.Writer, s string) {
	t.T().Helper()

	_, err := w.Write([]byte(s))
	t.Require().NoError(err)
}
