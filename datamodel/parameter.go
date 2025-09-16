package datamodel

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/localhots/blip/noctx/log"

	"github.com/localhots/SimulaTR69/datamodel/noise"
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

	genfn *noise.Func
	gen   noise.Generator
}

// NormalizeParameters will normalize all datamodel parameters.
func NormalizeParameters(params map[string]Parameter) {
	for path, param := range params {
		param.Normalize()
		params[path] = param
	}
}

// Name returns parameter name.
func (p *Parameter) Name() string {
	tokens := strings.Split(p.Path, ".")
	return tokens[len(tokens)-1]
}

// GetValue returns a parameter value. If the parameter has a generator function
// it will be used to produce a value, otherwise the value from the parameter
// will be returned.
func (p *Parameter) GetValue() string {
	if p.gen != nil {
		switch rpc.NoXSD(p.Type) {
		case rpc.TypeInt, rpc.TypeLong:
			return strconv.FormatInt(int64(p.gen()), 10)
		case rpc.TypeUnsignedInt, rpc.TypeUnsignedLong:
			return strconv.FormatUint(uint64(p.gen()), 10)
		case rpc.TypeFloat, rpc.TypeDouble:
			return strconv.FormatFloat(p.gen(), 'f', -1, 64)
		case rpc.TypeBoolean:
			return strconv.FormatBool(int(p.gen()) == 1)
		default:
			return ""
		}
	}
	return p.Value
}

// Encode converts a parameter into RPC ParameterValue structure.
func (p *Parameter) Encode() rpc.ParameterValueEncoder {
	return rpc.ParameterValueEncoder{
		Name: p.Path,
		Value: rpc.ValueEncoder{
			Type:  p.Type,
			Value: p.GetValue(),
		},
	}
}

// Normalize attempts to normalize parameter type and value in order to make it
// fully compliant with SOAP data types spec.
func (p *Parameter) Normalize() {
	if p.Object {
		p.Type = rpc.TypeObject
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
		log.Warn("Failed to parse parameter type", log.Cause(err), log.F{
			"parameter": p.Path,
			"type":      p.Type,
			"fallback":  rpc.XSD(rpc.TypeString),
		})
		p.Type = rpc.XSD(rpc.TypeString)
	}
	p.Type = td.String()
	p.Value = normalizeValue(td, p.Path, p.Value)
}

func (p *Parameter) initGenerator() error {
	if p.Type != rpc.TypeGenerator {
		return nil
	}

	var err error
	p.genfn, err = noise.ParseDef(p.Value)
	if err != nil {
		return fmt.Errorf("parse generator definition: %w", err)
	}
	p.gen, err = p.genfn.Generator()
	if err != nil {
		return fmt.Errorf("create generator: %w", err)
	}
	p.Type = p.genfn.Type

	return nil
}

// TODO: implement value ranges.
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
	switch strings.ToLower(val) {
	case "", "no", "off", "disabled":
		return "false"
	case "yes", "on", "enabled":
		return "true"
	}

	b, err := strconv.ParseBool(val)
	if err != nil {
		log.Warn("Invalid boolean value", log.F{
			"parameter": name,
			"value":     val,
			"fallback":  fallback,
		})
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
		log.Warn("Invalid integer value", log.F{
			"parameter": name,
			"value":     val,
			"fallback":  fallback,
		})
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
		log.Warn("Invalid unsigned integer value", log.F{
			"parameter": name,
			"value":     val,
			"fallback":  fallback,
		})
		return fallback
	}
	return val
}
