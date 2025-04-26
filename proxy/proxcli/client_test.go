package proxcli_test

import (
	"bufio"
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy/proxcli"
	"github.com/cerfical/socks2http/socks4"
	"github.com/cerfical/socks2http/socks5"
	"github.com/cerfical/socks2http/test/mocks"
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
	t.Run("connects to destination directly if Direct is used", func() {
		dstHost := addr.NewHost("localhost", 8080)

		dialer := mocks.NewDialer(t.T())
		dialer.EXPECT().
			Dial(mock.Anything, dstHost).
			Return(nil, nil)

		client, err := proxcli.New(
			proxcli.WithProxyAddr(addr.New(addr.Direct, "", 0)),
			proxcli.WithDialer(dialer),
		)
		t.Require().NoError(err)

		_, err = client.Dial(context.Background(), dstHost)
		t.Require().NoError(err)
	})
}

func (t *ClientTest) TestDial_HTTP() {
	t.Run("makes a CONNECT request to proxy", func() {
		dstHost := addr.NewHost("localhost", 8080)
		proxyConn := t.dialProxy(addr.HTTP, dstHost)

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
		proxyConn := t.dialProxy(addr.SOCKS4, addr.NewHost("localhost", 8080))

		req, err := socks4.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks4.CommandConnect, req.Command)

		reply := socks4.Reply{Status: socks4.StatusGranted}
		t.Require().NoError(reply.Write(proxyConn))
	})

	t.Run("performs name resolution locally when using SOCKS4", func() {
		dstAddr := addr.NewHost("localhost", 8080)
		proxyConn := t.dialProxy(addr.SOCKS4, dstAddr)

		req, err := socks4.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(addr.NewHost("127.0.0.1", 8080), &req.DstAddr)

		reply := socks4.Reply{Status: socks4.StatusGranted}
		t.Require().NoError(reply.Write(proxyConn))
	})

	t.Run("delegates name resolution to proxy when using SOCKS4a", func() {
		dstAddr := addr.NewHost("localhost", 8080)
		proxyConn := t.dialProxy(addr.SOCKS4a, dstAddr)

		req, err := socks4.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(dstAddr, &req.DstAddr)

		reply := socks4.Reply{Status: socks4.StatusGranted}
		t.Require().NoError(reply.Write(proxyConn))
	})
}

func (t *ClientTest) TestDial_SOCKS5() {
	dstAddr := addr.NewHost("localhost", 8080)

	t.Run("makes a CONNECT request to proxy", func() {
		proxyConn := t.dialProxy(addr.SOCKS5, dstAddr)
		t.socks5Authenticate(proxyConn)

		req, err := socks5.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(socks5.CommandConnect, req.Command)

		reply := socks5.Reply{Status: socks5.StatusOK}
		t.Require().NoError(reply.Write(proxyConn))
	})

	t.Run("performs name resolution locally when using SOCKS5", func() {
		proxyConn := t.dialProxy(addr.SOCKS5, dstAddr)
		t.socks5Authenticate(proxyConn)

		req, err := socks5.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(addr.NewHost("127.0.0.1", 8080), &req.DstAddr)

		reply := socks5.Reply{Status: socks5.StatusOK}
		t.Require().NoError(reply.Write(proxyConn))
	})

	t.Run("delegates name resolution to proxy when using SOCKS5h", func() {
		proxyConn := t.dialProxy(addr.SOCKS5h, dstAddr)
		t.socks5Authenticate(proxyConn)

		req, err := socks5.ReadRequest(bufio.NewReader(proxyConn))
		t.Require().NoError(err)

		t.Equal(dstAddr, &req.DstAddr)

		reply := socks5.Reply{Status: socks5.StatusOK}
		t.Require().NoError(reply.Write(proxyConn))
	})
}

func (t *ClientTest) dialProxy(proto string, dstHost *addr.Host) (proxyConn net.Conn) {
	clientConn, serverConn := net.Pipe()
	t.T().Cleanup(func() {
		clientConn.Close()
		serverConn.Close()
	})

	proxyAddr := addr.New(proto, "localhost", 1111)

	dialer := mocks.NewDialer(t.T())
	dialer.EXPECT().
		Dial(mock.Anything, &proxyAddr.Host).
		Return(clientConn, nil)

	client, err := proxcli.New(
		proxcli.WithProxyAddr(proxyAddr),
		proxcli.WithDialer(dialer),
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
