# Noise Generators

The `noise` package provides various algorithms to generate sequences of values 
that simulate sensor readings with different patterns. These generators can be 
used to create realistic test data for applications that process sensor data.

## Usage

Add a special parameter to the datamodel file with type `sim:generator` and a
value that follows the syntax described here.

### Random Walk

The `randomWalk` algorithm generates a sequence of values where each value is 
derived from the previous one by adding a small random change. This simulates a 
sensor that produces readings which vary in a random but continuous manner.

![Random Walk](images/random_walk.png)

```
randomWalk(startValue=50, minValue=30, maxValue=70, step=5) as xsd:int
```

- `startValue`: The initial value of the sequence.
- `minValue`: The minimum value the sequence can take.
- `maxValue`: The maximum value the sequence can take.
- `step`: The maximum change between consecutive values.

### Piecewise Linear

The `piecewiseLinear` algorithm generates a sequence of values where the trend 
changes direction at regular intervals. This simulates a sensor that produces 
readings which follow a piecewise linear pattern, with occasional random 
fluctuations.

![Piecewise Linear](images/piecewise_linear.png)

```
piecewiseLinear(startValue=50, minValue=30, maxValue=70, step=5) as xsd:int
```

- `startValue`: The initial value of the sequence.
- `minValue`: The minimum value the sequence can take.
- `maxValue`: The maximum value the sequence can take.
- `step`: The maximum change between consecutive values.

### Sine Wave with Noise

The `sineWithNoise` algorithm generates a sequence of values based on a sine 
wave with added random noise. This simulates a sensor that produces readings 
which follow a sinusoidal pattern with some random fluctuations.

![Sine Wave with Noise](images/sine_wave_with_noise.png)

```
sineWithNoise(offset=50, amplitude=20, frequency=0.1, phase=0, noiseScale=5) as xsd:int
```

- `offset`: The baseline value of the sine wave.
- `amplitude`: The peak value of the sine wave.
- `frequency`: The number of cycles the sine wave completes in a unit time.
- `noiseScale`: The level of random noise added to the sine wave.

### Perlin Noise

The `perlinNoise` algorithm generates a sequence of values based on Perlin 
noise. This simulates a sensor that produces readings which follow a smooth, 
natural pattern.

![Perlin Noise](images/perlin_noise.png)

```
perlinNoise(offset=50, alpha=2, beta=2, seed=42, scale=40) as xsd:int
```

- `alpha`: Controls the smoothness of the Perlin noise. Higher values make the 
  noise smoother.
- `beta`: Controls the frequency of the Perlin noise. Higher values increase the
  frequency.
- `seed`: A seed value for the random number generator to ensure 
  reproducibility.
- `scale`: A scaling factor to adjust the amplitude of the noise.
- `offset`: A constant value to be added to the generated noise values.

## Experimenting

To experiment with the noise generator output and find the best algorithm and 
parameters for simulating your sensor, modify the `example_test.go` file and 
generate previews using the following command:

```
go test -tags=preview ./datamodel/noise
```