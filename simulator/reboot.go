package simulator

import (
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleReboot(envID string, r *rpc.RebootRequest) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "Reboot").Msg("Received message")
	resp := rpc.NewEnvelope(envID)
	resp.Body.RebootResponse = &rpc.RebootResponseEncoder{}
	s.dm.SetCommandKey(r.CommandKey)

	go func() {
		s.dm.AddEvent(rpc.EventBoot)
		s.pretendOfflineFor(Config.RebootDelay)
	}()
	return resp
}
