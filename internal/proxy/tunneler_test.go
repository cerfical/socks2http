package proxy_test

import (
	"context"
	"io"
	"net"
	"testing"
	"time"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/stretchr/testify/suite"
	"go.uber.org/goleak"
)

const TunnelCloseTimeout = 2 * time.Second

func TestTunneler(t *testing.T) {
	suite.Run(t, new(TunnelerTest))
}

type TunnelerTest struct {
	suite.Suite
}

func (t *TunnelerTest) TestTunnel() {
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

func (t *TunnelerTest) openTunnel() (tunnelDone <-chan error, srcConn, dstConn net.Conn) {
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

	tunnelErr := make(chan error, 1)
	go func() {
		tunneler := proxy.DefaultTunneler
		tunnelErr <- tunneler.Tunnel(context.Background(), srcProxyConn, dstProxyConn)
	}()

	return tunnelErr, srcClientConn, dstServerConn
}

func (t *TunnelerTest) assertTunnelClosed(tunnelDone <-chan error) {
	select {
	case err := <-tunnelDone:
		t.NoError(err)
	case <-time.After(TunnelCloseTimeout):
		t.Fail("tunnel wasn't closed", "tunnel was expected to close in %v, but it didn't", TunnelCloseTimeout)
	}
}

func (t *TunnelerTest) readString(r io.Reader, n int) string {
	t.T().Helper()

	buf := make([]byte, n)
	_, err := r.Read(buf)
	t.Require().NoError(err)

	return string(buf)
}

func (t *TunnelerTest) writeString(w io.Writer, s string) {
	t.T().Helper()

	_, err := w.Write([]byte(s))
	t.Require().NoError(err)
}
