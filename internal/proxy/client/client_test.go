package client_test

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/client"
	"github.com/cerfical/socks2http/internal/proxy/mocks"
	"github.com/cerfical/socks2http/internal/proxy/socks4"
	"github.com/cerfical/socks2http/internal/proxy/socks5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

func TestClient(t *testing.T) {
	suite.Run(t, new(ClientTest))
}

type ClientTest struct {
	suite.Suite
}

func (t *ClientTest) TestDial() {
	t.Run("connects to destination directly if no proxy is used", func() {
		dstHost := addr.NewAddr("localhost", 8080)

		dialer := mocks.NewDialer(t.T())
		dialer.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(nil, nil)

		client, err := client.New(client.WithDialer(dialer))
		t.Require().NoError(err)

		_, err = client.Dial(context.Background(), dstHost)
		t.Require().NoError(err)
	})
}

func (t *ClientTest) TestDial_HTTP() {
	t.Run("makes a CONNECT request to proxy", func() {
		dstHost := addr.NewAddr("localhost", 8080)
		proxyConn := t.dialProxy(addr.ProtoHTTP, dstHost)

		req, err := http.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(dstHost.String(), req.Host)
		t.Equal(http.MethodConnect, req.Method)

		resp := httptest.NewRecorder()
		resp.WriteHeader(http.StatusOK)

		t.Require().NoError(resp.Result().Write(proxyConn))
	})
}

func (t *ClientTest) TestDial_SOCKS4() {
	t.Run("makes a CONNECT request to proxy", func() {
		proxyConn := t.dialProxy(addr.ProtoSOCKS4, addr.NewAddr("localhost", 8080))

		req, err := socks4.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks4.CommandConnect, req.Command)

		reply := socks4.Reply{Status: socks4.StatusGranted}
		t.Require().NoError(reply.Write(proxyConn))
	})

	t.Run("performs name resolution locally when using SOCKS4", func() {
		dstAddr := addr.NewAddr("localhost", 8080)
		proxyConn := t.dialProxy(addr.ProtoSOCKS4, dstAddr)

		req, err := socks4.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(addr.NewAddr("127.0.0.1", 8080), &req.DstAddr)

		reply := socks4.Reply{Status: socks4.StatusGranted}
		t.Require().NoError(reply.Write(proxyConn))
	})

	t.Run("delegates name resolution to proxy when using SOCKS4a", func() {
		dstAddr := addr.NewAddr("localhost", 8080)
		proxyConn := t.dialProxy(addr.ProtoSOCKS4a, dstAddr)

		req, err := socks4.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(dstAddr, &req.DstAddr)

		reply := socks4.Reply{Status: socks4.StatusGranted}
		t.Require().NoError(reply.Write(proxyConn))
	})
}

func (t *ClientTest) TestDial_SOCKS5() {
	dstAddr := addr.NewAddr("localhost", 8080)

	t.Run("makes a CONNECT request to proxy", func() {
		proxyConn := t.dialProxy(addr.ProtoSOCKS5, dstAddr)
		t.socks5Authenticate(proxyConn)

		req, err := socks5.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks5.CommandConnect, req.Command)

		reply := socks5.Reply{Status: socks5.StatusOK}
		t.Require().NoError(reply.Write(proxyConn))
	})

	t.Run("performs name resolution locally when using SOCKS5", func() {
		proxyConn := t.dialProxy(addr.ProtoSOCKS5, dstAddr)
		t.socks5Authenticate(proxyConn)

		req, err := socks5.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(addr.NewAddr("127.0.0.1", 8080), &req.DstAddr)

		reply := socks5.Reply{Status: socks5.StatusOK}
		t.Require().NoError(reply.Write(proxyConn))
	})

	t.Run("delegates name resolution to proxy when using SOCKS5h", func() {
		proxyConn := t.dialProxy(addr.ProtoSOCKS5h, dstAddr)
		t.socks5Authenticate(proxyConn)

		req, err := socks5.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(dstAddr, &req.DstAddr)

		reply := socks5.Reply{Status: socks5.StatusOK}
		t.Require().NoError(reply.Write(proxyConn))
	})
}

func (t *ClientTest) dialProxy(p addr.Proto, dstHost *addr.Addr) (proxyConn net.Conn) {
	clientConn, serverConn := net.Pipe()
	t.T().Cleanup(func() {
		clientConn.Close()
		serverConn.Close()
	})

	proxyAddr := addr.NewAddr("localhost", 1111)

	dialer := mocks.NewDialer(t.T())
	dialer.EXPECT().
		Dial(mock.Anything, proxyAddr).
		Return(clientConn, nil)

	client, err := client.New(
		client.WithProxyURL(addr.NewURL(p, proxyAddr.Host, proxyAddr.Port)),
		client.WithDialer(dialer),
	)
	t.Require().NoError(err)

	errChan := make(chan error, 1)
	go func() {
		_, err := client.Dial(context.Background(), dstHost)
		errChan <- err
	}()

	t.T().Cleanup(func() {
		t.Require().NoError(<-errChan)
	})

	return serverConn
}

func (t *ClientTest) socks5Authenticate(c net.Conn) {
	greet, err := socks5.ReadGreeting(bufio.NewReader(c))
	t.Require().NoError(err)

	t.ElementsMatch([]socks5.AuthMethod{socks5.AuthNone}, greet.AuthMethods)

	greetReply := socks5.GreetingReply{AuthMethod: socks5.AuthNone}
	t.Require().NoError(greetReply.Write(c))
}
