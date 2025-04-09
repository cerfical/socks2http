package proxy_test

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy"
	"github.com/cerfical/socks2http/socks"
	"github.com/stretchr/testify/suite"
)

func TestConnector(t *testing.T) {
	suite.Run(t, new(ConnectorTest))
}

type ConnectorTest struct {
	suite.Suite
}

func (t *ConnectorTest) TestConnect() {
	t.Run("makes a CONNECT request to a SOCKS proxy", func() {
		// Use a raw IPv4 address, as SOCKS4 doesn't support domain names
		dstHost := addr.NewHost("127.0.0.1", 9999)
		proxyServerConn := t.proxyConnect(addr.SOCKS4, dstHost)

		// Assert that the proxy server receives the correct request
		req, err := socks.ReadRequest(bufio.NewReader(proxyServerConn))
		t.Require().NoError(err)

		t.Equal(socks.V4, req.Version)
		t.Equal(dstHost, &req.Host)
		t.Equal(socks.Connect, req.Command)

		// Write a SOCKS Granted reply so the client can proceed
		t.Require().NoError(socks.Granted.Write(proxyServerConn))
	})

	t.Run("make a CONNECT request to an HTTP proxy", func() {
		dstHost := addr.NewHost("localhost", 9999)
		proxyServerConn := t.proxyConnect(addr.HTTP, dstHost)

		// Assert that the proxy server receives the correct request
		req, err := http.ReadRequest(bufio.NewReader(proxyServerConn))
		t.Require().NoError(err)

		t.Equal(dstHost.String(), req.Host)
		t.Equal(http.MethodConnect, req.Method)

		// Write a 200 OK response
		resp := httptest.NewRecorder()
		resp.WriteHeader(http.StatusOK)

		t.Require().NoError(resp.Result().Write(proxyServerConn))
	})
}

func (t *ConnectorTest) proxyConnect(proto string, dstHost *addr.Host) (proxyServerConn net.Conn) {
	proxyClientConn, proxyServerConn := net.Pipe()
	t.T().Cleanup(func() {
		proxyClientConn.Close()
		proxyServerConn.Close()
	})

	connector, err := proxy.NewConnector(proto)
	t.Require().NoError(err)

	errChan := make(chan error)
	go func() {
		errChan <- connector.Connect(proxyClientConn, dstHost)
	}()

	// Wait for the client to finish
	t.T().Cleanup(func() {
		t.Require().NoError(<-errChan)
	})

	return proxyServerConn
}
