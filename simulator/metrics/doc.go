// Package metrics provides a set of Prometheus metrics for monitoring the
// performance and behavior of a TR-069 simulator. This package includes
// counters, histograms, and gauges to track various aspects such as session
// attempts, connection latency, method calls, and parameter interactions.
// The metrics collected here are relevant for understanding the operational
// characteristics and performance of the TR-069 simulator. Additionally, this
// package includes a no-op implementation for cases where metrics collection
// is not required or needs to be disabled.
package metrics
