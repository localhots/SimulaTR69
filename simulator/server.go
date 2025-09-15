package simulator

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/localhots/blip"
	"github.com/localhots/blip/noctx/log"
)

// server is a common interface for connection requests servers.
type server interface {
	listenPort() int
	url() string
	stop(context.Context) error
}

type crParams struct {
	ts  string // Timestamp
	id  string // Message ID
	un  string // Username
	cn  string // Cnonce
	sig string // Signature
}

// crHandlerFn is a function that handles connection requests.
type crHandlerFn func(context.Context, crParams) error

//
// HTTP server
//

// httpServer implements an HTTP connection request server.
type httpServer struct {
	httpServer *http.Server
	handler    crHandlerFn
	port       int
	logger     *blip.Logger
}

func newHTTPServer(h crHandlerFn, logger *blip.Logger) (server, error) {
	var err error
	if Config.Host == "" {
		Config.Host, err = getIP()
		if err != nil {
			return nil, fmt.Errorf("get ip address: %w", err)
		}
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", Config.Host, Config.Port))
	if err != nil {
		return nil, fmt.Errorf("create TCP listener: %w", err)
	}

	// Config.Port can be set to 0 in order to bind to a random available port.
	port := listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	s := &httpServer{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", Config.Host, port),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
		},
		handler: h,
		port:    port,
		logger:  logger,
	}
	mux.HandleFunc("/cwmp", s.handleConnectionRequest)
	go func() {
		if err := s.httpServer.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			logger.Error(context.TODO(), "Server error", log.Cause(err))
		}
	}()

	return s, nil
}

func (s *httpServer) handleConnectionRequest(w http.ResponseWriter, r *http.Request) {
	s.logger.Info(context.TODO(), "Received HTTP connection request", log.F{
		"remote_addr": r.RemoteAddr,
		"method":      r.Method,
		"url":         r.URL.String(),
	})
	params := parseCrParams(r.URL)
	err := s.handler(r.Context(), params)
	if errors.Is(err, errServiceUnavailable) {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if errors.Is(err, errForbidden) {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (s *httpServer) listenPort() int {
	return s.port
}

func (s *httpServer) url() string {
	return fmt.Sprintf("http://%s:%d/cwmp", Config.Host, s.port)
}

func (s *httpServer) stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func getIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}
	return "0.0.0.0", nil
}

//
// UDP server
//

type udpServer struct {
	ip       string
	port     int
	listener *net.UDPConn
	handler  crHandlerFn
}

func newUDPServer(ctx context.Context, port int, h crHandlerFn) (server, error) {
	ip, err := getIP()
	if err != nil {
		return nil, fmt.Errorf("get ip address: %w", err)
	}

	listener, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.ParseIP(ip),
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("create TCP listener: %w", err)
	}

	go func() {
		var buf [1024]byte
		for {
			n, addr, err := listener.ReadFromUDP(buf[:])
			if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
				log.Error("Error reading UDP connection", log.Cause(err))
				continue
			}
			if addr == nil {
				continue
			}

			log.Info("Accepted UDP connection request", log.F{"addr": addr.String()})
			if n == 0 {
				log.Warn("Received empty UDP message")
				continue
			}
			u, err := parseUDPMessage(buf[:])
			if err != nil {
				log.Warn("Failed to parse UDP message", log.Cause(err))
				continue
			}
			params := parseCrParams(u)

			if err := h(ctx, params); err != nil {
				log.Error("Failed to handle connection request", log.Cause(err))
			}
		}
	}()

	return &udpServer{
		ip:       ip,
		port:     port,
		listener: listener,
		handler:  h,
	}, nil
}

func (s *udpServer) listenPort() int {
	return s.port
}

func (s *udpServer) url() string {
	return fmt.Sprintf("%s:%d", s.ip, s.port)
}

func (s *udpServer) stop(_ context.Context) error {
	// Safe to ignore any errors here
	_ = s.listener.Close()
	return nil
}

//
// No-op server
//

// noopServer is a no-op connection request server.
type noopServer struct{}

func newNoopServer() server {
	return noopServer{}
}

func (s noopServer) listenPort() int {
	return 0
}

func (s noopServer) url() string {
	return ""
}

func (s noopServer) stop(_ context.Context) error {
	return nil
}

//
// Utils
//

func parseUDPMessage(buf []byte) (*url.URL, error) {
	tokens := strings.Fields(string(buf))
	if len(tokens) < 3 || tokens[0] != "GET" || !strings.HasPrefix(tokens[2], "HTTP/") {
		return nil, errors.New("invalid UDP message format")
	}
	u, err := url.Parse(tokens[1])
	if err != nil {
		return nil, fmt.Errorf("parse UDP message URL: %w", err)
	}
	return u, nil
}

func parseCrParams(u *url.URL) crParams {
	q := u.Query()
	return crParams{
		ts:  q.Get("ts"),
		id:  q.Get("id"),
		un:  q.Get("un"),
		cn:  q.Get("cn"),
		sig: q.Get("sig"),
	}
}
