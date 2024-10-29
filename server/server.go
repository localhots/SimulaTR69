package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-xmlfmt/xmlfmt"
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/datamodel"
	"github.com/localhots/SimulaTR69/rpc"
	"github.com/localhots/SimulaTR69/server/metrics"
)

// Server is a server that can connect to an ACS and receive connection
// requests.
type Server struct {
	httpServer           *http.Server
	dm                   *datamodel.DataModel
	cookies              http.CookieJar
	informScheduleUpdate chan struct{}
	startedAt            time.Time
	informMux            sync.Mutex
	metrics              *metrics.Metrics
}

// Start starts the server.
func (s *Server) Start(ctx context.Context) error {
	log.Info().Str("server_url", s.URL()).Msg("Starting server")
	log.Info().Str("acs_url", Config.ACSURL).Msg("Connecting to ACS")
	s.startedAt = time.Now()
	go s.periodicInform(ctx)
	return s.httpServer.ListenAndServe()
}

// Stop stops the server.
func (s *Server) Stop(ctx context.Context) error {
	if err := s.dm.SaveState(Config.StateFilePath); err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	return s.httpServer.Shutdown(ctx)
}

// URL returns the URL of the server.
func (s *Server) URL() string {
	return fmt.Sprintf("http://%s:%d/cwmp", Config.Host, Config.Port)
}

// New creates a new server instance.
func New(dm *datamodel.DataModel) *Server {
	mux := http.NewServeMux()
	jar, _ := cookiejar.New(nil)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", Config.Host, Config.Port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	s := &Server{
		httpServer:           httpServer,
		dm:                   dm,
		cookies:              jar,
		informScheduleUpdate: make(chan struct{}, 1),
		metrics:              metrics.NewNoop(),
	}
	mux.HandleFunc("/cwmp", s.handleConnectionRequest)
	s.dm.SetConnectionRequestURL(s.URL())
	return s
}

func NewWithMetrics(dm *datamodel.DataModel, m *metrics.Metrics) *Server {
	s := New(dm)
	s.metrics = m
	return s
}

func (s *Server) handleConnectionRequest(w http.ResponseWriter, r *http.Request) {
	// Simulate downtime
	if s.dm.DownUntil().After(time.Now()) {
		retryAfter := int(time.Until(s.dm.DownUntil()).Seconds())
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
		return
	}

	log.Info().Msg("Received HTTP connection request")
	s.dm.AddEvent(rpc.EventConnectionRequest)
	go s.Inform(context.WithoutCancel(r.Context()))
}

// nolint:gocyclo
func (s *Server) handleEnvelope(env *rpc.EnvelopeDecoder) *rpc.EnvelopeEncoder {
	envID := env.Header.ID.Value
	switch {
	case env.Body.GetRPCMethods != nil:
		return s.handleGetRPCMethods(envID)
	case env.Body.SetParameterValues != nil:
		return s.handleSetParameterValues(envID, env.Body.SetParameterValues)
	case env.Body.GetParameterValues != nil:
		return s.handleGetParameterValues(envID, env.Body.GetParameterValues)
	case env.Body.GetParameterNames != nil:
		return s.handleGetParameterNames(envID, env.Body.GetParameterNames)
	case env.Body.SetParameterAttributes != nil:
		return s.handleSetParameterAttributes(envID, env.Body.SetParameterAttributes)
	case env.Body.GetParameterAttributes != nil:
		return s.handleGetParameterAttributes(envID, env.Body.GetParameterAttributes)
	case env.Body.AddObject != nil:
		return s.handleAddObject(envID, env.Body.AddObject)
	case env.Body.DeleteObject != nil:
		return s.handleDeleteObject(envID, env.Body.DeleteObject)
	case env.Body.Reboot != nil:
		return s.handleReboot(envID, env.Body.Reboot)
	case env.Body.Download != nil:
		return s.handleDownload(envID, env.Body.Download)
	case env.Body.Upload != nil:
		return s.handleUpload(envID, env.Body.Upload)
	case env.Body.FactoryReset != nil:
		return s.handleFactoryReset(envID)
	case env.Body.GetQueuedTransfers != nil:
		return s.handleGetQueuedTransfers(envID)
	case env.Body.GetAllQueuedTransfers != nil:
		return s.handleGetAllQueuedTransfers(envID)
	case env.Body.ScheduleInform != nil:
		return s.handleScheduleInform(envID)
	case env.Body.SetVouchers != nil:
		return s.handleSetVouchers(envID)
	case env.Body.GetOptions != nil:
		return s.handleGetOptions(envID)
	case env.Body.Fault != nil:
		return s.handleFault(envID, env.Body.Fault)
	default:
		log.Warn().Msg("Unknown method")
		return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
	}
}

func (s *Server) handleGetQueuedTransfers(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "GetQueuedTransfers").Msg("Received message")
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}

func (s *Server) handleGetAllQueuedTransfers(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "GetAllQueuedTransfers").Msg("Received message")
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}

func (s *Server) handleScheduleInform(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "ScheduleInform").Msg("Received message")
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}

func (s *Server) handleSetVouchers(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "SetVouchers").Msg("Received message")
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}

func (s *Server) handleGetOptions(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "GetOptions").Msg("Received message")
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}

func (s *Server) handleFault(envID string, r *rpc.FaultPayload) *rpc.EnvelopeEncoder {
	log.Error().
		Str("env_id", envID).
		Str("code", r.Detail.Fault.FaultCode.String()).
		Str("string", r.Detail.Fault.FaultString).
		Msg("ACS fault")
	return nil
}

func (s *Server) pretendOfflineFor(dur time.Duration) {
	downUntil := time.Now().Add(dur)
	s.dm.SetDownUntil(downUntil)
	s.dm.SetPeriodicInformTime(downUntil)
	s.resetInformTimer()
	s.startedAt = downUntil
}

var envelopeID uint64

func newEnvelope() *rpc.EnvelopeEncoder {
	return rpc.NewEnvelope(nextEnvelopeID())
}

func nextEnvelopeID() string {
	id := atomic.AddUint64(&envelopeID, 1)
	return strconv.FormatUint(id, 10)
}

func prettyXML(b []byte) string {
	return strings.TrimSpace(xmlfmt.FormatXML(string(b), "", "    "))
}
