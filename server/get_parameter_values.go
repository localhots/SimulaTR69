package server

import (
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) handleGetParameterValues(envID string, r *rpc.GetParameterValuesRequest) rpc.EnvelopeEncoder {
	log.Info().Str("method", "GetParameterValues").Msg("Received message")
	r.Debug()
	names := r.ParameterNames.Names
	params := []rpc.ParameterValueEncoder{}
	for _, path := range names {
		batch := s.dm.Get(path)
		for _, p := range batch {
			params = append(params, rpc.ParameterValueEncoder{
				Name: p.Path,
				Value: rpc.ValueEncoder{
					Type:  p.Type,
					Value: p.Value,
				},
			})
		}
	}

	resp := rpc.NewEnvelope(envID)
	resp.Body.GetParameterValuesResponse = &rpc.GetParameterValuesResponseEncoder{
		ParameterList: rpc.ParameterListEncoder{
			ArrayType:       rpc.ArrayType("cwmp:ParameterValue", len(params)),
			ParameterValues: params,
		},
	}

	return resp
}
