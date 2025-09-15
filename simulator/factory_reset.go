package simulator

import (
	"context"

	"github.com/localhots/blip/noctx/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleFactoryReset(envID string) *rpc.EnvelopeEncoder {
	resp := rpc.NewEnvelope(envID)
	resp.Body.FactoryResetResponse = &rpc.FactoryResetResponseEncoder{}

	s.tasks <- func() taskFn {
		s.logger.Debug(context.TODO(), "Simulating factory reset", log.F{"delay": Config.UpgradeDelay})
		s.pretendOfflineFor(Config.UpgradeDelay)

		s.dm.Reset()
		s.dm.SetConnectionRequestURL(s.httpServer.url())
		s.dm.SetUDPConnectionRequestAddress(s.udpServer.url())
		if Config.SerialNumber != "" {
			s.dm.SetSerialNumber(Config.SerialNumber)
		}

		s.logger.Debug(context.TODO(), "Starting up")
		s.pendingEvents <- rpc.EventBootstrap
		return nil
	}
	return resp
}
