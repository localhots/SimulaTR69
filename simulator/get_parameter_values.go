package simulator

import (
	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleGetParameterValues(envID string, r *rpc.GetParameterValuesRequest) *rpc.EnvelopeEncoder {
	resp := rpc.NewEnvelope(envID)
	names := r.ParameterNames.Names
	params := []rpc.ParameterValueEncoder{}
	for _, path := range names {
		batch, ok := s.dm.GetAll(path)
		if !ok {
			return resp.WithFault(rpc.FaultInvalidParameterName)
		}

		for _, p := range batch {
			if p.Object {
				continue
			}
			params = append(params, p.Encode())
		}
	}

	s.metrics.ParametersRead.Add(float64(len(params)))
	resp.Body.GetParameterValuesResponse = &rpc.GetParameterValuesResponseEncoder{
		ParameterList: rpc.ParameterListEncoder{
			ArrayType:       rpc.ArrayType("cwmp:ParameterValue", len(params)),
			ParameterValues: params,
		},
	}

	return resp
}
