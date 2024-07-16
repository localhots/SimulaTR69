package server

import (
	"time"

	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) handleReboot(envID string, r *rpc.RebootRequest) rpc.EnvelopeEncoder {
	log.Info().Str("method", "Reboot").Msg("Received message")
	resp := rpc.NewEnvelope(envID)
	resp.Body.RebootResponse = &rpc.RebootResponseEncoder{}
	s.dm.SetCommandKey(r.CommandKey)

	go func() {
		time.Sleep(5 * time.Second)
		s.dm.AddEvent(rpc.EventBoot)
		s.Inform()
	}()
	return resp
}
