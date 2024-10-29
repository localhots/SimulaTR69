package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type Metrics struct {
	Registrer         prometheus.Registerer
	RequestFailures   prometheus.Counter
	ResponseStatus    prometheus_CounterVec
	ConnectionLatency prometheus.Histogram
	InformDuration    prometheus.Histogram
	ConcurrentInforms prometheus.Gauge
}

// Odd name, I know.
// nolint:revive,stylecheck
type prometheus_CounterVec interface {
	prometheus.Collector
	With(prometheus.Labels) prometheus.Counter
}

func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		RequestFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "acs_request_failures",
			Help: "Number of times requests to ACS failed.",
		}),
		ResponseStatus: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "acs_response_status",
				Help: "ACS response status code.",
			},
			[]string{"status"},
		),
		ConcurrentInforms: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "concurrent_informs",
			Help: "The number of inform sessions in progress.",
		}),
		ConnectionLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:                        "acs_connection_latency",
			Help:                        "Time it takes to establish a connection to ACS in milliseconds.",
			Buckets:                     prometheus.LinearBuckets(0.5, 0.5, 10),
			NativeHistogramMaxExemplars: 100,
			NativeHistogramExemplarTTL:  15 * time.Second,
		}),
		InformDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:                        "inform_duration",
			Help:                        "Inform length in milliseconds.",
			Buckets:                     prometheus.ExponentialBuckets(1, 10, 10),
			NativeHistogramMaxExemplars: 100,
			NativeHistogramExemplarTTL:  15 * time.Minute,
		}),
	}
	reg.MustRegister(
		m.RequestFailures,
		m.ResponseStatus,
		m.ConcurrentInforms,
		m.ConnectionLatency,
		m.InformDuration,
	)
	return m
}
