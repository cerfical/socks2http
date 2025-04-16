package proxy_test

import (
	"context"
	"net"
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

func (t *ProxyTest) TestProxy_OpenTunnel() {
	t.Run("closing a destination end of a tunnel closes the tunnel", func() {
		tunnelDone, _, dstConn := t.openTunnel()
		dstConn.Close()

		t.assertTunnelClosed(tunnelDone)
	})

	t.Run("closing a source end of a tunnel closes the tunnel", func() {
		tunnelDone, srcConn, _ := t.openTunnel()
		srcConn.Close()

		t.assertTunnelClosed(tunnelDone)
	})
}

func (t *ProxyTest) assertTunnelClosed(tunnelDone <-chan error) {
	select {
	case err := <-tunnelDone:
		t.NoError(err)
	case <-time.After(TunnelCloseTimeout):
		t.Fail("tunnel wasn't closed", "tunnel was expected to close in %v, but it didn't", TunnelCloseTimeout)
	}
}

func (t *ProxyTest) openTunnel() (done <-chan error, srcConn, dstConn net.Conn) {
	srcClientConn, srcProxyConn := net.Pipe()
	dstServerConn, dstProxyConn := net.Pipe()
	t.T().Cleanup(func() {
		srcClientConn.Close()
		srcProxyConn.Close()
		dstServerConn.Close()
		dstProxyConn.Close()

		goleak.VerifyNone(t.T())
	})

	dstHost := addr.NewHost("localhost", 1111)

	dialer := mocks.NewDialer(t.T())
	dialer.EXPECT().
		Dial(mock.Anything, dstHost).
		Return(dstProxyConn, nil)

	proxy := proxy.New(dialer)

	tunnelDone, err := proxy.OpenTunnel(context.Background(), srcProxyConn, dstHost)
	t.Require().NoError(err)

	return tunnelDone, srcClientConn, dstServerConn
}
