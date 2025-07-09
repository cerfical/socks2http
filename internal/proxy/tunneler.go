package proxy

import (
	"context"
	"io"
	"net"
	"time"
)

var DefaultTunneler Tunneler = &defaultTunneler{}

type Tunneler interface {
	Tunnel(ctx context.Context, srcConn, dstConn net.Conn) error
}

type defaultTunneler struct{}

func (t *defaultTunneler) Tunnel(ctx context.Context, srcConn, dstConn net.Conn) error {
	dst2SrcDone, dst2SrcStop := transfer(dstConn, srcConn)
	src2DstDone, src2DstStop := transfer(srcConn, dstConn)

	// The first side that finishes the transfer stops the other side to prevent hanging connections
	select {
	case err := <-dst2SrcDone:
		src2DstStop()
		return err
	case err := <-src2DstDone:
		dst2SrcStop()
		return err
	}
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
