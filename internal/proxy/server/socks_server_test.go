package server_test

import (
	"bufio"
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/mocks"
	"github.com/cerfical/socks2http/internal/proxy/server"
	"github.com/cerfical/socks2http/internal/proxy/socks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestSOCKSServer(t *testing.T) {
	suite.Run(t, new(SOCKSServerTest))
}

type SOCKSServerTest struct {
	suite.Suite
}

func (t *SOCKSServerTest) TestServeSOCKS() {
	t.Run("performs graceful shutdown on context cancellation", func() {
		listener := NewIdleListener(1000, 50*time.Millisecond)

		server := server.SOCKSServer{
			Tunneler: proxy.DefaultTunneler,
			Dialer:   proxy.DirectDialer,
			Log:      proxy.DiscardLogger,
		}

		serveCtx, serveStop := context.WithCancel(context.Background())
		serveErr := make(chan error)
		go func() {
			serveErr <- server.ServeSOCKS(serveCtx, listener)
		}()

		serveStop()
		t.Require().NoError(<-serveErr)
		t.Equal(0, listener.OpenConns())
	})
}

func (t *SOCKSServerTest) TestServeSOCKS_V4() {
	t.Run("CONNECT opens a tunnel to destination", func() {
		dstHost := addr.NewAddr("127.0.0.1", 1111)
		dstConn := NewDummyConn()

		dial := mocks.NewDialer(t.T())
		dial.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(dstConn, nil)
		tun := mocks.NewTunneler(t.T())
		tun.EXPECT().
			Tunnel(mock.Anything, mock.Anything, dstConn).
			Return(nil)

		proxyConn := t.openProxyConn(tun, dial)

		req := socks.Request{
			Version: socks.V4,
			Command: socks.CommandConnect,
			DstAddr: *dstHost,
		}
		t.Require().NoError(req.Write(proxyConn))

		reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks.StatusGranted, reply.Status)
	})

	t.Run("replies to CONNECT with General-Failure if destination is unreachable", func() {
		dstHost := addr.NewAddr("127.0.0.1", 1080)

		dial := mocks.NewDialer(t.T())
		dial.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(nil, dial)

		req := socks.Request{
			Version: socks.V4,
			Command: socks.CommandConnect,
			DstAddr: *dstHost,
		}
		t.Require().NoError(req.Write(proxyConn))

		reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks.StatusGeneralFailure, reply.Status)
	})
}

func (t *SOCKSServerTest) TestServeSOCKS_V5() {
	t.Run("CONNECT opens a tunnel to destination", func() {
		dstHost := addr.NewAddr("localhost", 1111)
		dstConn := NewDummyConn()

		dial := mocks.NewDialer(t.T())
		dial.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(dstConn, nil)

		tun := mocks.NewTunneler(t.T())
		tun.EXPECT().
			Tunnel(mock.Anything, mock.Anything, dstConn).
			Return(nil)

		proxyConn := t.openProxyConn(tun, dial)
		t.socks5Authenticate(proxyConn)

		req := socks.Request{
			Version: socks.V5,
			Command: socks.CommandConnect,
			DstAddr: *dstHost,
		}
		t.Require().NoError(req.Write(proxyConn))

		reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks.StatusGranted, reply.Status)
	})

	t.Run("replies to non-CONNECT requests with Command-Not-Supported", func() {
		dstHost := addr.NewAddr("localhost", 1111)

		proxyConn := t.openProxyConn(nil, nil)
		t.socks5Authenticate(proxyConn)

		req := socks.Request{
			Version: socks.V5,
			Command: socks.CommandBind,
			DstAddr: *dstHost,
		}
		t.Require().NoError(req.Write(proxyConn))

		reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks.StatusCommandNotSupported, reply.Status)
	})

	t.Run("replies to unsupported auth methods with Not-Acceptable", func() {
		proxyConn := t.openProxyConn(nil, nil)

		greet := socks.Greeting{
			Version: socks.V5,
			Auth:    []socks.Auth{0xf0},
		}
		t.Require().NoError(greet.Write(proxyConn))

		greetReply, err := socks.ReadGreetingReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks.AuthNotAcceptable, greetReply.Auth)
	})

	t.Run("replies to CONNECT with Host-Unreachable if destination is unreachable", func() {
		dstHost := addr.NewAddr("localhost", 1111)

		dial := mocks.NewDialer(t.T())
		dial.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(nil, dial)
		t.socks5Authenticate(proxyConn)

		req := socks.Request{
			Version: socks.V5,
			Command: socks.CommandConnect,
			DstAddr: *dstHost,
		}
		t.Require().NoError(req.Write(proxyConn))

		reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks.StatusHostUnreachable, reply.Status)
	})
}

func (t *SOCKSServerTest) openProxyConn(tun proxy.Tunneler, dial proxy.Dialer) net.Conn {
	server := server.SOCKSServer{
		Tunneler: tun,
		Dialer:   dial,
		Log:      proxy.DiscardLogger,
	}

	l, err := net.Listen("tcp", "localhost:0")
	t.Require().NoError(err)

	serveErr := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		serveErr <- server.ServeSOCKS(ctx, l)
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

func (t *SOCKSServerTest) socks5Authenticate(c net.Conn) {
	greet := socks.Greeting{
		Version: socks.V5,
		Auth:    []socks.Auth{socks.AuthNone},
	}
	t.Require().NoError(greet.Write(c))

	greetReply, err := socks.ReadGreetingReply(bufio.NewReader(c))
	t.Require().NoError(err)

	t.Equal(socks.AuthNone, greetReply.Auth)
}
