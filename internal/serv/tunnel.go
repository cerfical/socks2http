package serv

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"
)

func tunnel(clientConn, servConn net.Conn) <-chan error {
	errChan := make(chan error)

	okResp := http.Response{StatusCode: http.StatusOK, ProtoMajor: 1, ProtoMinor: 1}
	if err := okResp.Write(clientConn); err != nil {
		errChan <- err
		return errChan
	}

	go func() {
		defer close(errChan)
		cli2Serv := startTransfer(servConn, clientConn)
		serv2Cli := startTransfer(clientConn, servConn)

		for cli2Serv != nil || serv2Cli != nil {
			select {
			case err, ok := <-cli2Serv:
				if ok {
					errChan <- fmt.Errorf("client to server transfer: %w", err)
				} else {
					cli2Serv = nil
				}
			case err, ok := <-serv2Cli:
				if ok {
					errChan <- fmt.Errorf("server to client transfer: %w", err)
				} else {
					serv2Cli = nil
				}
			}
		}
	}()
	return errChan
}

func startTransfer(dest, src net.Conn) <-chan error {
	errChan := make(chan error)
	resetConn := func(conn net.Conn) {
		// when a transfer error occurs, deadlines preemptively terminate Read() and Write() calls
		// to prevent goroutines from being blocked indefinitely
		if err := conn.SetDeadline(time.Now()); err != nil {
			errChan <- fmt.Errorf("aborting: %w", err)
		}

		// also reset participating TCP connections
		if err := conn.(*net.TCPConn).SetLinger(0); err != nil {
			errChan <- fmt.Errorf("connection reset: %w", err)
		}
	}

	go func() {
		defer close(errChan)
		if err := transfer(dest, src); err != nil && !errors.Is(err, os.ErrDeadlineExceeded) {
			errChan <- err
			resetConn(dest)
			resetConn(src)
		}
	}()
	return errChan
}

func transfer(dest io.Writer, src io.Reader) error {
	buf := make([]byte, 1024)
	for isEof := false; !isEof; {
		n, err := src.Read(buf)
		if err != nil {
			if err != io.EOF {
				return err
			}
			isEof = true
		}

		if _, err = dest.Write(buf[:n]); err != nil {
			return err
		}
	}
	return nil
}
