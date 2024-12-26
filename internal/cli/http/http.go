package http

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/url"

	"github.com/cerfical/socks2http/internal/addr"
)

func Connect(proxyConn net.Conn, destAddr *addr.Addr) error {
	if destAddr.Scheme == addr.HTTP {
		// with plain HTTP no preliminary connection is needed
		return nil
	}

	// send HTTP CONNECT request
	connReq := http.Request{
		Method: http.MethodConnect,
		URL: &url.URL{
			Host: destAddr.Host(),
		},
	}
	if err := connReq.Write(proxyConn); err != nil {
		return err
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), &connReq)
	if err != nil {
		return err
	}

	// ignore the Close() errors
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		code, msg := resp.StatusCode, http.StatusText(resp.StatusCode)
		return fmt.Errorf("HTTP client: %v %v", code, msg)
	}
	return nil
}
