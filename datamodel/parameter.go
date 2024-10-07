package datamodel

import (
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/localhots/SimulaTR69/rpc"
)

// Parameter describes a datamodel paremeter.
type Parameter struct {
	Path         string
	Object       bool
	Writable     bool
	Type         string
	Value        string
	Notification rpc.AttributeNotification
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

// Normalize attempts to normalize parameter type and value in order to make it
// fully compliant with SOAP data types spec.
func (p *Parameter) Normalize() {
	if p.Object {
		p.Type = rpc.TypeObject
		if !strings.HasSuffix(p.Path, ".") {
			p.Path += "."
		}
		return
	}
	if p.Type == "" {
		// Assume string if no type is specified. This should never happen but
		// practice shows that datamodel dumps can contain anything.
		p.Type = rpc.XSD(rpc.TypeString)
		return
	}

	td, err := parseTypeDef(p.Type)
	if err != nil {
		log.Warn().
			Err(err).
			Str("parameter", p.Path).
			Str("type", p.Type).
			Str("fallback", rpc.XSD(rpc.TypeString)).
			Msg("Failed to parse parameter type")
		p.Type = rpc.XSD(rpc.TypeString)
	}
	p.Type = td.String()
	p.Value = normalizeValue(td, p.Path, p.Value)
}

// TODO: implement value ranges
func normalizeValue(td *typeDef, name, val string) string {
	val = strings.TrimSpace(val)
	switch td.name {
	case rpc.TypeBoolean:
		return normalizeBool(name, val)
	case rpc.TypeInt, rpc.TypeLong:
		return normalizeInt(name, val)
	case rpc.TypeUnsignedInt, rpc.TypeUnsignedLong:
		return normalizeUint(name, val)
	default:
		return val
	}
}

func normalizeBool(name, val string) string {
	const fallback = "false"
	if val == "" {
		return fallback
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		log.Warn().
			Str("parameter", name).
			Str("value", val).
			Str("fallback", fallback).
			Msg("Invalid boolean value")
		return fallback
	}
	return strconv.FormatBool(b)
}

func normalizeInt(name, val string) string {
	const fallback = "0"
	if val == "" {
		return fallback
	}
	if _, err := strconv.ParseInt(val, 10, 64); err != nil {
		log.Warn().
			Str("parameter", name).
			Str("value", val).
			Str("fallback", fallback).
			Msg("Invalid integer value")
		return fallback
	}
	return val
}

func normalizeUint(name, val string) string {
	const fallback = "0"
	if val == "" {
		return fallback
	}
	if _, err := strconv.ParseUint(val, 10, 64); err != nil {
		log.Warn().
			Str("parameter", name).
			Str("value", val).
			Str("fallback", fallback).
			Msg("Invalid unsigned integer value")
		return fallback
	}
	return val
}
