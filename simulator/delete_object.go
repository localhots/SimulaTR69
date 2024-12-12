package simulator

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleDeleteObject(envID string, r *rpc.DeleteObjectRequest) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "DeleteObject").Msg("Received message")
	r.Debug()
	resp := rpc.NewEnvelope(envID)
	if !strings.HasSuffix(r.ObjectName, ".") {
		return resp.WithFault(rpc.FaultInvalidParameterName)
	}
	s.dm.DeleteObject(r.ObjectName)
	s.dm.SetParameterKey(r.ParameterKey)

	resp.Body.DeleteObjectResponse = &rpc.DeleteObjectResponseEncoder{
		Status: 0,
	}
	return resp
}
