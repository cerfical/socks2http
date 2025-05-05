package proxy

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/cerfical/socks2http/proxy/addr"
)

func New(d Dialer) Proxy {
	return &proxy{d}
}

type Proxy interface {
	ForwardHTTP(ctx context.Context, r *http.Request, dstHost *addr.Addr) (*http.Response, error)
	OpenTunnel(ctx context.Context, srcConn net.Conn, dstHost *addr.Addr) (done <-chan error, err error)
}

type proxy struct {
	dialer Dialer
}

func (p *proxy) ForwardHTTP(ctx context.Context, r *http.Request, dstHost *addr.Addr) (*http.Response, error) {
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

func (p *proxy) OpenTunnel(ctx context.Context, srcConn net.Conn, dstHost *addr.Addr) (<-chan error, error) {
	dstConn, err := p.dialer.Dial(ctx, dstHost)
	if err != nil {
		return nil, fmt.Errorf("dial %v: %w", dstHost, err)
	}

	dst2SrcDone, dst2SrcStop := transfer(dstConn, srcConn)
	src2DstDone, src2DstStop := transfer(srcConn, dstConn)

	errChan := make(chan error, 1)
	go func() {
		defer dstConn.Close()

		// The first side that finishes the transfer stops the other side,
		// so that there are no hanging connections
		select {
		case err := <-dst2SrcDone:
			src2DstStop()
			errChan <- err
		case err := <-src2DstDone:
			dst2SrcStop()
			errChan <- err
		}
	}()

	return errChan, nil
}

func transfer(dst net.Conn, src net.Conn) (done <-chan error, stop func()) {
	errChan := make(chan error, 1)
	go func() {
		_, err := io.Copy(dst, src)
		errChan <- err
	}()

	return errChan, func() {
		// Stop the ongoing read operation and wait for it to return
		src.SetReadDeadline(time.Now())
		<-errChan
	}
}
