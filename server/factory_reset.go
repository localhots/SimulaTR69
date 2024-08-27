package server

import (
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/datamodel"
	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) handleFactoryReset(envID string) rpc.EnvelopeEncoder {
	log.Info().Str("method", "FactoryReset").Msg("Received message")
	resp := rpc.NewEnvelope(envID)

	dm, err := datamodel.Load(Config.DataModelPath, "")
	if err != nil {
		return resp.WithFaultMsg(rpc.FaultInternalError, err.Error())
	}
	dm.SetConnectionRequestURL(s.URL())
	if Config.SerialNumber != "" {
		dm.SetSerialNumber(Config.SerialNumber)
	}
	s.dm = dm

	s.pretendOfflineFor(Config.UpgradeDelay)

	resp.Body.FactoryResetResponse = &rpc.FactoryResetResponseEncoder{}
	return resp
}
