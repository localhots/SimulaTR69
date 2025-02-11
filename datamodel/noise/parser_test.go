package noise

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDef(t *testing.T) {
	tests := []struct {
		in  string
		exp *Func
		err string
	}{
		{
			in: "randomWalk(startValue=50, minValue=0, maxValue=100, step=5) as xsd:int",
			exp: &Func{
				Name: "randomWalk",
				Args: map[string]float64{"startValue": 50, "minValue": 0, "maxValue": 100, "step": 5},
				Type: "xsd:int",
			},
			err: "",
		},
		{
			in: "foo() as int",
			exp: &Func{
				Name: "foo",
				Args: map[string]float64{},
				Type: "int",
			},
			err: "",
		},
		{
			in: "foo(a=1, b=2) as int",
			exp: &Func{
				Name: "foo",
				Args: map[string]float64{"a": 1, "b": 2},
				Type: "int",
			},
			err: "",
		},
		{
			in: "foo(a=1,) as int",
			exp: &Func{
				Name: "foo",
				Args: map[string]float64{"a": 1},
				Type: "int",
			},
			err: "",
		},
		{
			in: "foo(a=-1) as int",
			exp: &Func{
				Name: "foo",
				Args: map[string]float64{"a": -1},
				Type: "int",
			},
			err: "",
		},
		{
			in:  "foo(a=bar) as int",
			exp: nil,
			err: `invalid generator definition`,
		},
		{
			in:  "invalid",
			exp: nil,
			err: "invalid generator definition",
		},
		{
			in:  "no_type()",
			exp: nil,
			err: "invalid generator definition",
		},
		{
			in:  "foo(a=1) as banana",
			exp: nil,
			err: `unsupported type: banana`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			out, err := ParseDef(tt.in)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.exp.Name, out.Name)
			assert.Equal(t, tt.exp.Type, out.Type)
			assert.Equal(t, tt.exp.Args, out.Args)
		})
	}
}

func TestGenerator(t *testing.T) {
	tests := []struct {
		name string
		fn   Func
		err  string
	}{
		{
			name: "valid randomWalk",
			fn: Func{
				Name: "randomWalk",
				Args: map[string]float64{
					"startValue": 50,
					"minValue":   0,
					"maxValue":   100,
					"step":       5,
				},
				Type: "xsd:int",
			},
			err: "",
		},
		{
			name: "missing function name",
			fn: Func{
				Name: "",
				Args: map[string]float64{
					"startValue": 50,
					"minValue":   0,
					"maxValue":   100,
					"step":       5,
				},
				Type: "xsd:int",
			},
			err: "function name is empty",
		},
		{
			name: "missing type",
			fn: Func{
				Name: "randomWalk",
				Args: map[string]float64{
					"startValue": 50,
					"minValue":   0,
					"maxValue":   100,
					"step":       5,
				},
				Type: "",
			},
			err: "value type is empty",
		},
		{
			name: "unknown function name",
			fn: Func{
				Name: "unknownFunc",
				Args: map[string]float64{
					"startValue": 50,
					"minValue":   0,
					"maxValue":   100,
					"step":       5,
				},
				Type: "xsd:int",
			},
			err: "unknown generator function",
		},
		{
			name: "invalid argument",
			fn: Func{
				Name: "randomWalk",
				Args: map[string]float64{
					"startValue": 50,
					"minValue":   0,
					"maxValue":   100,
					"invalid":    5,
				},
				Type: "xsd:int",
			},
			err: "missing required argument: step",
		},
		{
			name: "missing argument",
			fn: Func{
				Name: "randomWalk",
				Args: map[string]float64{
					"startValue": 50,
					"minValue":   0,
					"maxValue":   100,
				},
				Type: "xsd:int",
			},
			err: "invalid number of arguments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := tt.fn.Generator()
			if tt.err != "" {
				assert.ErrorContains(t, err, tt.err)
				assert.Nil(t, gen)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gen)
			}
		})
	}
}
