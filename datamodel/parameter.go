package datamodel

import (
	"github.com/localhots/SimulaTR69/rpc"
)

// Parameter describes a datamodel paremeter.
type Parameter struct {
	Path         string
	Object       bool
	Writable     bool
	Type         string
	Value        string
	Notification int
	ACL          []string
}

// Encode converts a parameter into RPC ParameterValue structure.
func (p Parameter) Encode() rpc.ParameterValueEncoder {
	return rpc.ParameterValueEncoder{
		Name: p.Path,
		Value: rpc.ValueEncoder{
			Type:  p.Type,
			Value: p.Value,
		},
	}
}
