package simulator

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// server is a common interface for connection requests servers.
type server interface {
	listenPort() int
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
}

// crHandlerFn is a function that handles connection requests.
type crHandlerFn func(context.Context) error

func newHTTPServer(h crHandlerFn) (server, error) {
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
	}
	mux.HandleFunc("/cwmp", s.handleConnectionRequest)
	go func() {
		if err := s.httpServer.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("Server error")
		}
	}()

	return s, nil
}

func (s *httpServer) handleConnectionRequest(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Received HTTP connection request")
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
	fmt.Println(ip, port)

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
			_, addr, err := listener.ReadFromUDP(buf[:])
			if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
				log.Error().Err(err).Msg("Error reading UDP connection")
				continue
			}
			if addr == nil {
				continue
			}

			log.Info().Str("addr", addr.String()).Msg("Accepted UDP connection request")
			if err := h(ctx); err != nil {
				log.Error().Err(err).Msg("Failed to handle connection request")
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
