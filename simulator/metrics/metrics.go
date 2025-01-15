// Package metrics provides a set of Prometheus metrics for monitoring the
// performance and behavior of a TR-069 simulator. This package includes
// counters, histograms, and gauges to track various aspects such as session
// attempts, connection latency, method calls, and parameter interactions.
// The metrics collected here are relevant for understanding the operational
// characteristics and performance of the TR-069 simulator. Additionally, this
// package includes a no-op implementation for cases where metrics collection
// is not required or needs to be disabled.
package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds Prometheus metrics for monitoring the TR-069 simulator's
// performance and behavior. It includes counters, histograms, and gauges
// to track session attempts, connection latency, method calls, and more.
type Metrics struct {
	Bootstrapped        prometheus.Counter
	ConnectionLatency   prometheus.Histogram
	MethodCalls         prometheus_CounterVec
	RequestFailures     prometheus.Counter
	ResponseStatus      prometheus_CounterVec
	SessionsAttempted   prometheus.Counter
	SessionsEstablished prometheus.Counter
	SessionsCompleted   prometheus.Counter
	SessionDuration     prometheus_HistogramVec
	ConcurrentSessions  prometheus.Gauge
	InformEvents        prometheus_CounterVec
	ParametersRead      prometheus.Counter
	ParametersWritten   prometheus.Counter
}

// prometheus.CounterVec is a struct, not an interface. We can't reimplement it
// so instead a custom interface is defined.
// nolint:revive,stylecheck
type prometheus_CounterVec interface {
	prometheus.Collector
	With(prometheus.Labels) prometheus.Counter
}

// prometheus.HistogramVec is a struct, not an interface. We can't reimplement
// it so instead a custom interface is defined.
// nolint:revive,stylecheck
type prometheus_HistogramVec interface {
	prometheus.Collector
	With(prometheus.Labels) prometheus.Observer
}

// New creates and registers a new Metrics instance with the provided
// Prometheus registerer.
func New(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		Bootstrapped: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "bootstrapped",
			Help: "Number of successful bootstraps",
		}),
		ConnectionLatency: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:                        "acs_connection_latency",
			Help:                        "Time it takes to establish a connection to ACS in milliseconds",
			Buckets:                     prometheus.LinearBuckets(0.5, 0.5, 10),
			NativeHistogramMaxExemplars: 100,
			NativeHistogramExemplarTTL:  15 * time.Second,
		}),
		MethodCalls: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "rpc_method_calls",
				Help: "Number of RPC method calls",
			},
			[]string{"method"},
		),
		RequestFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "acs_request_failures",
			Help: "Number of times requests to ACS failed",
		}),
		ResponseStatus: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "acs_response_status",
				Help: "ACS response status code",
			},
			[]string{"status"},
		),
		SessionsAttempted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sessions_attempted",
			Help: "Number of attempted sessions",
		}),
		SessionsEstablished: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sessions_established",
			Help: "Number of established sessions",
		}),
		SessionsCompleted: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "sessions_completed",
			Help: "Number of completed sessions",
		}),
		SessionDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:                        "session_duration_per_event",
			Help:                        "Session length in milliseconds",
			Buckets:                     prometheus.ExponentialBuckets(1, 10, 10),
			NativeHistogramMaxExemplars: 100,
			NativeHistogramExemplarTTL:  15 * time.Minute,
		}, []string{"event"}),
		ConcurrentSessions: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "concurrent_sessions",
			Help: "The number of concurrent in-progress sessions",
		}),
		InformEvents: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "inform_events",
				Help: "Inform event count by event",
			},
			[]string{"event"},
		),
		ParametersRead: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "parameters_read",
			Help: "Number of parameters accessed via GetParameterValues",
		}),
		ParametersWritten: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "parameters_written",
			Help: "Number of parameters changed via SetParameterValues",
		}),
	}
	reg.MustRegister(
		m.Bootstrapped,
		m.ConnectionLatency,
		m.MethodCalls,
		m.RequestFailures,
		m.ResponseStatus,
		m.SessionsAttempted,
		m.SessionsEstablished,
		m.SessionsCompleted,
		m.SessionDuration,
		m.ConcurrentSessions,
		m.InformEvents,
		m.ParametersRead,
		m.ParametersWritten,
	)
	return m
}
