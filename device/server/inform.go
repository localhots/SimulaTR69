package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/icholy/digest"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) periodicInform() {
	for {
		if s.dm.PeriodicInformEnabled() {
			it := s.dm.PeriodicInformTime()
			if it.After(time.Now()) {
				log.Info().Time("time", it).Msg("Inform delayed")
				time.Sleep(time.Until(it))
			}

			ii := s.dm.PeriodicInformInterval()
			log.Info().Str("delay", ii.String()).Msg("Scheduling next Inform request")
			select {
			case <-time.After(ii):
				s.dm.AddEvent(rpc.EventPeriodic)
				s.Inform()
			case <-s.resetIformTimer:
			}
		} else {
			log.Info().Msg("Periodic inform disabled")
			<-s.resetIformTimer
		}
	}
}

func (s *Server) ResetInformTimer() {
	s.resetIformTimer <- struct{}{}
}

func (s *Server) Inform() {
	ctx := context.Background()
	u, err := url.Parse(Config.ACSURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse ACS URL")
	}
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", u.Hostname(), u.Port()))
	if err != nil {
		log.Error().Err(err).Msg("Failed to create a TCP connection to ACS")
		return
	}
	defer conn.Close()

	tr := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return conn, nil
		},
		Dial: func(network, addr string) (net.Conn, error) {
			return conn, nil
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

	informEnv := s.makeInformEnvelope()
	resp, err := s.request(ctx, &client, &informEnv)
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
	if tcr := s.dm.TryGetTransferComplete(); tcr != nil {
		tcrEnv := newEnvelope()
		tcrEnv.Body.TransferCompleteRequest = tcr
		nextEnv = &tcrEnv
		// TODO: Need to completely rewrite this.
	}
	for {
		log.Debug().Msg("Sending post-inform request")
		resp, err := s.request(ctx, &client, nextEnv)
		if err != nil {
			log.Error().Err(err).Msg("Failed to make request")
			return
		}
		if resp.Body == nil {
			log.Debug().Msg("Got empty response from ACS, inform finished")
			break
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read request")
			return
		}
		if len(b) == 0 {
			log.Debug().Msg("Got empty response from ACS, inform finished")
			break
		} else {
			log.Trace().Msg("Response from ACS\n" + prettyXML(b))
		}

		acsRequestEnv, err := rpc.Decode(b)
		if err != nil {
			log.Error().Err(err).Str("body", string(b)).Msg("Failed to decode envelope")
			return
		}

		acsResponseEnv := s.handleEnvelope(acsRequestEnv)
		nextEnv = &acsResponseEnv
	}

	events := informEnv.Body.Inform.Event.Events
	if len(events) == 1 && events[0].EventCode == rpc.EventBootstrap {
		s.dm.Bootstrapped = true
	}
}

func (s *Server) makeInformEnvelope() rpc.EnvelopeEncoder {
	deviceID := s.dm.DeviceID()
	events := []rpc.EventStruct{}
	for _, evt := range s.dm.PendingEvents() {
		events = append(events, rpc.EventStruct{
			EventCode:  evt,
			CommandKey: s.dm.GetCommandKey(),
		})
	}
	params := []rpc.ParameterValueEncoder{
		s.dm.ConnectionRequestURL().Encode(),
	}

	env := newEnvelope()
	env.Body.Inform = &rpc.InformRequestEncoder{
		DeviceId: rpc.DeviceID{
			Manufacturer: deviceID.Manufacturer,
			OUI:          deviceID.OUI,
			ProductClass: deviceID.ProductClass,
			SerialNumber: deviceID.SerialNumber,
		},
		Event: rpc.EventEncoder{
			ArrayType: rpc.ArrayType("cwmp:EventStruct", 1),
			Events:    events,
		},
		MaxEnvelopes: rpc.MaxEnvelopes,
		CurrentTime:  time.Now().Format(time.RFC3339),
		RetryCount:   int(s.dm.RetryAttempts),
		ParameterList: rpc.ParameterListEncoder{
			ArrayType:       rpc.ArrayType("cwmp:ParameterValueStruct", len(params)),
			ParameterValues: params,
		},
	}
	return env
}

func (s *Server) respond(w http.ResponseWriter, env rpc.EnvelopeEncoder) {
	b, err := env.EncodePretty()
	if err != nil {
		log.Error().Err(err).Msg("Failed to encode envelope")
	}
	_, err = w.Write(b)
	if err != nil {
		log.Error().Err(err).Msg("Failed to write response")
	}
}

// Returns false only if request to ACS was attempted and failed.
func (s *Server) request(ctx context.Context, client *http.Client, env *rpc.EnvelopeEncoder) (*http.Response, error) {
	var buf io.Reader
	if env != nil {
		b, err := env.EncodePretty()
		if err != nil {
			return nil, fmt.Errorf("encode envelope: %w", err)
		}
		buf = bytes.NewBuffer(b)
		logger := log.Info().Str("method", env.Method())
		if env.Body.Inform != nil {
			logger.Strs("events", s.dm.PendingEvents())
		}
		logger.Msg("Sending envelope")

		gpn := env.Body.GetParameterNamesResponse
		gpv := env.Body.GetParameterValuesResponse
		if gpn != nil && len(gpn.ParameterList.Parameters) > 100 {
			log.Debug().Msg("Sending all parameter names")
		} else if gpv != nil && len(gpv.ParameterList.ParameterValues) > 100 {
			log.Debug().Msg("Sending all parameter values")
		} else {
			if log.Logger.GetLevel() == zerolog.TraceLevel {
				log.Trace().Msg("Request to ACS\n" + strings.TrimSpace(string(b)))
			} else {
				log.Debug().Msg("Request to ACS")
			}
		}
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
