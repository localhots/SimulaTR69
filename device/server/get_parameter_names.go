package server

import (
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) handleGetParameterNames(envID string, r *rpc.GetParameterNamesRequest) rpc.EnvelopeEncoder {
	log.Info().Str("method", "GetParameterNames").Msg("Received message")
	r.Debug()
	names := s.dm.ParameterNames(r.ParameterPath, r.NextLevel)
	params := make([]rpc.ParameterInfoStruct, 0, len(names))
	for _, p := range names {
		params = append(params, rpc.ParameterInfoStruct{
			Name:     p.Path,
			Writable: p.Writable,
		})
	}
	resp := rpc.NewEnvelope(envID)
	resp.Body.GetParameterNamesResponse = &rpc.GetParameterNamesResponseEncoder{
		ParameterList: rpc.ParameterInfoEncoder{
			ArrayType:  rpc.ArrayType("cwmp:ParameterInfoStruct", len(params)),
			Parameters: params,
		},
	}
	return resp
}
