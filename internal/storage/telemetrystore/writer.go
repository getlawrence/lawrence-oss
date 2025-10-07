package telemetrystore

import (
	"context"

	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/pdata/plog"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

type Writer interface {
	WriteTraces(ctx context.Context, td ptrace.Traces) error
	WriteMetrics(ctx context.Context, md pmetric.Metrics) error
	WriteLogs(ctx context.Context, ld plog.Logs) error
}