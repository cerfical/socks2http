package client

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/cerfical/socks2http/internal/proxy/addr"
)

type HTTPClient struct{}

func (c *HTTPClient) Connect(proxyConn net.Conn, dstAddr *addr.Addr) error {
	connReq, err := http.NewRequest(http.MethodConnect, "", nil)
	if err != nil {
		return err
	}
	connReq.Host = dstAddr.String()

	if err := connReq.WriteProxy(proxyConn); err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), connReq)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		code, msg := resp.StatusCode, http.StatusText(resp.StatusCode)
		return fmt.Errorf("connection rejected: %v %v", code, msg)
	}
	return nil
}
