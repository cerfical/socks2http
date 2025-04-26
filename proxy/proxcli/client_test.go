package proxcli_test

import (
	"bufio"
	"context"
	"io"
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
	t.Run("connects to an HTTP proxy", func() {
		dstHost := addr.NewHost("localhost", 8080)
		proxyConn := t.dialProxy(addr.HTTP, dstHost)

		req := t.readHTTPRequest(proxyConn)
		t.Equal(dstHost.String(), req.Host)
		t.Equal(http.MethodConnect, req.Method)

		t.writeHTTPStatus(http.StatusOK, proxyConn)
	})

	t.Run("connects to a SOCKS4 proxy", func() {
		dstHost := addr.NewHost("127.0.0.1", 8080)
		proxyConn := t.dialProxy(addr.SOCKS4, dstHost)

		req := t.readSOCKSRequest(proxyConn)
		t.Equal(dstHost, &req.DstAddr)
		t.Equal(socks4.CommandConnect, req.Command)

		t.writeSOCKSReply(socks4.StatusGranted, proxyConn)
	})

	t.Run("connects to a SOCKS4a proxy", func() {
		dstHost := addr.NewHost("localhost", 8080)
		proxyConn := t.dialProxy(addr.SOCKS4a, dstHost)

		req := t.readSOCKSRequest(proxyConn)
		t.Equal(dstHost, &req.DstAddr)
		t.Equal(socks4.CommandConnect, req.Command)

		t.writeSOCKSReply(socks4.StatusGranted, proxyConn)
	})

	t.Run("establishes a direct connection to the destination if Direct is used", func() {
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

func (t *ClientTest) TestDial_SOCKS5() {
	dstAddr := addr.NewHost("localhost", 8080)
	requestTests := map[string]struct {
		proto string

		wantGreeting func(*socks5.Greeting)
		wantRequest  func(*socks5.Request)
	}{
		"makes a CONNECT request to proxy": {
			wantRequest: func(r *socks5.Request) {
				t.Equal(socks5.CommandConnect, r.Command)
			},
		},

		"resolves destination address when using SOCKS5 client": {
			wantRequest: func(r *socks5.Request) {
				t.Equal(addr.NewHost("127.0.0.1", dstAddr.Port), &r.DstAddr)
			},
		},

		"doesn't resolve destination address when using SOCKS5h client": {
			proto: addr.SOCKS5h,
			wantRequest: func(r *socks5.Request) {
				t.Equal(dstAddr, &r.DstAddr)
			},
		},

		"uses no authentication": {
			wantGreeting: func(g *socks5.Greeting) {
				t.ElementsMatch([]socks5.AuthMethod{socks5.AuthNone}, g.AuthMethods)
			},
		},
	}

	for name, test := range requestTests {
		t.Run(name, func() {
			if test.proto == "" {
				test.proto = addr.SOCKS5
			}
			proxyConn := t.dialProxy(test.proto, dstAddr)

			greet, err := socks5.ReadGreeting(bufio.NewReader(proxyConn))
			t.Require().NoError(err)
			if test.wantGreeting != nil {
				test.wantGreeting(greet)
			}

			greetReply := socks5.GreetingReply{AuthMethod: socks5.AuthNone}
			t.Require().NoError(greetReply.Write(proxyConn))

			req, err := socks5.ReadRequest(bufio.NewReader(proxyConn))
			t.Require().NoError(err)
			if test.wantRequest != nil {
				test.wantRequest(req)
			}

			reply := socks5.Reply{Status: socks5.StatusOK}
			t.Require().NoError(reply.Write(proxyConn))
		})
	}
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

func (t *ClientTest) writeHTTPStatus(status int, w io.Writer) {
	t.T().Helper()

	resp := httptest.NewRecorder()
	resp.WriteHeader(status)

	t.Require().NoError(resp.Result().Write(w))
}

func (t *ClientTest) readHTTPRequest(r io.Reader) *http.Request {
	t.T().Helper()

	req, err := http.ReadRequest(bufio.NewReader(r))
	t.Require().NoError(err)

	return req
}

func (t *ClientTest) writeSOCKSReply(s socks4.Status, w io.Writer) {
	t.T().Helper()

	reply := socks4.NewReply(s, nil)
	t.Require().NoError(reply.Write(w))
}

func (t *ClientTest) readSOCKSRequest(r io.Reader) *socks4.Request {
	t.T().Helper()

	req, err := socks4.ReadRequest(bufio.NewReader(r))
	t.Require().NoError(err)

	return req
}
