package prox

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

	"github.com/cerfical/socks2http/internal/addr"
	"github.com/cerfical/socks2http/internal/log"
	"github.com/cerfical/socks2http/internal/socks"
)

func NewServer(servAddr *addr.Addr, timeout time.Duration, proxy *Client, log *log.Logger) (*Server, error) {
	s := Server{addr: servAddr, timeout: timeout, proxy: proxy, log: log}
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
	numReq     int
}

func (s *Server) Run(ctx context.Context) error {
	lc := net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", s.addr.Host())
	if err != nil {
		return err
	}

	for {
		log := s.log.WithAttr("id", strconv.Itoa(s.numReq))
		s.numReq++

		cliConn, err := listener.Accept()
		if err != nil {
			log.Errorf("opening a client connection: %v", err)
			continue
		}

		go func() {
			defer func() {
				if closeErr := cliConn.Close(); closeErr != nil {
					log.Errorf("closing a client connection: %v", closeErr)
				}
			}()

			ctx := ctx
			if s.timeout != 0 {
				c, cancel := context.WithTimeout(ctx, s.timeout)
				defer cancel()
				ctx = c
			}
			s.handleConn(ctx, cliConn, log)
		}()
	}
}

func (s *Server) handleConn(ctx context.Context, cliConn net.Conn, log *log.Logger) {
	h := s.newHandler(handlerConfig{cliConn, bufio.NewReader(cliConn), log, s.proxy})

	destAddr, err := h.parseRequest()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			log.Errorf("%v", err)
		}
		return
	}
	defer h.close()

	h.dumpRequest()

	servConn, err := s.proxy.Open(ctx, destAddr)
	if err != nil {
		log.Errorf("opening a server connection: %v", err)
		h.reject()
		return
	}
	defer func() {
		if closeErr := servConn.Close(); closeErr != nil {
			log.Errorf("closing a server connection: %v", closeErr)
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
	h.log.WithAttrs("command", "CONNECT", "host", h.destAddr.Host()).
		Infof("incoming socks request")
}

func (h *socksHandler) grant(servConn net.Conn) {
	if err := socks.WriteGrant(h.cliConn); err != nil {
		h.log.Errorf("writing socks reply: %v", err)
		return
	}

	if err := tunnel(h.cliBufr, h.cliConn, servConn); err != nil {
		h.log.Errorf("socks tunnel: %v", err)
	}
}

func (h *socksHandler) reject() {
	if err := socks.WriteReject(h.cliConn); err != nil {
		h.log.Errorf("writing socks reply: %v", err)
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
	h.log.WithAttrs("method", h.Method, "uri", h.RequestURI, "proto", h.Proto).
		Infof("incoming http request")
}

func (h *httpHandler) grant(servConn net.Conn) {
	if h.Method == http.MethodConnect {
		h.httpWriteReply(true)
		if err := tunnel(h.cliBufr, h.cliConn, servConn); err != nil {
			h.log.Errorf("http tunnel: %v", err)
		}
	} else {
		if err := h.forward(servConn); err != nil {
			h.log.Errorf("http forwarding: %v", err)
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
			h.log.Errorf("writing http response: %v", err)
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
