package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/cerfical/socks2http/internal/proxy"
	"github.com/cerfical/socks2http/internal/proxy/socks4"
	"github.com/cerfical/socks2http/internal/proxy/socks5"
)

type SOCKSServer struct {
	Dialer   proxy.Dialer
	Tunneler proxy.Tunneler

	Log proxy.Logger
}

func (s *SOCKSServer) ServeSOCKS4(ctx context.Context, l net.Listener) error {
	return s.serve(ctx, l, s.socks4Serve)
}

func (s *SOCKSServer) ServeSOCKS5(ctx context.Context, l net.Listener) error {
	return s.serve(ctx, l, s.socks5Serve)
}

func (s *SOCKSServer) ServeSOCKS(ctx context.Context, l net.Listener) error {
	return s.serve(ctx, l, s.socksServe)
}

type serveConnFunc func(context.Context, *bufio.Reader, net.Conn)

func (s *SOCKSServer) serve(ctx context.Context, l net.Listener, serveConn serveConnFunc) error {
	var activeConns sync.WaitGroup
	go func() {
		for {
			activeConns.Add(1)

			clientConn, err := l.Accept()
			if err != nil {
				activeConns.Done()

				if errors.Is(err, net.ErrClosed) {
					break
				}
				s.Log.Error("Failed to establish a client connection", "error", err)
				continue
			}

			go func() {
				defer func() {
					clientConn.Close()
					activeConns.Done()
				}()

				serveConn(context.Background(), bufio.NewReader(clientConn), clientConn)
			}()
		}
	}()

	// Wait for server shutdown
	<-ctx.Done()
	err := l.Close()
	activeConns.Wait()

	if err != nil {
		return fmt.Errorf("close listener: %w", err)
	}
	return nil
}

func (s *SOCKSServer) socks5Serve(ctx context.Context, clientRead *bufio.Reader, clientConn net.Conn) {
	greet, err := socks5.ReadGreeting(clientRead)
	if err != nil {
		s.serverError(fmt.Errorf("read greeting: %w", err))
		return
	}

	greetReply := socks5.GreetingReply{AuthMethod: selectSOCKS5AuthMethod(greet.AuthMethods)}
	if err := greetReply.Write(clientConn); err != nil {
		s.serverError(fmt.Errorf("write greeting reply: %w", err))
		return
	}

	req, err := socks5.ReadRequest(clientRead)
	if err != nil {
		s.serverError(fmt.Errorf("read request: %w", err))
		return
	}

	switch req.Command {
	case socks5.CommandConnect:
		dstConn, err := s.Dialer.Dial(ctx, &req.DstAddr)
		if err != nil {
			s.socks5Status(clientConn, req, socks5.StatusHostUnreachable, fmt.Errorf("connect to destination: %w", err))
			return
		}
		defer dstConn.Close()

		if !s.socks5Status(clientConn, req, socks5.StatusOK, nil) {
			return
		}
		if err := s.Tunneler.Tunnel(ctx, clientConn, dstConn); err != nil {
			s.serverError(fmt.Errorf("proxy tunnel: %w", err))
			return
		}
	default:
		s.socks5Status(clientConn, req, socks5.StatusCommandNotSupported, nil)
		return
	}
}

func (s *SOCKSServer) socks5Status(clientConn net.Conn, r *socks5.Request, status socks5.Status, err error) bool {
	msg := fmt.Sprintf("%v %v", r.Command, &r.DstAddr)
	fields := []any{
		"status", status,
		"proto", "SOCKS5",
		"client", clientConn.RemoteAddr().String(),
	}

	if err != nil {
		s.Log.Error(msg, append(fields,
			"error", err,
		)...)
	} else {
		s.Log.Info(msg, fields...)
	}

	reply := socks5.Reply{
		Status: status,
	}

	if err := reply.Write(clientConn); err != nil {
		s.serverError(fmt.Errorf("write reply: %w", err))
		return false
	}
	return true
}

func (s *SOCKSServer) socks4Serve(ctx context.Context, clientRead *bufio.Reader, clientConn net.Conn) {
	req, err := socks4.ReadRequest(clientRead)
	if err != nil {
		s.serverError(fmt.Errorf("read request: %w", err))
		return
	}

	switch req.Command {
	case socks4.CommandConnect:
		dstConn, err := s.Dialer.Dial(ctx, &req.DstAddr)
		if err != nil {
			s.socks4Status(clientConn, req, socks4.StatusRejectedOrFailed, fmt.Errorf("connect to destination: %w", err))
			return
		}
		defer dstConn.Close()

		if !s.socks4Status(clientConn, req, socks4.StatusGranted, nil) {
			return
		}

		if err := s.Tunneler.Tunnel(ctx, clientConn, dstConn); err != nil {
			s.serverError(fmt.Errorf("proxy tunnel: %w", err))
			return
		}
	default:
		s.socks4Status(clientConn, req, socks4.StatusRejectedOrFailed, nil)
		return
	}
}

func (s *SOCKSServer) socks4Status(clientConn net.Conn, r *socks4.Request, status socks4.Status, err error) bool {
	msg := fmt.Sprintf("%v %v", r.Command, &r.DstAddr)
	fields := []any{
		"status", status,
		"proto", "SOCKS4",
		"client", clientConn.RemoteAddr().String(),
	}

	if err != nil {
		s.Log.Error(msg, append(fields,
			"error", err,
		)...)
	} else {
		s.Log.Info(msg, fields...)
	}

	reply := socks4.Reply{
		Status: status,
	}

	if err := reply.Write(clientConn); err != nil {
		s.serverError(fmt.Errorf("write reply: %w", err))
		return false
	}
	return true
}

func (s *SOCKSServer) socksServe(ctx context.Context, clientRead *bufio.Reader, clientConn net.Conn) {
	version, err := clientRead.Peek(1)
	if err != nil {
		s.serverError(fmt.Errorf("read version: %w", err))
		return
	}

	switch version[0] {
	case socks4.VersionCode:
		s.socks4Serve(ctx, clientRead, clientConn)
	case socks5.VersionCode:
		s.socks5Serve(ctx, clientRead, clientConn)
	default:
		s.serverError(fmt.Errorf("invalid version (%#02x)", version[0]))
	}
}

func (s *SOCKSServer) serverError(err error) {
	// Ignore errors caused by client closing the connection
	if errors.Is(err, io.EOF) {
		return
	}
	s.Log.Error("SOCKS failure", "error", err)
}

func selectSOCKS5AuthMethod(methods []socks5.AuthMethod) socks5.AuthMethod {
	for _, m := range methods {
		if m == socks5.AuthNone {
			return m
		}
	}
	return socks5.AuthNotAcceptable
}
