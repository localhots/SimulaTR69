package simulator

import (
	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleGetRPCMethods(envID string) *rpc.EnvelopeEncoder {
	s.logger.Info().Str("method", "GetRPCMethods").Msg("Received message")
	methods := rpc.SupportedMethods()
	for _, m := range methods {
		s.logger.Debug().Str("method", m).Msg("GetRPCMethodsResponse")
	}
	resp := rpc.NewEnvelope(envID)
	resp.Body.GetRPCMethodsResponse = &rpc.GetRPCMethodsResponseEncoder{
		MethodList: rpc.MethodListEncoder{
			ArrayType: rpc.ArrayType("string", len(methods)),
			Methods:   methods,
		},
	}
	return resp
}
