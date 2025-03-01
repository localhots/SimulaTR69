// Package noise provides algorithms to generate sequences of values that
// simulate sensor readings with various patterns, including random walk,
// piecewise linear, sine wave with noise, and Perlin noise.
package noise

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/aquilax/go-perlin"
)

// RandomWalk algorithm generates a sequence of values where each value is
// derived from the previous one by adding a small random change. When
// generating fake values, this method simulates a sensor that produces
// readings which vary in a random but continuous manner.
func RandomWalk(startValue, minValue, maxValue, step float64) Generator {
	prevValue := startValue
	return func() float64 {
		// nolint:gosec
		// It's okay to use the default random number generator here.
		change := (rand.Float64()*2 - 1) * step
		newValue := clamp(prevValue+change, minValue, maxValue)
		prevValue = newValue
		return newValue
	}
}

// PiecewiseLinear algorithm generates a sequence of values where the trend
// changes direction at regular intervals. This method simulates a sensor that
// produces readings which follow a piecewise linear pattern, with occasional
// random fluctuations.
func PiecewiseLinear(startValue, minValue, maxValue, step float64) Generator {
	i := 0
	prevValue := startValue
	direction := 1.0
	return func() float64 {
		if i%20 == 0 {
			direction *= -1
		}
		// nolint:gosec
		// It's okay to use the default random number generator here.
		change := direction*step + (rand.Float64()*2-1)*(step/2)
		newValue := clamp(prevValue+change, minValue, maxValue)
		prevValue = newValue
		i++
		return newValue
	}
}

// SineWithNoise algorithm generates a sequence of values based on a sine wave
// with added random noise. This method simulates a sensor that produces
// readings which follow a sinusoidal pattern with some random fluctuations.
//
// Arguments:
//   - offset: The baseline value of the sine wave. Shifts the entire
//     wave up or down.
//   - amplitude: The peak value of the sine wave. Determines the height of the
//     wave.
//   - frequency: The number of cycles the sine wave completes in a unit
//     interval. Determines the wave's speed.
//   - phase: The initial angle of the sine wave at the start of the sequence.
//     Shifts the wave left or right.
//   - noiseScale: The amplitude of the random noise added to the sine wave.
//     Determines the intensity of the noise.
func SineWithNoise(offset, amplitude, frequency, phase, noiseScale float64) Generator {
	i := 0
	return func() float64 {
		// nolint:gosec
		// It's okay to use the default random number generator here.
		value := offset + amplitude*math.Sin(frequency*float64(i)+phase) + rand.Float64()*noiseScale
		i++
		return value
	}
}

// PerlinNoise algorithm generates a sequence of values based on Perlin noise.
// This method simulates a sensor that produces readings which follow a smooth,
// pseudo-random pattern.
//
// Arguments:
//   - alpha: Controls the smoothness of the Perlin noise. Higher values make
//     the noise smoother.
//   - beta: Controls the frequency of the Perlin noise. Higher values increase
//     the frequency.
//   - scale: A scaling factor to adjust the amplitude of the noise.
//   - offset: A constant value to be added to the generated noise values.
func PerlinNoise(offset, alpha, beta float64, scale float64) Generator {
	p := perlin.NewPerlin(alpha, beta, 3, time.Now().UnixNano())
	i := 0
	return func() float64 {
		value := p.Noise1D(float64(i) * 0.1)
		i++
		return offset + scale*value
	}
}

// TrendWithNoise algorithm generates a sequence of values that follow a
// linear trend with added random noise. This method simulates a sensor that
// produces readings which increase or decrease steadily with some random
// fluctuations.
//
// Arguments:
//   - startValue: The initial value of the sequence.
//   - step: The amount by which the value increases or decreases at each step.
//   - noiseScale: The amplitude of the random noise added to the trend.
//     Determines the intensity of the noise.
func TrendWithNoise(startValue, step, noiseScale float64) func() float64 {
	prevValue := startValue
	return func() float64 {
		newValue := prevValue + step
		// nolint:gosec
		// It's okay to use the default random number generator here.
		noise := (rand.Float64()*2 - 1) * noiseScale
		if step < 0 {
			newValue = min(newValue, newValue+noise)
		}
		newValue = max(newValue, newValue+noise)
		prevValue = newValue
		return newValue
	}
}

// clamp restricts a value to be within the specified range [min, max].
func clamp(value, minValue, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}
