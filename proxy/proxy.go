package proxy

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
)

var Direct Dialer = DialerFunc(directDial)

func New(ops *Options) (Proxy, error) {
	switch scheme := ops.Addr.Scheme; scheme {
	case addr.SOCKS4:
		return &socksProxy{ops}, nil
	case addr.HTTP:
		return &httpProxy{ops}, nil
	default:
		return nil, fmt.Errorf("unsupported protocol scheme: %v", scheme)
	}
}

func directDial(ctx context.Context, h *addr.Host) (net.Conn, error) {
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", h.String())
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func transfer(dest io.Writer, src io.Reader, errChan chan<- error) {
	if _, err := io.Copy(dest, src); !errors.Is(err, net.ErrClosed) {
		errChan <- err
	}
}

type Options struct {
	Addr   addr.Addr
	Dialer Dialer

	Log *log.Logger
}

type Proxy interface {
	Connect(context.Context, *addr.Host) (net.Conn, error)
	Serve(context.Context, net.Conn) error
}

type Dialer interface {
	Dial(context.Context, *addr.Host) (net.Conn, error)
}

type DialerFunc func(context.Context, *addr.Host) (net.Conn, error)

func (f DialerFunc) Dial(ctx context.Context, h *addr.Host) (net.Conn, error) {
	return f(ctx, h)
}
