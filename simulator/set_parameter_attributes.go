package simulator

import (
	"github.com/localhots/SimulaTR69/rpc"
)

// AccessList values are intentionally not respected.
func (s *Simulator) handleSetParameterAttributes(envID string, r *rpc.SetParameterAttributesRequest) *rpc.EnvelopeEncoder {
	attrs := r.ParameterList.ParameterAttributes
	for _, attr := range attrs {
		s.dm.SetParameterAttribute(attr.Name,
			int(attr.Notification), attr.NotificationChange,
			attr.AccessList.Values, attr.AccessListChange,
		)
	}

	resp := rpc.NewEnvelope(envID)
	resp.Body.SetParameterAttributesResponse = &rpc.SetParameterAttributesResponseEncoder{}
	return resp
}
