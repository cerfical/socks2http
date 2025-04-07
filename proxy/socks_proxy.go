package proxy

import (
	"bufio"
	"context"
	"fmt"
	"net"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/socks"
)

type socksProxy struct {
	*Options
}

func (p *socksProxy) Serve(ctx context.Context, clientConn net.Conn) error {
	req, err := socks.ReadRequest(bufio.NewReader(clientConn))
	if err != nil {
		return fmt.Errorf("parse request: %w", err)
	}

	p.Log.Info("New SOCKS request",
		"version", req.Version.String(),
		"command", req.Command.String(),
		"host", req.Host.String(),
	)

	serverConn, err := p.Dialer.Dial(ctx, &req.Host)
	if err != nil {
		socks.Rejected.Write(clientConn)
		return fmt.Errorf("connect to %v: %w", &req.Host, err)
	}
	defer serverConn.Close()

	if err := socks.Granted.Write(clientConn); err != nil {
		return fmt.Errorf("write reply: %w", err)
	}
	if err := tunnel(clientConn, serverConn); err != nil {
		return fmt.Errorf("proxy tunnel: %w", err)
	}

	return nil
}

func tunnel(clientConn, serverConn net.Conn) error {
	errChan := make(chan error)
	go transfer(serverConn, clientConn, errChan)
	go transfer(clientConn, serverConn, errChan)

	// Wait for both transfers to finish
	var firstErr error
	for err := range errChan {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (p *socksProxy) Connect(ctx context.Context, h *addr.Host) (net.Conn, error) {
	proxyConn, err := p.Dialer.Dial(ctx, &p.Addr.Host)
	if err != nil {
		return nil, fmt.Errorf("connect to proxy %v: %w", &p.Addr.Host, err)
	}

	if err := p.connect(proxyConn, h); err != nil {
		proxyConn.Close()
		return nil, fmt.Errorf("SOCKS CONNECT: %w", err)
	}

	return proxyConn, nil
}

func (p *socksProxy) connect(proxyConn net.Conn, h *addr.Host) error {
	connReq := socks.NewRequest(socks.V4, socks.Connect, h)
	if err := connReq.Write(proxyConn); err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	reply, err := socks.ReadReply(bufio.NewReader(proxyConn))
	if err != nil {
		return fmt.Errorf("parse reply: %w", err)
	}

	if reply != socks.Granted {
		return fmt.Errorf("unexpected reply: %v", reply)
	}
	return nil
}
