package simulator

import (
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) handleFactoryReset(envID string) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "FactoryReset").Msg("Received message")
	resp := rpc.NewEnvelope(envID)

	// TODO: Make it so factory reset action is executed at the end of Inform
	s.dm.Reset()
	s.dm.SetConnectionRequestURL(s.URL())
	if Config.SerialNumber != "" {
		s.dm.SetSerialNumber(Config.SerialNumber)
	}

	s.pretendOfflineFor(Config.UpgradeDelay)

	resp.Body.FactoryResetResponse = &rpc.FactoryResetResponseEncoder{}
	return resp
}
