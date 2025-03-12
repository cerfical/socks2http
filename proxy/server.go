package proxy

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/cerfical/socks2http/addr"
	"github.com/cerfical/socks2http/log"
	"github.com/cerfical/socks2http/socks"
)

func NewServer(servAddr *addr.Addr, timeout time.Duration, proxy *Client, l *log.Logger) (*Server, error) {
	s := Server{addr: servAddr, timeout: timeout, proxy: proxy, log: l}
	switch servAddr.Scheme {
	case addr.HTTP:
		s.newHandler = func(h handlerConfig) requestHandler { return &httpHandler{h, nil} }
	case addr.SOCKS4:
		s.newHandler = func(h handlerConfig) requestHandler { return &socksHandler{h, nil} }
	default:
		return nil, fmt.Errorf("unsupported server protocol scheme %v", servAddr.Scheme)
	}
	return &s, nil
}

type Server struct {
	newHandler func(handlerConfig) requestHandler
	addr       *addr.Addr
	timeout    time.Duration
	proxy      *Client
	log        *log.Logger
	numConn    int
}

func (s *Server) Run(ctx context.Context) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.addr.Host())
	if err != nil {
		return err
	}

	for {
		l := log.New(
			log.WithLogger(s.log),
			log.WithFields(log.Fields{"id": s.numConn}),
		)
		s.numConn++

		cliConn, err := listener.Accept()
		if err != nil {
			l.Error("opening a client connection", err)
			continue
		}

		go func() {
			defer func() {
				if closeErr := cliConn.Close(); closeErr != nil {
					l.Error("closing a client connection", closeErr)
				}
			}()

			ctx := ctx
			if s.timeout != 0 {
				c, cancel := context.WithTimeout(ctx, s.timeout)
				defer cancel()
				ctx = c
			}
			s.handleConn(ctx, cliConn, l)
		}()
	}
}

func (s *Server) handleConn(ctx context.Context, cliConn net.Conn, log *log.Logger) {
	h := s.newHandler(handlerConfig{cliConn, bufio.NewReader(cliConn), log, s.proxy})

	destAddr, err := h.parseRequest()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			log.Error(err.Error(), nil)
		}
		return
	}
	defer h.close()

	h.dumpRequest()

	servConn, err := s.proxy.Open(ctx, destAddr)
	if err != nil {
		log.Error("opening a server connection", err)
		h.reject()
		return
	}
	defer func() {
		if closeErr := servConn.Close(); closeErr != nil {
			log.Error("closing a server connection", closeErr)
		}
	}()

	h.grant(servConn)
}

type requestHandler interface {
	parseRequest() (*addr.Addr, error)
	dumpRequest()

	reject()
	grant(servConn net.Conn)

	close()
}

type handlerConfig struct {
	cliConn net.Conn
	cliBufr *bufio.Reader
	log     *log.Logger
	proxy   *Client
}

type socksHandler struct {
	handlerConfig
	destAddr *addr.Addr
}

func (h *socksHandler) parseRequest() (*addr.Addr, error) {
	req, err := socks.ReadRequest(h.cliBufr)
	if err != nil {
		return nil, fmt.Errorf("parsing socks request: %v", err)
	}

	h.destAddr = &addr.Addr{Hostname: req.DestIP.String(), Port: req.DestPort}
	return h.destAddr, nil
}

func (h *socksHandler) dumpRequest() {
	h.log.Info("incoming socks request", log.Fields{
		"command": "CONNECT",
		"host":    h.destAddr.Host(),
	})
}

func (h *socksHandler) grant(servConn net.Conn) {
	if err := socks.WriteGrant(h.cliConn); err != nil {
		h.log.Error("writing socks reply", err)
		return
	}

	if err := tunnel(h.cliBufr, h.cliConn, servConn); err != nil {
		h.log.Error("socks tunnel", err)
	}
}

func (h *socksHandler) reject() {
	if err := socks.WriteReject(h.cliConn); err != nil {
		h.log.Error("writing socks reply", err)
	}
}

func (*socksHandler) close() {}

type httpHandler struct {
	handlerConfig
	*http.Request
}

func (h *httpHandler) parseRequest() (*addr.Addr, error) {
	req, err := http.ReadRequest(h.cliBufr)
	if err != nil {
		return nil, fmt.Errorf("parsing http request: %w", err)
	}

	destAddr, err := addrFromURL(req.URL)
	if err != nil {
		return nil, fmt.Errorf("parsing http request: %w", err)
	}

	h.Request = req
	return destAddr, nil
}

func addrFromURL(url *url.URL) (*addr.Addr, error) {
	port := url.Port()
	if port == "" {
		p, err := net.LookupPort("tcp", url.Scheme)
		if err != nil {
			return nil, err
		}
		port = strconv.Itoa(p)
	}

	p, err := addr.ParsePort(port)
	if err != nil {
		return nil, err
	}

	return &addr.Addr{
		Scheme:   url.Scheme,
		Hostname: url.Hostname(),
		Port:     p,
	}, nil
}

func (h *httpHandler) dumpRequest() {
	h.log.Info("incoming http request", log.Fields{
		"method": h.Method,
		"uri":    h.RequestURI,
		"proto":  h.Proto,
	})
}

func (h *httpHandler) grant(servConn net.Conn) {
	if h.Method == http.MethodConnect {
		h.httpWriteReply(true)
		if err := tunnel(h.cliBufr, h.cliConn, servConn); err != nil {
			h.log.Error("http tunnel", err)
		}
	} else {
		if err := h.forward(servConn); err != nil {
			h.log.Error("http forwarding", err)
		}
	}
}

func (h *httpHandler) forward(servConn net.Conn) error {
	if h.proxy.addr.Scheme == addr.HTTP {
		if err := h.WriteProxy(servConn); err != nil {
			return err
		}
	} else {
		if err := h.Write(servConn); err != nil {
			return err
		}
	}

	if _, err := io.Copy(h.cliConn, servConn); err != nil {
		return err
	}
	return nil
}

func (h *httpHandler) reject() {
	h.httpWriteReply(false)
}

func (h *httpHandler) httpWriteReply(ok bool) {
	if h.Method == http.MethodConnect {
		resp := http.Response{ProtoMajor: 1, ProtoMinor: 1}
		if ok {
			resp.StatusCode = http.StatusOK
		} else {
			resp.StatusCode = http.StatusForbidden
		}

		if err := resp.Write(h.cliConn); err != nil {
			h.log.Error("writing http response", err)
		}
	}
}

func (h *httpHandler) close() {
	h.Body.Close()
}

func tunnel(r *bufio.Reader, cliConn, servConn net.Conn) error {
	errChan := make(chan error)
	go transfer(servConn, r, errChan)
	go transfer(cliConn, servConn, errChan)

	for err := range errChan {
		if err != nil {
			resetConn(cliConn)
			resetConn(servConn)
			return err
		}
	}
	return nil
}

func transfer(dest io.Writer, src io.Reader, errChan chan<- error) {
	if _, err := io.Copy(dest, src); !errors.Is(err, net.ErrClosed) {
		errChan <- err
	}
}

func resetConn(conn net.Conn) {
	conn.(*net.TCPConn).SetLinger(0)
}
