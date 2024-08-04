package server

import (
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/server/rpc"
)

// AccessList values are intentionally not respected.
func (s *Server) handleSetParameterAttributes(envID string, r *rpc.SetParameterAttributesRequest) rpc.EnvelopeEncoder {
	log.Info().Str("method", "SetParameterAttributes").Msg("Received message")
	r.Debug()
	attrs := r.ParameterList.ParameterAttributes
	for _, attr := range attrs {
		s.dm.SetParameterAttribute(attr.Name,
			int(attr.Notification), attr.NotificationChange,
			attr.AccessList.Values, attr.AccessListChange,
		)
	}

	resp := rpc.NewEnvelope(envID)
	resp.Body.SetParameterAttributesResponse = &rpc.SetParameterAttributesResponseEncoder{}
	return resp
}
