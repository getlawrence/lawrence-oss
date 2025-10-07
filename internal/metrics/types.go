// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package metrics

import "time"

// Counter tracks monotonically increasing values
type Counter interface {
	// Inc increments the counter by 1
	Inc(value int64)
}

// Gauge tracks values that can go up and down
type Gauge interface {
	// Update sets the gauge to an arbitrary value
	Update(value int64)
}

// Timer tracks durations and latencies
type Timer interface {
	// Record records a duration
	Record(duration time.Duration)
}

// Histogram tracks distribution of values
type Histogram interface {
	// Record records a value in the histogram
	Record(value float64)
}

// Options are used to specify metric options
type Options struct {
	Name string
	Tags map[string]string
	Help string
}

// TimerOptions are used to specify timer options
type TimerOptions struct {
	Name    string
	Tags    map[string]string
	Help    string
	Buckets []float64
}

// HistogramOptions are used to specify histogram options
type HistogramOptions struct {
	Name    string
	Tags    map[string]string
	Help    string
	Buckets []float64
}

// Factory creates metrics
type Factory interface {
	Counter(options Options) Counter
	Gauge(options Options) Gauge
	Timer(options TimerOptions) Timer
	Histogram(options HistogramOptions) Histogram
}

// NullFactory is a no-op metrics factory
var NullFactory Factory = &nullFactory{}

type nullFactory struct{}

func (*nullFactory) Counter(Options) Counter {
	return &nullCounter{}
}

func (*nullFactory) Gauge(Options) Gauge {
	return &nullGauge{}
}

func (*nullFactory) Timer(TimerOptions) Timer {
	return &nullTimer{}
}

func (*nullFactory) Histogram(HistogramOptions) Histogram {
	return &nullHistogram{}
}

type nullCounter struct{}

func (*nullCounter) Inc(int64) {}

type nullGauge struct{}

func (*nullGauge) Update(int64) {}

type nullTimer struct{}

func (*nullTimer) Record(time.Duration) {}

type nullHistogram struct{}

func (*nullHistogram) Record(float64) {}
