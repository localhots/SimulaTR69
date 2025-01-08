package simulator

import (
	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleReboot(envID string, r *rpc.RebootRequest) *rpc.EnvelopeEncoder {
	s.logger.Info().Str("method", "Reboot").Msg("Received message")
	resp := rpc.NewEnvelope(envID)
	resp.Body.RebootResponse = &rpc.RebootResponseEncoder{}
	s.dm.SetCommandKey(r.CommandKey)

	s.tasks <- func() taskFn {
		s.logger.Debug().Dur("delay", Config.RebootDelay).Msg("Simulating reboot")
		s.pretendOfflineFor(Config.RebootDelay)
		s.logger.Debug().Msg("Starting up")
		s.pendingEvents <- rpc.EventBoot
		return nil
	}
	return resp
}
