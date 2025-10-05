package otlp

import (
	"time"
)

// LogData represents log data for storage insertion
type LogData struct {
	Timestamp          time.Time
	TraceId            string
	SpanId             string
	TraceFlags         uint32
	SeverityText       string
	SeverityNumber     int32
	ServiceName        string
	Body               string
	ResourceSchemaUrl  string
	ResourceAttributes map[string]string
	ScopeSchemaUrl     string
	ScopeName          string
	ScopeVersion       string
	ScopeAttributes    map[string]string
	LogAttributes      map[string]string
	AgentID            string // OSS: Agent ID instead of OrganizationID
	GroupID            string
	GroupName          string
}

// TraceData represents trace data for storage insertion
type TraceData struct {
	Timestamp          time.Time
	TraceId            string
	SpanId             string
	ParentSpanId       string
	TraceState         string
	SpanName           string
	SpanKind           int32
	ServiceName        string
	ResourceAttributes map[string]string
	ScopeName          string
	ScopeVersion       string
	ScopeAttributes    map[string]string
	SpanAttributes     map[string]string
	Duration           int64
	StatusCode         string
	StatusMessage      string
	Events             []EventData
	Links              []LinkData
	AgentID            string // OSS: Agent ID instead of OrganizationID
	GroupID            string
	GroupName          string
}

// EventData represents span event data
type EventData struct {
	Name       string
	Timestamp  time.Time
	Attributes map[string]string
}

// LinkData represents span link data
type LinkData struct {
	TraceId    string
	SpanId     string
	TraceState string
	Attributes map[string]string
}

// MetricSumData represents sum/counter metric data for storage insertion
type MetricSumData struct {
	ResourceAttributes     map[string]string
	ResourceSchemaUrl      string
	ScopeName              string
	ScopeVersion           string
	ScopeAttributes        map[string]string
	ScopeDroppedAttrCount  uint32
	ScopeSchemaUrl         string
	ServiceName            string
	MetricName             string
	MetricDescription      string
	MetricUnit             string
	Attributes             map[string]string
	StartTimeUnix          time.Time
	TimeUnix               time.Time
	Value                  float64
	Flags                  uint32
	AggregationTemporality int32
	IsMonotonic            bool
	AgentID                string // OSS: Agent ID instead of OrganizationID
	GroupID                string
	GroupName              string
}

// MetricGaugeData represents gauge metric data for storage insertion
type MetricGaugeData struct {
	ResourceAttributes     map[string]string
	ResourceSchemaUrl      string
	ScopeName              string
	ScopeVersion           string
	ScopeAttributes        map[string]string
	ScopeDroppedAttrCount  uint32
	ScopeSchemaUrl         string
	ServiceName            string
	MetricName             string
	MetricDescription      string
	MetricUnit             string
	Attributes             map[string]string
	StartTimeUnix          time.Time
	TimeUnix               time.Time
	Value                  float64
	Flags                  uint32
	AggregationTemporality int32
	IsMonotonic            bool
	AgentID                string // OSS: Agent ID instead of OrganizationID
	GroupID                string
	GroupName              string
}

// MetricHistogramData represents histogram metric data for storage insertion
type MetricHistogramData struct {
	ResourceAttributes     map[string]string
	ResourceSchemaUrl      string
	ScopeName              string
	ScopeVersion           string
	ScopeAttributes        map[string]string
	ScopeDroppedAttrCount  uint32
	ScopeSchemaUrl         string
	ServiceName            string
	MetricName             string
	MetricDescription      string
	MetricUnit             string
	Attributes             map[string]string
	StartTimeUnix          time.Time
	TimeUnix               time.Time
	Count                  uint64
	Sum                    float64
	BucketCounts           []uint64
	ExplicitBounds         []float64
	Flags                  uint32
	Min                    float64
	Max                    float64
	AggregationTemporality int32
	DataType               string
	AgentID                string // OSS: Agent ID instead of OrganizationID
	GroupID                string
	GroupName              string
}
