package prox

import (
	"bufio"
	"io"
	"net"

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/socks"
)

type socksHandler struct{}

func (socksHandler) readRequest(r *bufio.Reader) (request, error) {
	req, err := socks.ReadRequest(r)
	if err != nil {
		return nil, err
	}

	return &socksRequest{req, addr.Addr{
		Hostname: req.DestIP.String(),
		Port:     req.DestPort,
	}}, nil
}

type socksRequest struct {
	*socks.Request
	dest addr.Addr
}

func (r *socksRequest) destAddr() *addr.Addr {
	return &r.dest
}

func (r *socksRequest) logAttrs(log *log.Logger) *log.Logger {
	return log.WithAttrs(
		"command", "CONNECT",
		"host", r.dest.Host(),
	)
}

func (*socksRequest) writeGrant(w io.Writer) error {
	rep := socks.Reply{Code: socks.RequestGranted}
	return rep.Write(w)
}

func (*socksRequest) writeReject(w io.Writer) error {
	rep := socks.Reply{Code: socks.RequestRejectedOrFailed}
	return rep.Write(w)
}

func (r *socksRequest) do(cliConn, servConn net.Conn, _ string) error {
	return tunnel(cliConn, servConn)
}

func (r *socksRequest) Close() error {
	return nil
}
