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
	"github.com/cerfical/socks2http/internal/proxy/socks"
)

type SOCKSServer struct {
	Version socks.Version

	Dialer   proxy.Dialer
	Tunneler proxy.Tunneler

	Log proxy.Logger
}

func (s *SOCKSServer) ServeSOCKS(ctx context.Context, l net.Listener) error {
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
				s.serverError(fmt.Errorf("accept connection: %w", err))
				continue
			}

			go func() {
				defer func() {
					clientConn.Close()
					activeConns.Done()
				}()

				s.serve(context.Background(), clientConn)
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

func (s *SOCKSServer) serve(ctx context.Context, clientConn net.Conn) {
	bufr := bufio.NewReader(clientConn)
	if s.Version == socks.V5 || s.Version == 0 {
		s.auth(clientConn, bufr)
	}

	req, err := socks.ReadRequest(bufr)
	if err != nil {
		s.serverError(fmt.Errorf("read request: %w", err))
		return
	}

	switch req.Command {
	case socks.CommandConnect:
		dstConn, err := s.Dialer.Dial(ctx, &req.DstAddr)
		if err != nil {
			s.reply(clientConn, req, socks.StatusHostUnreachable, fmt.Errorf("dial destination: %w", err))
			return
		}
		defer dstConn.Close()

		if !s.reply(clientConn, req, socks.StatusGranted, nil) {
			return
		}

		if err := s.Tunneler.Tunnel(ctx, clientConn, dstConn); err != nil {
			s.serverError(fmt.Errorf("proxy tunnel: %w", err))
			return
		}
	default:
		s.reply(clientConn, req, socks.StatusCommandNotSupported, nil)
		return
	}
}

func (s *SOCKSServer) auth(clientConn net.Conn, clientRead *bufio.Reader) {
	greet, err := socks.ReadGreeting(clientRead)
	if err != nil {
		if errors.Is(err, socks.ErrInvalidVersion) {
			return
		}
		s.serverError(fmt.Errorf("read greeting: %w", err))
		return
	}

	greetReply := socks.GreetingReply{
		Version: greet.Version,
		Auth:    selectSOCKSAuth(greet.Auth),
	}
	if err := greetReply.Write(clientConn); err != nil {
		s.serverError(fmt.Errorf("write greeting reply: %w", err))
		return
	}
}

func (s *SOCKSServer) reply(clientConn net.Conn, r *socks.Request, status socks.Status, err error) bool {
	msg := fmt.Sprintf("%v %v", r.Command, &r.DstAddr)
	fields := []any{
		"status", status,
		"proto", r.Version,
		"client", clientConn.RemoteAddr().String(),
	}

	if err != nil {
		s.Log.Error(msg, append(fields,
			"error", err,
		)...)
	} else {
		s.Log.Info(msg, fields...)
	}

	reply := socks.Reply{
		Version: r.Version,
		Status:  status,
	}
	if err := reply.Write(clientConn); err != nil {
		s.serverError(fmt.Errorf("write reply: %w", err))
		return false
	}
	return true
}

func (s *SOCKSServer) serverError(err error) {
	// Ignore errors caused by client closing the connection
	if errors.Is(err, io.EOF) {
		return
	}
	s.Log.Error("SOCKS failure", "error", err)
}

func selectSOCKSAuth(auth []socks.Auth) socks.Auth {
	for _, a := range auth {
		if a == socks.AuthNone {
			return a
		}
	}
	return socks.AuthNotAcceptable
}
