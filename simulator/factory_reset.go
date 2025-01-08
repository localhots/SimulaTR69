package simulator

import (
	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleFactoryReset(envID string) *rpc.EnvelopeEncoder {
	resp := rpc.NewEnvelope(envID)
	resp.Body.FactoryResetResponse = &rpc.FactoryResetResponseEncoder{}

	s.tasks <- func() taskFn {
		s.logger.Debug().Dur("delay", Config.UpgradeDelay).Msg("Simulating factory reset")
		s.pretendOfflineFor(Config.UpgradeDelay)

		s.dm.Reset()
		s.dm.SetConnectionRequestURL(s.server.url())
		if Config.SerialNumber != "" {
			s.dm.SetSerialNumber(Config.SerialNumber)
		}

		s.logger.Debug().Msg("Starting up")
		s.pendingEvents <- rpc.EventBootstrap
		return nil
	}
	return resp
}
