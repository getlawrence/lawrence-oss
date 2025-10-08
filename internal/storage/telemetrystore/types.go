package telemetrystore

// Re-export types from the types package for convenience
import "github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore/types"

// Type aliases for convenience
type Reader = types.Reader
type Writer = types.Writer
type Metric = types.Metric
type MetricType = types.MetricType
type Log = types.Log
type Trace = types.Trace
type MetricQuery = types.MetricQuery
type LogQuery = types.LogQuery
type TraceQuery = types.TraceQuery
type Rollup = types.Rollup
type RollupInterval = types.RollupInterval
type RollupQuery = types.RollupQuery

// Re-export constants
const (
	MetricTypeGauge     = types.MetricTypeGauge
	MetricTypeCounter   = types.MetricTypeCounter
	MetricTypeHistogram = types.MetricTypeHistogram
	RollupInterval1m    = types.RollupInterval1m
	RollupInterval5m    = types.RollupInterval5m
	RollupInterval1h    = types.RollupInterval1h
	RollupInterval1d    = types.RollupInterval1d
)
