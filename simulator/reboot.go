package simulator

import (
	"context"

	"github.com/localhots/blip/noctx/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleReboot(envID string, r *rpc.RebootRequest) *rpc.EnvelopeEncoder {
	s.logger.Info(context.TODO(), "Received message", log.F{"method": "Reboot"})
	resp := rpc.NewEnvelope(envID)
	resp.Body.RebootResponse = &rpc.RebootResponseEncoder{}
	s.dm.SetCommandKey(r.CommandKey)

	s.tasks <- func() taskFn {
		s.logger.Debug(context.TODO(), "Simulating reboot", log.F{"delay": Config.RebootDelay})
		s.pretendOfflineFor(Config.RebootDelay)
		s.logger.Debug(context.TODO(), "Starting up")
		s.pendingEvents <- rpc.EventBoot
		return nil
	}
	return resp
}
