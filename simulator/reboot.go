package simulator

import (
	"context"

	"github.com/localhots/blip/noctx/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleReboot(ctx context.Context, envID string, r *rpc.RebootRequest) *rpc.EnvelopeEncoder {
	s.logger.Info(ctx, "Received message", log.F{"method": "Reboot"})
	resp := rpc.NewEnvelope(envID)
	resp.Body.RebootResponse = &rpc.RebootResponseEncoder{}
	s.dm.SetCommandKey(r.CommandKey)

	s.tasks <- func() taskFn {
		s.logger.Debug(ctx, "Simulating reboot", log.F{"delay": Config.RebootDelay})
		s.pretendOfflineFor(Config.RebootDelay)
		s.logger.Debug(ctx, "Starting up")
		s.pendingEvents <- rpc.EventBoot
		return nil
	}
	return resp
}
