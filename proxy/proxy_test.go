package proxy_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy"
	"github.com/stretchr/testify/suite"
)

type ProxyTest struct {
	suite.Suite
}

func (t *ProxyTest) startProxyServer(proto string) *proxy.Server {
	t.T().Helper()

	server, err := proxy.NewServer(
		proxy.WithListenAddr(addr.New(proto, "localhost", 0)),
	)
	t.Require().NoError(err)

	t.Require().NoError(server.Start(context.Background()))
	t.T().Cleanup(func() { server.Stop() })

	go func() {
		server.Serve(context.Background())
	}()

	return server
}

func (t *ProxyTest) startHTTPEchoServer() (serverHost string) {
	t.T().Helper()

	server := httptest.NewServer(http.HandlerFunc(t.echoHTTP))
	t.T().Cleanup(server.Close)

	url, err := url.Parse(server.URL)
	t.Require().NoError(err)

	return url.Host
}

func (t *ProxyTest) startHTTPSEchoServer() (serverHost string) {
	t.T().Helper()

	server := httptest.NewTLSServer(http.HandlerFunc(t.echoHTTP))
	t.T().Cleanup(server.Close)

	url, err := url.Parse(server.URL)
	t.Require().NoError(err)

	return url.Host
}

func (t *ProxyTest) echoHTTP(w http.ResponseWriter, r *http.Request) {
	t.T().Helper()

	bytes, _ := io.ReadAll(r.Body)
	w.Write(bytes)
}
