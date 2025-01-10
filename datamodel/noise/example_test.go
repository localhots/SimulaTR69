//go:build preview
// +build preview

package noise

import (
	"fmt"
	"os"
	"testing"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
)

func TestGeneratePreviews(t *testing.T) {
	const steps = 200
	page := components.NewPage()
	page.SetLayout(components.PageFlexLayout)
	page.AddCharts(
		renderChart(t, "randomWalk(startValue=50, minValue=30, maxValue=70, step=5) as xsd:int", steps),
		renderChart(t, "perlinNoise(offset=50, alpha=2, beta=2, seed=42, scale=40) as xsd:int", steps),
		renderChart(t, "piecewiseLinear(startValue=50, minValue=30, maxValue=70, step=5) as xsd:int", steps),
		renderChart(t, "sineWithNoise(offset=50, amplitude=20, frequency=0.1, phase=0, noiseScale=5) as xsd:int", steps),
	)

	// Render the page as HTML
	fd, err := os.Create("examples.html")
	if err != nil {
		panic(err)
	}
	defer fd.Close()
	if err := page.Render(fd); err != nil {
		panic(err)
	}

	fmt.Println("Charts rendered successfully: examples.html")
}

// renderChart generates a chart from a Generator function
func renderChart(t *testing.T, gendef string, steps int) *charts.Line {
	t.Helper()

	genfn, err := ParseDef(gendef)
	if err != nil {
		t.Fatalf("Failed to parse generator definition: %v", err)
	}
	gen, err := genfn.Generator()
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	xAxis := make([]int, steps)
	yAxis := make([]opts.LineData, steps)

	for i := 0; i < steps; i++ {
		xAxis[i] = i
		yAxis[i] = opts.LineData{Value: gen()}
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTitleOpts(opts.Title{Title: genfn.FullName(), Subtitle: gendef}),
		charts.WithXAxisOpts(opts.XAxis{Name: "Time Step"}),
		charts.WithYAxisOpts(opts.YAxis{Name: ""}),
	)
	line.SetXAxis(xAxis).AddSeries("Sensor Data", yAxis)
	return line
}
