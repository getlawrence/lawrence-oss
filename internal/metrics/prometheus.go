// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PrometheusFactory creates Prometheus-backed metrics
type PrometheusFactory struct {
	namespace string
	registry  prometheus.Registerer
}

// NewPrometheusFactory creates a new Prometheus metrics factory
func NewPrometheusFactory(namespace string, registry prometheus.Registerer) *PrometheusFactory {
	if registry == nil {
		registry = prometheus.DefaultRegisterer
	}
	return &PrometheusFactory{
		namespace: namespace,
		registry:  registry,
	}
}

func (f *PrometheusFactory) Counter(options Options) Counter {
	opts := prometheus.CounterOpts{
		Namespace: f.namespace,
		Name:      options.Name,
		Help:      options.Help,
		ConstLabels: options.Tags,
	}
	counter := promauto.With(f.registry).NewCounter(opts)
	return &promCounter{counter: counter}
}

func (f *PrometheusFactory) Gauge(options Options) Gauge {
	opts := prometheus.GaugeOpts{
		Namespace: f.namespace,
		Name:      options.Name,
		Help:      options.Help,
		ConstLabels: options.Tags,
	}
	gauge := promauto.With(f.registry).NewGauge(opts)
	return &promGauge{gauge: gauge}
}

func (f *PrometheusFactory) Timer(options TimerOptions) Timer {
	opts := prometheus.HistogramOpts{
		Namespace: f.namespace,
		Name:      options.Name,
		Help:      options.Help,
		ConstLabels: options.Tags,
		Buckets:   options.Buckets,
	}
	if len(opts.Buckets) == 0 {
		// Default buckets for latency in seconds
		opts.Buckets = []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
	}
	histogram := promauto.With(f.registry).NewHistogram(opts)
	return &promTimer{histogram: histogram}
}

func (f *PrometheusFactory) Histogram(options HistogramOptions) Histogram {
	opts := prometheus.HistogramOpts{
		Namespace: f.namespace,
		Name:      options.Name,
		Help:      options.Help,
		ConstLabels: options.Tags,
		Buckets:   options.Buckets,
	}
	if len(opts.Buckets) == 0 {
		opts.Buckets = prometheus.DefBuckets
	}
	histogram := promauto.With(f.registry).NewHistogram(opts)
	return &promHistogram{histogram: histogram}
}

// promCounter wraps a Prometheus counter
type promCounter struct {
	counter prometheus.Counter
}

func (c *promCounter) Inc(value int64) {
	c.counter.Add(float64(value))
}

// promGauge wraps a Prometheus gauge
type promGauge struct {
	gauge prometheus.Gauge
}

func (g *promGauge) Update(value int64) {
	g.gauge.Set(float64(value))
}

// promTimer wraps a Prometheus histogram for timing
type promTimer struct {
	histogram prometheus.Histogram
}

func (t *promTimer) Record(duration time.Duration) {
	t.histogram.Observe(duration.Seconds())
}

// promHistogram wraps a Prometheus histogram
type promHistogram struct {
	histogram prometheus.Histogram
}

func (h *promHistogram) Record(value float64) {
	h.histogram.Observe(value)
}
