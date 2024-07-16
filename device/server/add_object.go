package server

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) handleAddObject(envID string, r *rpc.AddObjectRequest) rpc.EnvelopeEncoder {
	log.Info().Str("method", "AddObject").Msg("Received message")
	r.Debug()
	resp := rpc.NewEnvelope(envID)
	if !strings.HasSuffix(r.ObjectName, ".") {
		return resp.WithFault(rpc.FaultInvalidParameterName)
	}

	i := s.dm.AddObject(r.ObjectName)
	if i == nil {
		return rpc.NewEnvelope(envID).WithFault(rpc.FaultInvalidParameterName)
	}
	s.dm.SetParameterKey(r.ParameterKey)

	resp.Body.AddObjectResponse = &rpc.AddObjectResponseEncoder{
		InstanceNumber: uint(*i),
		Status:         0,
	}
	return resp
}
