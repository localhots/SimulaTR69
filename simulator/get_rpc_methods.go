package simulator

import (
	"context"

	"github.com/localhots/blip/noctx/log"

	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleGetRPCMethods(envID string) *rpc.EnvelopeEncoder {
	s.logger.Info(context.TODO(), "Received message", log.F{"method": "GetRPCMethods"})
	methods := rpc.SupportedMethods()
	for _, m := range methods {
		s.logger.Debug(context.TODO(), "GetRPCMethodsResponse", log.F{"method": m})
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
