package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

var (
	_ prometheus.Registerer = noopRegisterer{}
	_ prometheus_CounterVec = noopCounterVec{}
	_ prometheus.Counter    = noopCounter{}
	_ prometheus.Histogram  = noopHistogram{}
	_ prometheus.Gauge      = noopGauge{}
)

type (
	noopRegisterer struct{}
	noopCollector  struct{}
	noopCounter    struct{ noopCollector }
	noopCounterVec struct{ noopCollector }
	noopHistogram  struct{ noopCollector }
	noopGauge      struct{ noopCollector }
)

func NewNoop() *Metrics {
	return &Metrics{
		Registrer:         noopRegisterer{},
		RequestFailures:   noopCounter{},
		ResponseStatus:    noopCounterVec{},
		ConnectionLatency: noopHistogram{},
		InformDuration:    noopHistogram{},
		ConcurrentInforms: noopGauge{},
	}
}

func (noopRegisterer) Register(prometheus.Collector) error  { return nil }
func (noopRegisterer) MustRegister(...prometheus.Collector) {}
func (noopRegisterer) Unregister(prometheus.Collector) bool { return true }

func (noopCollector) Collect(chan<- prometheus.Metric)         {}
func (noopCollector) Desc() *prometheus.Desc                   { return nil }
func (noopCollector) Describe(chan<- *prometheus.Desc)         {}
func (noopCollector) Inc()                                     {}
func (noopCollector) Write(*io_prometheus_client.Metric) error { return nil }

func (noopCounterVec) With(prometheus.Labels) prometheus.Counter { return noopCounter{} }

func (noopCounter) Add(float64) {}

func (noopHistogram) Observe(float64) {}

func (noopGauge) Add(float64)       {}
func (noopGauge) Set(float64)       {}
func (noopGauge) Inc()              {}
func (noopGauge) Dec()              {}
func (noopGauge) Sub(float64)       {}
func (noopGauge) SetToCurrentTime() {}
