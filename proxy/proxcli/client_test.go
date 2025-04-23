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
	"github.com/cerfical/socks2http/socks"
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
		t.Equal(socks.V4, req.Version)
		t.Equal(dstHost, &req.DstAddr)
		t.Equal(socks.Connect, req.Command)

		t.writeSOCKSReply(socks.Granted, proxyConn)
	})

	t.Run("connects to a SOCKS4a proxy", func() {
		dstHost := addr.NewHost("localhost", 8080)
		proxyConn := t.dialProxy(addr.SOCKS4a, dstHost)

		req := t.readSOCKSRequest(proxyConn)
		t.Equal(socks.V4a, req.Version)
		t.Equal(dstHost, &req.DstAddr)
		t.Equal(socks.Connect, req.Command)

		t.writeSOCKSReply(socks.Granted, proxyConn)
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

func (t *ClientTest) writeSOCKSReply(s socks.ReplyCode, w io.Writer) {
	t.T().Helper()

	reply := socks.NewReply(s, nil)
	t.Require().NoError(reply.Write(w))
}

func (t *ClientTest) readSOCKSRequest(r io.Reader) *socks.Request {
	t.T().Helper()

	req, err := socks.ReadRequest(bufio.NewReader(r))
	t.Require().NoError(err)

	return req
}
