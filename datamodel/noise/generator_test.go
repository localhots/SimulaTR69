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
	scale      = 1.0
)

func TestRandomWalkBounds(t *testing.T) {
	gen := RandomWalk(startValue, minValue, maxValue, step)
	for range 100 {
		value := gen()
		if value < minValue || value > maxValue {
			t.Errorf("Value out of bounds: got %v, want between %v and %v", value, minValue, maxValue)
		}
	}
}

func TestPiecewiseLinearBounds(t *testing.T) {
	gen := PiecewiseLinear(startValue, minValue, maxValue, step)
	for range 100 {
		value := gen()
		if value < minValue || value > maxValue {
			t.Errorf("Value out of bounds: got %v, want between %v and %v", value, minValue, maxValue)
		}
	}
}

func TestSineWithNoiseBounds(t *testing.T) {
	gen := SineWithNoise(offset, amplitude, frequency, phase, noiseScale)
	for range 100 {
		value := gen()
		// Since sine wave values range between -amplitude and +amplitude, we add noiseScale to the bounds
		if value < -amplitude-noiseScale || value > amplitude+noiseScale {
			t.Errorf("Value out of bounds: got %v, want between %v and %v", value, -amplitude-noiseScale, amplitude+noiseScale)
		}
	}
}

func TestPerlinNoiseBounds(t *testing.T) {
	gen := PerlinNoise(offset, alpha, beta, scale)
	for range 100 {
		value := gen()
		// Perlin noise values are typically between -1 and 1, scaled and offset
		if value < -scale+offset || value > scale+offset {
			t.Errorf("Value out of bounds: got %v, want between %v and %v", value, -scale+offset, scale+offset)
		}
	}
}

func TestTrendWithNoiseBounds(t *testing.T) {
	gen := TrendWithNoise(startValue, step, noiseScale)
	for range 100 {
		value := gen()
		// Since the trend can go indefinitely, we only check that the noise does not exceed the noiseScale
		if step >= 0 && value < startValue || step < 0 && value > startValue {
			t.Errorf("Value out of bounds: got %v, want at least %v", value, startValue)
		}
	}
}

func BenchmarkRandomWalk(b *testing.B) {
	gen := RandomWalk(startValue, minValue, maxValue, step)
	for b.Loop() {
		gen()
	}
}

func BenchmarkPiecewiseLinear(b *testing.B) {
	gen := PiecewiseLinear(startValue, minValue, maxValue, step)
	for b.Loop() {
		gen()
	}
}

func BenchmarkSineWithNoise(b *testing.B) {
	gen := SineWithNoise(offset, amplitude, frequency, phase, noiseScale)
	for b.Loop() {
		gen()
	}
}

func BenchmarkPerlinNoise(b *testing.B) {
	gen := PerlinNoise(offset, alpha, beta, scale)
	b.ResetTimer()
	for b.Loop() {
		gen()
	}
}

func BenchmarkTrendWithNoise(b *testing.B) {
	gen := TrendWithNoise(startValue, step, noiseScale)
	for b.Loop() {
		gen()
	}
}
