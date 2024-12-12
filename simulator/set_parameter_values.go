package simulator

import (
	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/datamodel"
	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Server) handleSetParameterValues(envID string, r *rpc.SetParameterValuesRequest) *rpc.EnvelopeEncoder {
	log.Info().Str("method", "SetParameterValues").Msg("Received message")
	r.Debug()
	vals := r.ParameterList.ParameterValues
	params := make([]datamodel.Parameter, 0, len(vals))
	for _, v := range vals {
		params = append(params, datamodel.Parameter{
			Path:  v.Name,
			Type:  v.Value.Type,
			Value: v.Value.Value,
		})
	}

	var faults []rpc.SetParameterValuesFault
	for _, p := range params {
		if fc := s.dm.CanSetValue(p); fc != nil {
			faults = append(faults, rpc.SetParameterValuesFault{
				ParameterName: p.Path,
				FaultCode:     *fc,
				FaultString:   fc.String(),
			})
		}
	}
	if len(faults) > 0 {
		resp := rpc.NewEnvelope(envID).WithFault(rpc.FaultInvalidArguments)
		resp.Body.Fault.Detail.Fault.SetParameterValuesFault = faults
		return resp
	}

	var resetPeriodicInform bool
	for _, p := range params {
		if s.dm.IsPeriodicInformParameter(p.Path) {
			resetPeriodicInform = true
			break
		}
	}
	if resetPeriodicInform {
		defer s.resetInformTimer()
	}

	s.dm.SetValues(params)
	s.dm.SetParameterKey(r.ParameterKey)
	resp := rpc.NewEnvelope(envID)
	resp.Body.SetParameterValuesResponse = &rpc.SetParameterValuesResponseEncoder{
		Status: 0,
	}
	return resp
}
