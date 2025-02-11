package noise

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"

	"github.com/localhots/SimulaTR69/rpc"
)

// Func represents a noise generator function with its name, arguments, and type.
type Func struct {
	Name string
	Args map[string]float64
	Type string
}

// Generator is a function type that generates a float64 value.
type Generator func() float64

const (
	randomWalk      = "randomWalk"
	piecewiseLinear = "piecewiseLinear"
	sineWithNoise   = "sineWithNoise"
	perlinNoise     = "perlinNoise"
	trendWithNoise  = "trendWithNoise"
)

var (
	genReg  = regexp.MustCompile(`(?P<func_name>\w+)\((?P<args>(?:\w+=-?[0-9\.]+,?\s*)*)\)\s+as\s+(?P<type_name>[\w:]+)`)
	argsReg = regexp.MustCompile(`(\w+)=(-?[0-9\.]+)`)
)

// ParseDef parses generator function definitions from a string format.
// For example, it can parse a definition like this:
//
//	randomWalk(startValue=50, minValue=0, maxValue=100, step=5) as xsd:int
//
// The function returns a Func struct containing the parsed function name,
// arguments, and type, or an error if the parsing fails. A further call to
// Generator() is needed to create a generator function.
func ParseDef(str string) (*Func, error) {
	matches := genReg.FindStringSubmatch(str)
	if matches == nil {
		return nil, errors.New("invalid generator definition")
	}

	argMatches := argsReg.FindAllStringSubmatch(str, -1)
	args := make(map[string]float64)
	for _, match := range argMatches {
		k, v := match[1], match[2]
		if _, ok := args[k]; ok {
			return nil, fmt.Errorf("duplicate argument: %s", k)
		}
		val, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("parse float (%v): %w", v, err)
		}
		args[k] = val
	}

	// Validate return type
	switch rpc.NoXSD(matches[3]) {
	case rpc.TypeInt, rpc.TypeLong:
	case rpc.TypeUnsignedInt, rpc.TypeUnsignedLong:
	case rpc.TypeFloat, rpc.TypeDouble:
	case rpc.TypeBoolean:
	default:
		return nil, fmt.Errorf("unsupported type: %s", matches[3])
	}

	return &Func{
		Name: matches[1],
		Args: args,
		Type: matches[3],
	}, nil
}

// FullName returns the full descriptive name of the noise generator function.
func (fn *Func) FullName() string {
	switch fn.Name {
	case randomWalk:
		return "Random Walk"
	case piecewiseLinear:
		return "Piecewise Linear"
	case sineWithNoise:
		return "Sine Wave with Noise"
	case perlinNoise:
		return "Perlin Noise"
	case trendWithNoise:
		return "Trend With Noise"
	default:
		return "Unknown"
	}
}

// Generator creates a generator function based on the parsed function
// definition.
func (fn *Func) Generator() (Generator, error) {
	if fn.Name == "" {
		return nil, errors.New("function name is empty")
	}
	if fn.Type == "" {
		return nil, errors.New("value type is empty")
	}

	return createGenerator(fn.Name, fn.Args)
}

func createGenerator(name string, args map[string]float64) (Generator, error) {
	switch name {
	case randomWalk:
		if err := requireArgs(args, "startValue", "minValue", "maxValue", "step"); err != nil {
			return nil, err
		}
		return RandomWalk(args["startValue"], args["minValue"], args["maxValue"], args["step"]), nil
	case piecewiseLinear:
		if err := requireArgs(args, "startValue", "minValue", "maxValue", "step"); err != nil {
			return nil, err
		}
		return PiecewiseLinear(args["startValue"], args["minValue"], args["maxValue"], args["step"]), nil
	case sineWithNoise:
		if err := requireArgs(args, "offset", "amplitude", "frequency", "phase", "noiseScale"); err != nil {
			return nil, err
		}
		return SineWithNoise(args["offset"], args["amplitude"], args["frequency"], args["phase"], args["noiseScale"]), nil
	case perlinNoise:
		if err := requireArgs(args, "offset", "alpha", "beta", "scale"); err != nil {
			return nil, err
		}
		return PerlinNoise(args["offset"], args["alpha"], args["beta"], args["scale"]), nil
	case trendWithNoise:
		if err := requireArgs(args, "startValue", "step", "noiseScale"); err != nil {
			return nil, err
		}
		return TrendWithNoise(args["startValue"], args["step"], args["noiseScale"]), nil
	default:
		return nil, errors.New("unknown generator function")
	}
}

func requireArgs(args map[string]float64, reqs ...string) error {
	if len(args) != len(reqs) {
		return errors.New("invalid number of arguments")
	}
	for _, req := range reqs {
		if _, ok := args[req]; !ok {
			return fmt.Errorf("missing required argument: %s", req)
		}
	}
	return nil
}
