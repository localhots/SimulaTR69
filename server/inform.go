package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/icholy/digest"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) periodicInform(ctx context.Context) {
	s.inform(ctx)
	for !s.stopped() {
		if !s.dm.PeriodicInformEnabled() {
			log.Info().Msg("Periodic inform disabled")
		}

		delay := time.Until(s.nextInformTime())
		log.Info().Str("delay", delay.String()).Msg("Scheduling next Inform request")

		select {
		case <-time.After(delay):
			s.dm.AddEvent(rpc.EventPeriodic)
			s.inform(ctx)
		case <-s.informScheduleUpdate:
		case <-s.stop:
			return
		}
	}
}

func (s *Server) nextInformTime() time.Time {
	return calcInformTime(
		s.dm.PeriodicInformTime(),
		s.startedAt,
		time.Now(),
		s.dm.PeriodicInformEnabled(),
		s.dm.PeriodicInformInterval(),
	)
}

func (s *Server) resetInformTimer() {
	s.informScheduleUpdate <- struct{}{}
}

// inform initiates an inform message to the ACS.
// nolint:gocyclo
func (s *Server) inform(ctx context.Context) {
	if s.stopped() {
		return
	}

	// Allow only one session at a time
	if ok := s.informMux.TryLock(); !ok {
		log.Warn().Msg("Inform in progress, dropping request")
		return
	}
	defer s.informMux.Unlock()

	u, err := url.Parse(Config.ACSURL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to parse ACS URL")
		return
	}

	connectionStartTime := time.Now()
	defer func() {
		s.metrics.ConcurrentInforms.Dec()
		s.metrics.InformDuration.Observe(float64(time.Since(connectionStartTime).Milliseconds()))
	}()
	s.metrics.ConcurrentInforms.Inc()
	s.metrics.ConnectionLatency.Observe(float64(time.Since(connectionStartTime).Milliseconds()))

	log.Info().Str("acs_url", Config.ACSURL).Msg("Connecting to ACS")
	client, closeFn, err := newClient(u.Hostname(), tcpPort(u))
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to ACS")
		s.metrics.RequestFailures.Inc()
		s.dm.IncrRetryAttempts()
		return
	}
	defer func() { _ = closeFn() }()

	informEnv := s.makeInformEnvelope()
	resp, err := s.request(ctx, &client, informEnv)
	if err != nil {
		log.Error().Err(err).Msg("Failed to make request")
		s.dm.IncrRetryAttempts()
		return
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("Failed to read response")
		s.metrics.RequestFailures.Inc()
		s.dm.IncrRetryAttempts()
		return
	}
	log.Trace().Msg("Response from ACS\n" + prettyXML(b))
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Error().Int("status", resp.StatusCode).Msg("Unexpected response status")
		s.dm.IncrRetryAttempts()
		return
	}

	s.dm.ResetRetryAttempts()
	s.dm.ClearEvents()
	var nextEnv *rpc.EnvelopeEncoder
	for {
		log.Debug().Msg("Sending post-inform request")
		resp, err := s.request(ctx, &client, nextEnv)
		if err != nil {
			log.Error().Err(err).Msg("Failed to make request")
			return
		}
		if resp.Body == nil {
			log.Info().Msg("Got empty response from ACS, inform finished")
			break
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read request")
			s.metrics.RequestFailures.Inc()
			return
		}
		_ = resp.Body.Close()
		if len(b) == 0 {
			log.Info().Msg("Got empty response from ACS, inform finished")
			break
		}
		log.Trace().Msg("Response from ACS\n" + prettyXML(b))

		acsRequestEnv, err := rpc.Decode(b)
		if err != nil {
			log.Error().Err(err).Str("body", string(b)).Msg("Failed to decode envelope")
			return
		}

		nextEnv = s.handleEnvelope(acsRequestEnv)
		if nextEnv == nil {
			return
		}
	}

	events := informEnv.Body.Inform.Event.Events
	if len(events) == 1 && events[0].EventCode == rpc.EventBootstrap {
		s.dm.SetBootstrapped(true)
	}
}

func (s *Server) makeInformEnvelope() *rpc.EnvelopeEncoder {
	s.dm.SetUptime(time.Since(s.startedAt))
	deviceID := s.dm.DeviceID()
	events := []rpc.EventStruct{}
	for _, evt := range s.dm.PendingEvents() {
		events = append(events, rpc.EventStruct{
			EventCode:  evt,
			CommandKey: s.dm.CommandKey(),
		})
	}
	params, _ := s.dm.GetValues(s.dm.NotifyParams()...)
	encParams := make([]rpc.ParameterValueEncoder, 0, len(params))
	for _, p := range params {
		encParams = append(encParams, p.Encode())
	}

	env := s.newEnvelope()
	env.Body.Inform = &rpc.InformRequestEncoder{
		DeviceId: rpc.DeviceID{
			Manufacturer: deviceID.Manufacturer,
			OUI:          deviceID.OUI,
			ProductClass: deviceID.ProductClass,
			SerialNumber: deviceID.SerialNumber,
		},
		Event: rpc.EventEncoder{
			ArrayType: rpc.ArrayType("cwmp:EventStruct", len(events)),
			Events:    events,
		},
		MaxEnvelopes: rpc.MaxEnvelopes,
		CurrentTime:  time.Now().Format(time.RFC3339),
		RetryCount:   int(s.dm.RetryAttempts()),
		ParameterList: rpc.ParameterListEncoder{
			ArrayType:       rpc.ArrayType("cwmp:ParameterValueStruct", len(params)),
			ParameterValues: encParams,
		},
	}
	return env
}

