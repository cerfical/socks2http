package client

import (
	"bufio"
	"fmt"
	"net"

	"github.com/cerfical/socks2http/internal/proxy/addr"
	"github.com/cerfical/socks2http/internal/proxy/socks"
)

type SOCKSClient struct {
	Version      socks.Version
	ResolveNames bool
}

func (c *SOCKSClient) Connect(proxyConn net.Conn, dstAddr *addr.Addr) error {
	if c.ResolveNames {
		ip, err := net.ResolveIPAddr("ip", dstAddr.Host)
		if err != nil {
			return fmt.Errorf("resolve destination: %w", err)
		}
		dstAddr = addr.NewAddr(ip.String(), dstAddr.Port)
	}

	bufr := bufio.NewReader(proxyConn)
	if c.Version == socks.V5 {
		if err := c.auth(proxyConn, bufr); err != nil {
			return err
		}
	}

	connReq := socks.Request{
		Version: c.Version,
		Command: socks.CommandConnect,
		DstAddr: *dstAddr,
	}
	if err := connReq.Write(proxyConn); err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	reply, err := socks.ReadReply(bufr)
	if err != nil {
		return fmt.Errorf("read reply: %w", err)
	}
	if reply.Status != socks.StatusGranted {
		return fmt.Errorf("connection rejected: %v", reply.Status)
	}

	return nil
}

func (c *SOCKSClient) auth(proxyConn net.Conn, proxyRead *bufio.Reader) error {
	greet := socks.Greeting{
		Version: c.Version,
		Auth:    []socks.Auth{socks.AuthNone},
	}
	if err := greet.Write(proxyConn); err != nil {
		return fmt.Errorf("write greeting: %w", err)
	}

	greetRep, err := socks.ReadGreetingReply(proxyRead)
	if err != nil {
		return fmt.Errorf("read greeting reply: %w", err)
	}

	switch greetRep.Auth {
	case socks.AuthNone:
		return nil
	case socks.AuthNotAcceptable:
		return fmt.Errorf("no acceptable auth method was selected")
	default:
		return fmt.Errorf("unsupported auth method: %v", greetRep.Auth)
	}
}
