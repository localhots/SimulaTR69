package noise

import (
	"testing"
)

const (
	startValue = 0.0
	minValue   = -1.0
	maxValue   = 1.0
	step       = 0.1
	offset     = 0.0
	amplitude  = 1.0
	frequency  = 0.1
	phase      = 0.0
	noiseScale = 0.1
	alpha      = 2.0
	beta       = 2.0
	seed       = int64(42)
	scale      = 1.0
)

func TestRandomWalkBounds(t *testing.T) {
	gen := RandomWalk(startValue, minValue, maxValue, step)
	for i := 0; i < 100; i++ {
		value := gen()
		if value < minValue || value > maxValue {
			t.Errorf("Value out of bounds: got %v, want between %v and %v", value, minValue, maxValue)
		}
	}
}

func TestPiecewiseLinearBounds(t *testing.T) {
	gen := PiecewiseLinear(startValue, minValue, maxValue, step)
	for i := 0; i < 100; i++ {
		value := gen()
		if value < minValue || value > maxValue {
			t.Errorf("Value out of bounds: got %v, want between %v and %v", value, minValue, maxValue)
		}
	}
}

func TestSineWithNoiseBounds(t *testing.T) {
	gen := SineWithNoise(offset, amplitude, frequency, phase, noiseScale)
	for i := 0; i < 100; i++ {
		value := gen()
		// Since sine wave values range between -amplitude and +amplitude, we add noiseScale to the bounds
		if value < -amplitude-noiseScale || value > amplitude+noiseScale {
			t.Errorf("Value out of bounds: got %v, want between %v and %v", value, -amplitude-noiseScale, amplitude+noiseScale)
		}
	}
}

func TestPerlinNoiseBounds(t *testing.T) {
	gen := PerlinNoise(offset, alpha, beta, seed, scale)
	for i := 0; i < 100; i++ {
		value := gen()
		// Perlin noise values are typically between -1 and 1, scaled and offset
		if value < -scale+offset || value > scale+offset {
			t.Errorf("Value out of bounds: got %v, want between %v and %v", value, -scale+offset, scale+offset)
		}
	}
}

func BenchmarkRandomWalk(b *testing.B) {
	gen := RandomWalk(startValue, minValue, maxValue, step)
	for i := 0; i < b.N; i++ {
		gen()
	}
}

func BenchmarkPiecewiseLinear(b *testing.B) {
	gen := PiecewiseLinear(startValue, minValue, maxValue, step)
	for i := 0; i < b.N; i++ {
		gen()
	}
}

func BenchmarkSineWithNoise(b *testing.B) {
	gen := SineWithNoise(offset, amplitude, frequency, phase, noiseScale)
	for i := 0; i < b.N; i++ {
		gen()
	}
}

func BenchmarkPerlinNoise(b *testing.B) {
	gen := PerlinNoise(offset, alpha, beta, seed, scale)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen()
	}
}
