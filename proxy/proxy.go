package proxy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/cerfical/socks2http/addr"
)

func New(d Dialer) Proxy {
	return &proxy{d}
}

type Proxy interface {
	ForwardHTTP(ctx context.Context, r *http.Request, dstHost *addr.Host) (*http.Response, error)
	OpenTunnel(ctx context.Context, srcConn net.Conn, dstHost *addr.Host) (done <-chan error, err error)
}

type proxy struct {
	dialer Dialer
}

func (p *proxy) ForwardHTTP(ctx context.Context, r *http.Request, dstHost *addr.Host) (*http.Response, error) {
	dstConn, err := p.dialer.Dial(ctx, dstHost)
	if err != nil {
		return nil, fmt.Errorf("dial %v: %w", dstHost, err)
	}
	defer dstConn.Close()

	if err := r.Write(dstConn); err != nil {
		return nil, fmt.Errorf("write request: %w", err)
	}

	resp, err := http.ReadResponse(bufio.NewReader(dstConn), r)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	return resp, nil
}

func (p *proxy) OpenTunnel(ctx context.Context, srcConn net.Conn, dstHost *addr.Host) (<-chan error, error) {
	dstConn, err := p.dialer.Dial(ctx, dstHost)
	if err != nil {
		return nil, fmt.Errorf("dial %v: %w", dstHost, err)
	}

	errChan := make(chan error)
	go transfer(dstConn, srcConn, errChan)
	go transfer(srcConn, dstConn, errChan)

	done := make(chan error, 1)
	go func() {
		defer dstConn.Close()

		// Wait for both transfers to finish
		var firstErr error
		for range 2 {
			if err := <-errChan; err != nil && firstErr == nil {
				firstErr = err
			}
		}
		done <- firstErr
	}()
	return done, nil
}

func transfer(dest io.Writer, src io.Reader, errChan chan<- error) {
	_, err := io.Copy(dest, src)
	errChan <- err
}
