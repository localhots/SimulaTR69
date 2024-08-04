package server

import (
	"github.com/localhots/SimulaTR69/rpc"
	"github.com/rs/zerolog/log"
)

func (s *Server) handleUpload(envID string, r *rpc.UploadRequest) rpc.EnvelopeEncoder {
	log.Info().Str("method", "Upload").Msg("Received message")
	r.Debug()
	// s.dm.SetCommandKey(r.CommandKey)
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}
