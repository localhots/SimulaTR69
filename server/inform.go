package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/icholy/digest"
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) periodicInform(ctx context.Context) {
	for {
		it := s.dm.PeriodicInformTime()
		if delay := time.Until(it); delay > 0 {
			log.Info().Time("time", it).Msg("Inform delayed")
			time.Sleep(delay)
			s.Inform(ctx)
		}

		if s.dm.PeriodicInformEnabled() {
			ii := s.dm.PeriodicInformInterval()
			log.Info().Str("delay", ii.String()).Msg("Scheduling next Inform request")
			select {
			case <-time.After(ii):
				s.dm.AddEvent(rpc.EventPeriodic)
				s.Inform(ctx)
			case <-s.informScheduleUpdate:
			}
		} else {
			log.Info().Msg("Periodic inform disabled")
			<-s.informScheduleUpdate
		}
	}
}

func (s *Server) resetInformTimer() {
	s.informScheduleUpdate <- struct{}{}
}

// Inform initiates an inform message to the ACS.
// nolint:gocyclo
func (s *Server) Inform(ctx context.Context) {
	u, err := url.Parse(Config.ACSURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse ACS URL")
	}

	client, closeFn, err := newClient(u.Hostname(), tcpPort(u))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to ACS")
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
	params := []rpc.ParameterValueEncoder{
		s.dm.ConnectionRequestURL().Encode(),
	}
	for _, p := range s.dm.NotifyParams() {
		params = append(params, s.dm.GetValue(p).Encode())
	}
	s.dm.ClearNotifyParams()

	env := newEnvelope()
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
			ParameterValues: params,
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
		return nil, fmt.Errorf("execute request: %w", err)
	}
	s.cookies.SetCookies(req.URL, resp.Cookies())

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
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return http.Client{}, nil, fmt.Errorf("create a TCP connection to ACS: %w", err)
	}

	tr := &http.Transport{
		DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
			return conn, nil
		},
		Dial: func(_, _ string) (net.Conn, error) {
			return conn, nil
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !Config.ACSVerifyTLS,
		},
	}
	client := http.Client{Transport: tr}

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
