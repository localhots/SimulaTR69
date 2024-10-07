package server

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) handleAddObject(envID string, r *rpc.AddObjectRequest) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "AddObject").Msg("Received message")
	r.Debug()
	resp := rpc.NewEnvelope(envID)
	if !strings.HasSuffix(r.ObjectName, ".") {
		return resp.WithFaultMsg(rpc.FaultInvalidParameterName, "object name must end with a dot")
	}

	i, err := s.dm.AddObject(r.ObjectName)
	if err != nil {
		return resp.WithFaultMsg(rpc.FaultInvalidParameterName, err.Error())
	}
	s.dm.SetParameterKey(r.ParameterKey)

	resp.Body.AddObjectResponse = &rpc.AddObjectResponseEncoder{
		InstanceNumber: i,
		Status:         0,
	}
	return resp
}
