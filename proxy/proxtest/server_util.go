package proxtest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/proxy/proxserv"
	"github.com/stretchr/testify/require"
)

func StartProxyServer(t *testing.T, proto string) *addr.Host {
	t.Helper()

	server, err := proxserv.New(context.Background(),
		proxserv.WithListenAddr(addr.New(proto, "localhost", 0)),
	)
	require.NoError(t, err)
	t.Cleanup(func() { server.Stop() })

	go func() {
		server.Serve(context.Background())
	}()

	return &server.ListenAddr().Host
}

func StartHTTPEchoServer(t *testing.T) *addr.Host {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(echoHTTP))
	t.Cleanup(server.Close)

	return hostFromURL(t, server.URL)
}

func StartHTTPSEchoServer(t *testing.T) *addr.Host {
	t.Helper()

	server := httptest.NewTLSServer(http.HandlerFunc(echoHTTP))
	t.Cleanup(server.Close)

	return hostFromURL(t, server.URL)
}

func hostFromURL(t *testing.T, s string) *addr.Host {
	t.Helper()

	url, err := url.Parse(s)
	require.NoError(t, err)

	h, err := addr.ParseHost(url.Host)
	require.NoError(t, err)

	return h
}

func echoHTTP(w http.ResponseWriter, r *http.Request) {
	bytes, _ := io.ReadAll(r.Body)
	w.Write(bytes)
}
