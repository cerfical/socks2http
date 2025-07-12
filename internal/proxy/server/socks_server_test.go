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
	"github.com/cerfical/socks2http/internal/proxy/socks4"
	"github.com/cerfical/socks2http/internal/proxy/socks5"
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

func (t *SOCKSServerTest) TestServeSOCKS4() {
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

		dial := mocks.NewDialer(t.T())
		dial.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(nil, dial)

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

func (t *SOCKSServerTest) TestServeSOCKS5() {
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

		proxyConn := t.openProxyConn(nil, nil)
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
		proxyConn := t.openProxyConn(nil, nil)

		greet := socks5.Greeting{AuthMethods: []socks5.AuthMethod{0xf0}}
		t.Require().NoError(greet.Write(proxyConn))

		greetReply, err := socks5.ReadGreetingReply(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks5.AuthNotAcceptable, greetReply.AuthMethod)
	})

	t.Run("replies to CONNECT with Host-Unreachable if destination is unreachable", func() {
		dstHost := addr.NewAddr("localhost", 1111)

		dial := mocks.NewDialer(t.T())
		dial.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(nil, errors.New("unreachable host"))

		proxyConn := t.openProxyConn(nil, dial)
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
	greet := socks5.Greeting{
		AuthMethods: []socks5.AuthMethod{socks5.AuthNone},
	}
	t.Require().NoError(greet.Write(c))

	greetReply, err := socks5.ReadGreetingReply(bufio.NewReader(c))
	t.Require().NoError(err)

	t.Equal(socks5.AuthNone, greetReply.AuthMethod)
}
