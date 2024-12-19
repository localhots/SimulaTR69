package simulator

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" //nolint:gosec
	"encoding/hex"
	"errors"
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
	"github.com/localhots/SimulaTR69/simulator/metrics"
)

// Simulator is a TR-069 device simulator.
type Simulator struct {
	httpServer server
	udpServer  server
	dm         *datamodel.DataModel
	cookies    http.CookieJar
	startedAt  time.Time
	envelopeID uint64
	metrics    *metrics.Metrics

	pendingEvents        chan string
	pendingRequests      chan func(*rpc.EnvelopeEncoder)
	informScheduleUpdate chan struct{}
	stop                 chan struct{}
	tasks                chan taskFn
	sessionMux           sync.Mutex
}

var (
	errServiceUnavailable = errors.New("service unavailable")
	errForbidden          = errors.New("forbidden")
)

// New creates a new simulator instance.
func New(dm *datamodel.DataModel) *Simulator {
	jar, _ := cookiejar.New(nil)
	return &Simulator{
		httpServer:           newNoopServer(),
		udpServer:            newNoopServer(),
		dm:                   dm,
		cookies:              jar,
		metrics:              metrics.NewNoop(),
		pendingEvents:        make(chan string, 5),
		pendingRequests:      make(chan func(*rpc.EnvelopeEncoder), 5),
		informScheduleUpdate: make(chan struct{}, 1),
		stop:                 make(chan struct{}),
		tasks:                make(chan taskFn, 5),
	}
}

func NewWithMetrics(dm *datamodel.DataModel, m *metrics.Metrics) *Simulator {
	s := New(dm)
	s.metrics = m
	return s
}

// Start starts the simulator and initiates an inform session.
func (s *Simulator) Start(ctx context.Context) error {
	if Config.ConnReqEnableHTTP {
		srv, err := newHTTPServer(s.handleConnectionRequest)
		if err != nil {
			return fmt.Errorf("start connection request server: %w", err)
		}
		s.httpServer = srv
		log.Info().Str("server_url", s.httpServer.url()).Msg("Started HTTP connection request server")
	}
	if Config.ConnReqEnableUDP {
		port := int(Config.Port)
		if s.httpServer.listenPort() != 0 {
			port = s.httpServer.listenPort()
		}
		if port != 0 {
			us, err := newUDPServer(ctx, port, s.handleConnectionRequest)
			if err != nil {
				return fmt.Errorf("start connection request server: %w", err)
			}
			s.udpServer = us
			log.Info().Str("server_url", s.udpServer.url()).Msg("Started UDP connection request server")
		} else {
			log.Warn().Msg("Can't start UDP connection request server on undefined port")
		}
	}

	s.startedAt = time.Now()
	s.dm.SetConnectionRequestURL(s.httpServer.url())
	s.dm.SetUDPConnectionRequestAddress(s.udpServer.url())
	s.SetPeriodicInformInterval(Config.InformInterval)
	go s.periodicInform(ctx)

	if !s.dm.IsBootstrapped() {
		s.pendingEvents <- rpc.EventBootstrap
	} else {
		s.pendingEvents <- rpc.EventBoot
	}

	return nil
}

// Stop stops the simulator.
func (s *Simulator) Stop(ctx context.Context) error {
	close(s.stop)
	if err := s.dm.SaveState(Config.StateFilePath); err != nil {
		return fmt.Errorf("save state: %w", err)
	}
	if err := s.httpServer.stop(ctx); err != nil {
		return fmt.Errorf("stop HTTP connection request server: %w", err)
	}
	if err := s.udpServer.stop(ctx); err != nil {
		return fmt.Errorf("stop HTTP connection request server: %w", err)
	}
	return nil
}

func (s *Simulator) SetPeriodicInformInterval(dur time.Duration) {
	if dur > 0 {
		s.dm.SetPeriodicInformInterval(int64(dur.Seconds()))
	}
}

func (s *Simulator) handleConnectionRequest(_ context.Context, params crParams) error {
	if s.dm.DownUntil().After(time.Now()) {
		return errServiceUnavailable
	}
	if params.un != s.dm.ConnectionRequestUsername().Value {
		return errForbidden
	}
	sig := sign(params.ts+params.id+params.un+params.cn, s.dm.ConnectionRequestPassword().Value)
	if sig != params.sig {
		return errForbidden
	}

	select {
	case s.pendingEvents <- rpc.EventConnectionRequest:
	default:
	}
	return nil
}

// nolint:gocyclo
func (s *Simulator) handleEnvelope(env *rpc.EnvelopeDecoder) *rpc.EnvelopeEncoder {
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
	case env.Body.TransferCompleteResponse != nil:
		return nil
	default:
		log.Warn().Msg("Unknown method")
		return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
	}
}

func (s *Simulator) handleGetQueuedTransfers(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "GetQueuedTransfers").Msg("Received message")
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}

func (s *Simulator) handleGetAllQueuedTransfers(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "GetAllQueuedTransfers").Msg("Received message")
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}

func (s *Simulator) handleScheduleInform(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "ScheduleInform").Msg("Received message")
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}

func (s *Simulator) handleSetVouchers(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "SetVouchers").Msg("Received message")
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}

func (s *Simulator) handleGetOptions(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "GetOptions").Msg("Received message")
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}

func (s *Simulator) handleFault(envID string, r *rpc.FaultPayload) *rpc.EnvelopeEncoder {
	log.Error().
		Str("env_id", envID).
		Str("code", r.Detail.Fault.FaultCode.String()).
		Str("string", r.Detail.Fault.FaultString).
		Msg("ACS fault")
	return nil
}

func (s *Simulator) pretendOfflineFor(dur time.Duration) {
	downUntil := time.Now().Add(dur)
	s.dm.SetDownUntil(downUntil)
	s.startedAt = downUntil
	time.Sleep(dur)
}

func (s *Simulator) stopped() bool {
	select {
	case <-s.stop:
		return true
	default:
		return false
	}
}

func (s *Simulator) newEnvelope() *rpc.EnvelopeEncoder {
	return rpc.NewEnvelope(s.nextEnvelopeID())
}

func (s *Simulator) nextEnvelopeID() string {
	id := atomic.AddUint64(&s.envelopeID, 1)
	return strconv.FormatUint(id, 10)
}

func sign(input, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(input))
	return hex.EncodeToString(h.Sum(nil))
}

func prettyXML(b []byte) string {
	return strings.TrimSpace(xmlfmt.FormatXML(string(b), "", "    "))
}
