package simulator

import (
	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleGetParameterAttributes(envID string, r *rpc.GetParameterAttributesRequest) *rpc.EnvelopeEncoder {
	resp := rpc.NewEnvelope(envID)
	names := r.ParameterNames.Names
	attrs := []rpc.ParameterAttributeStruct{}
	for _, path := range names {
		batch, ok := s.dm.GetAll(path)
		if !ok {
			return resp.WithFault(rpc.FaultInvalidParameterName)
		}

		for _, p := range batch {
			attrs = append(attrs, rpc.ParameterAttributeStruct{
				Name:         p.Path,
				Notification: p.Notification,
				AccessList: rpc.AccessListEncoder{
					ArrayType: rpc.ArrayType(rpc.XSD(rpc.TypeString), len(p.ACL)),
					Values:    p.ACL,
				},
			})
		}
	}

	resp.Body.GetParameterAttributesResponse = &rpc.GetParameterAttributesResponseEncoder{
		ParameterList: rpc.ParameterAttributeStructEncoder{
			ArrayType:           rpc.ArrayType("cwmp:ParameterAttributeStruct", len(attrs)),
			ParameterAttributes: attrs,
		},
	}
	return resp
}