// Returns false only if request to ACS was attempted and failed.
func (s *Server) request(ctx context.Context, client *http.Client, env *rpc.EnvelopeEncoder) (*http.Response, error) {
	var buf io.Reader
	if env != nil {
		s.debugEnvelope(env)
		b, err := env.EncodePretty()
		if err != nil {
			return nil, fmt.Errorf("encode envelope: %w", err)
		}
		log.Trace().Msg("Request from ACS\n" + prettyXML(b))
		buf = bytes.NewBuffer(b)
	} else {
		log.Info().Msg("Sending empty POST request")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, Config.ACSURL, buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "text/xml; encoding=utf-8")
	for _, c := range s.cookies.Cookies(req.URL) {
		req.AddCookie(c)
	}

	resp, err := client.Do(req)
	if err != nil {
		s.metrics.RequestFailures.Inc()
		return nil, fmt.Errorf("execute request: %w", err)
	}
	s.cookies.SetCookies(req.URL, resp.Cookies())
	s.metrics.ResponseStatus.With(prometheus.Labels{
		"status": strconv.Itoa(resp.StatusCode),
	}).Inc()

	return resp, nil
}

func (s *Server) debugEnvelope(env *rpc.EnvelopeEncoder) {
	logger := log.Info().Str("method", env.Method())
	if env.Body.Inform != nil {
		logger.Strs("events", s.dm.PendingEvents())
	}
	if env.Body.Fault != nil {
		f := env.Body.Fault.Detail.Fault
		logger.Str("code", f.FaultCode.String())
		logger.Str("error", f.FaultString)
	}
	logger.Msg("Sending envelope")

	gpn := env.Body.GetParameterNamesResponse
	gpv := env.Body.GetParameterValuesResponse
	switch {
	case gpn != nil && len(gpn.ParameterList.Parameters) > 100:
		log.Debug().Msg("Sending all parameter names")
	case gpv != nil && len(gpv.ParameterList.ParameterValues) > 100:
		log.Debug().Msg("Sending all parameter values")
	default:
		log.Debug().Msg("Request to ACS")
	}
}

func newClient(host, port string) (http.Client, func() error, error) {
	dialer := net.Dialer{
		Timeout: Config.ConnectionTimeout,
	}
	conn, err := dialer.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return http.Client{}, nil, fmt.Errorf("create a TCP connection to ACS: %w", err)
	}

	tr := &http.Transport{
		Dial: func(_, _ string) (net.Conn, error) {
			return conn, nil
		},
		TLSClientConfig: &tls.Config{
			// nolint:gosec
			InsecureSkipVerify: !Config.ACSVerifyTLS,
		},
	}
	client := http.Client{
		Transport: tr,
		Timeout:   Config.RequestTimeout,
	}
	if Config.ACSAuth == AuthDigest {
		client.Transport = &digest.Transport{
			Transport: tr,
			Username:  Config.ACSUsername,
			Password:  Config.ACSPassword,
		}
	}

	return client, conn.Close, nil
}

func tcpPort(u *url.URL) string {
	if u.Port() != "" {
		return u.Port()
	}
	if u.Scheme == "https" {
		return "443"
	}
	return "80"
}

// calcInformTime calculates the time of the next inform based on all relevant
// parameters. It is meant to be wrapped by Server.nextInformTime and is written
// in such a way that it has no side effects and can be easily tested with unit
// tests.
func calcInformTime(
	periodicInformTime time.Time,
	startedAt time.Time,
	now time.Time,
	periodicInformEnabled bool,
	periodicInformInterval time.Duration,
) time.Time {
	if periodicInformTime.IsZero() {
		periodicInformTime = startedAt
	}
	if periodicInformTime.After(now) {
		return periodicInformTime
	}

	if !periodicInformEnabled {
		// At this point simulator should never inform
		// Adding an arbitrarily large time offset to current time
		return now.Add(365 * 24 * time.Hour)
	}

	intervalsElapsed := math.Ceil(now.Sub(periodicInformTime).Seconds() / periodicInformInterval.Seconds())
	return periodicInformTime.Add(time.Duration(intervalsElapsed) * periodicInformInterval)
}
