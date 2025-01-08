package simulator

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// server is a common interface for connection requests servers.
type server interface {
	url() string
	stop(context.Context) error
}

//
// HTTP server
//

// httpServer implements an HTTP connection request server.
type httpServer struct {
	httpServer *http.Server
	handler    crHandlerFn
	port       int
	logger     zerolog.Logger
}

// crHandlerFn is a function that handles connection requests.
type crHandlerFn func(context.Context) error

func newHTTPServer(h crHandlerFn, logger zerolog.Logger) (server, error) {
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
			logger.Error().Err(err).Msg("Server error")
		}
	}()

	return s, nil
}

func (s *httpServer) handleConnectionRequest(w http.ResponseWriter, r *http.Request) {
	s.logger.Info().Msg("Received HTTP connection request")
	err := s.handler(r.Context())
	if errors.Is(err, errServiceUnavailable) {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
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
// No-op server
//

// noopServer is a no-op connection request server.
type noopServer struct{}

func newNoopServer() server {
	return noopServer{}
}

func (n noopServer) url() string {
	return ""
}

func (n noopServer) stop(_ context.Context) error {
	return nil
}
