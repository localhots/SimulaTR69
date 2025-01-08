package simulator

import (
	"github.com/localhots/SimulaTR69/rpc"
)

func (s *Simulator) handleUpload(envID string, _ *rpc.UploadRequest) *rpc.EnvelopeEncoder {
	return rpc.NewEnvelope(envID).WithFault(rpc.FaultMethodNotSupported)
}
