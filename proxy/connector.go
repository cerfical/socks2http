package proxy

import (
	"bufio"
	"fmt"
	"net"
	"net/http"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks"
)

func NewConnector(proto string) (Connector, error) {
	switch proto {
	case addr.SOCKS4:
		return ConnectorFunc(connectSOCKS), nil
	case addr.HTTP:
		return ConnectorFunc(connectHTTP), nil
	default:
		return nil, fmt.Errorf("unsupported protocol scheme: %v", proto)
	}
}

type Connector interface {
	Connect(proxyConn net.Conn, dstHost *addr.Host) error
}

type ConnectorFunc func(net.Conn, *addr.Host) error

func (f ConnectorFunc) Connect(proxyConn net.Conn, dstHost *addr.Host) error {
	return f(proxyConn, dstHost)
}

func connectSOCKS(proxyConn net.Conn, h *addr.Host) error {
	connReq := socks.NewRequest(socks.V4, socks.Connect, h)
	if err := connReq.Write(proxyConn); err != nil {
		return fmt.Errorf("SOCKS CONNECT: %w", err)
	}

	reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
	if err != nil {
		return fmt.Errorf("SOCKS CONNECT reply: %w", err)
	}

	if reply != socks.Granted {
		return fmt.Errorf("SOCKS CONNECT rejected: %v", reply)
	}
	return nil
}

func connectHTTP(proxyConn net.Conn, h *addr.Host) error {
	connReq, err := http.NewRequest(http.MethodConnect, "", nil)
	if err != nil {
		return err
	}
	connReq.Host = h.String()

	if err := connReq.WriteProxy(proxyConn); err != nil {
		return fmt.Errorf("HTTP CONNECT: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(proxyConn), connReq)
	if err != nil {
		return fmt.Errorf("HTTP CONNECT response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		code, msg := resp.StatusCode, http.StatusText(resp.StatusCode)
		return fmt.Errorf("HTTP CONNECT rejected: %v %v", code, msg)
	}

	return nil
}
