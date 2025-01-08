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
	noopRegisterer   struct{}
	noopCollector    struct{}
	noopCounter      struct{ noopCollector }
	noopCounterVec   struct{ noopCollector }
	noopHistogram    struct{ noopCollector }
	noopHistogramVec struct{ noopCollector }
	noopGauge        struct{ noopCollector }
	noopObserver     struct{}
)

func NewNoop() *Metrics {
	return &Metrics{
		Bootstrapped:        noopCounter{},
		ConnectionLatency:   noopHistogram{},
		MethodCalls:         noopCounterVec{},
		RequestFailures:     noopCounter{},
		ResponseStatus:      noopCounterVec{},
		SessionsAttempted:   noopCounter{},
		SessionsEstablished: noopCounter{},
		SessionsCompleted:   noopCounter{},
		SessionDuration:     noopHistogramVec{},
		ConcurrentSessions:  noopGauge{},
		InformEvents:        noopCounterVec{},
		ParametersRead:      noopCounter{},
		ParametersWritten:   noopCounter{},
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

func (noopCounterVec) With(prometheus.Labels) prometheus.Counter    { return noopCounter{} }
func (noopHistogramVec) With(prometheus.Labels) prometheus.Observer { return noopObserver{} }

func (noopCounter) Add(float64)      {}
func (noopObserver) Observe(float64) {}

func (noopHistogram) Observe(float64) {}

func (noopGauge) Add(float64)       {}
func (noopGauge) Set(float64)       {}
func (noopGauge) Inc()              {}
func (noopGauge) Dec()              {}
func (noopGauge) Sub(float64)       {}
func (noopGauge) SetToCurrentTime() {}
