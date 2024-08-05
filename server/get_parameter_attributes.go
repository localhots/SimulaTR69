package server

import (
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) handleGetParameterAttributes(envID string, r *rpc.GetParameterAttributesRequest) rpc.EnvelopeEncoder {
	log.Info().Str("method", "GetParameterAttributes").Msg("Received message")
	r.Debug()
	names := r.ParameterNames.Names
	attrs := []rpc.ParameterAttributeStruct{}
	for _, path := range names {
		batch := s.dm.GetAll(path)
		for _, p := range batch {
			attrs = append(attrs, rpc.ParameterAttributeStruct{
				Name:         p.Path,
				Notification: rpc.AttributeNotification(p.Notification),
				AccessList: rpc.AccessListEncoder{
					ArrayType: rpc.ArrayType("xsd:string", len(p.ACL)),
					Values:    p.ACL,
				},
			})
		}
	}

	resp := rpc.NewEnvelope(envID)
	resp.Body.GetParameterAttributesResponse = &rpc.GetParameterAttributesResponseEncoder{
		ParameterList: rpc.ParameterAttributeStructEncoder{
			ArrayType:           rpc.ArrayType("cwmp:ParameterAttributeStruct", len(attrs)),
			ParameterAttributes: attrs,
		},
	}
	return resp
}
