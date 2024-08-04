package server

import (
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/server/rpc"
)

func (s *Server) handleFactoryReset(envID string) rpc.EnvelopeEncoder {
	log.Info().Str("method", "FactoryReset").Msg("Received message")
	resp := rpc.NewEnvelope(envID)
	resp.Body.FactoryResetResponse = &rpc.FactoryResetResponseEncoder{}
	return resp
}
