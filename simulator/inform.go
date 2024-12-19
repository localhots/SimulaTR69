package simulator

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

type (
	sessionHandler func(ctx context.Context, client *http.Client)
	taskFn         func() taskFn
)

func (s *Simulator) periodicInform(ctx context.Context) {
	for !s.stopped() {
		if !s.dm.PeriodicInformEnabled() {
			log.Info().Msg("Periodic inform disabled")
		}

		delay := time.Until(s.nextInformTime())
		log.Info().
			Str("delay", delay.Truncate(time.Millisecond).String()).
			Msg("Scheduling next Inform request")

		select {
		case <-time.After(delay):
			s.dm.AddEvent(rpc.EventPeriodic)
			s.startSession(ctx, s.informHandler)
		case evt := <-s.pendingEvents:
			s.dm.AddEvent(evt)
			s.startSession(ctx, s.informHandler)
		case <-s.informScheduleUpdate:
		case <-s.stop:
			return
		}

		// Run all avialable tasks after session is finished
		log.Debug().Msg("Start processing tasks")
		s.processTasks()
		log.Debug().Msg("Finished processing tasks")
	}
}

func (s *Simulator) nextInformTime() time.Time {
	return calcInformTime(
		s.dm.PeriodicInformTime(),
		s.startedAt,
		time.Now(),
		s.dm.PeriodicInformEnabled(),
		s.dm.PeriodicInformInterval(),
	)
}

func (s *Simulator) resetInformTimer() {
	s.informScheduleUpdate <- struct{}{}
}

// startSession initiates a new session with the ACS.
func (s *Simulator) startSession(ctx context.Context, handler sessionHandler) {
	if s.stopped() {
		return
	}

	// Allow only one session at a time
	if ok := s.sessionMux.TryLock(); !ok {
		log.Warn().Msg("Session in progress, dropping request")
		return
	}
	defer s.sessionMux.Unlock()

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

	log.Info().Str("acs_url", Config.ACSURL).Msg("Connecting to ACS")
	client, closeFn, err := newClient(u.Hostname(), tcpPort(u))
	s.metrics.ConnectionLatency.Observe(float64(time.Since(connectionStartTime).Milliseconds()))
	if err != nil {
		log.Error().Err(err).Msg("Failed to connect to ACS")
		s.metrics.RequestFailures.Inc()
		s.dm.IncrRetryAttempts()
		return
	}
	defer func() { _ = closeFn() }()

	handler(ctx, &client)
}

// nolint:gocyclo
func (s *Simulator) informHandler(ctx context.Context, client *http.Client) {
	log.Info().Msg("Starting inform")
	informEnv := s.makeInformEnvelope()
	resp, err := s.request(ctx, client, informEnv)
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
pendingRequests:
	for {
		select {
		case envelopeBuilder := <-s.pendingRequests:
			env := s.newEnvelope()
			envelopeBuilder(env)

			acsResponseEnv, err := s.send(ctx, client, env)
			if err != nil {
				log.Error().Err(err).Msg("Failed to make request")
				s.metrics.RequestFailures.Inc()
				return
			}
			nextEnv = s.handleEnvelope(acsResponseEnv)
		default:
			break pendingRequests
		}
	}
	for {
		acsRequestEnv, err := s.send(ctx, client, nextEnv)
		if err != nil {
			log.Error().Err(err).Msg("Failed to make request")
			s.metrics.RequestFailures.Inc()
			return
		}
		if acsRequestEnv == nil {
			log.Info().Msg("Got empty response from ACS, inform finished")
			break
		}

		nextEnv = s.handleEnvelope(acsRequestEnv)
		if nextEnv == nil {
			break
		}
	}

	for _, evt := range informEnv.Body.Inform.Event.Events {
		if evt.EventCode == rpc.EventBootstrap {
			s.dm.SetBootstrapped(true)
			break
		}
	}
}

func (s *Simulator) send(ctx context.Context, client *http.Client, env *rpc.EnvelopeEncoder) (*rpc.EnvelopeDecoder, error) {
	log.Debug().Msg("Sending post-inform request")
	resp, err := s.request(ctx, client, env)
	if err != nil {
		return nil, fmt.Errorf("make request: %w", err)
	}
	if resp.Body == nil {
		// Got empty response from ACS, inform finished
		return nil, nil
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if err := resp.Body.Close(); err != nil {
		return nil, fmt.Errorf("close response buffer: %w", err)
	}
	if len(b) == 0 {
		// Got empty response from ACS, inform finished
		return nil, nil
	}

	// FIXME: make conditional call to prettyXML
	log.Trace().Msg("Response from ACS\n" + prettyXML(b))
	acsRequestEnv, err := rpc.Decode(b)
	if err != nil {
		return nil, fmt.Errorf("decode envelope: %w", err)
	}

	return acsRequestEnv, nil
}

func (s *Simulator) makeInformEnvelope() *rpc.EnvelopeEncoder {
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
func (s *Simulator) request(ctx context.Context, client *http.Client, env *rpc.EnvelopeEncoder) (*http.Response, error) {
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

func (s *Simulator) processTasks() {
	// Any tasks that are produced as a result of current batch will be executed
	// next time. This is done to allow tasks to schedule a session and a task
	// that needs to be run after that session completes.
	next := []taskFn{}
	defer func() {
		for _, t := range next {
			s.tasks <- t
		}
	}()

	// Process currently scheduled tasks.
	for {
		select {
		case task := <-s.tasks:
			if nt := task(); nt != nil {
				next = append(next, nt)
			}
		default:
			return
		}
	}
}

func (s *Simulator) debugEnvelope(env *rpc.EnvelopeEncoder) {
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
