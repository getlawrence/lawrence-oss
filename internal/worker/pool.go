package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/otlp"
	"github.com/getlawrence/lawrence-oss/internal/otlp/parser"
	"github.com/getlawrence/lawrence-oss/internal/otlp/processor"
	"github.com/getlawrence/lawrence-oss/internal/services"
	"go.uber.org/zap"
)

// TelemetryWriter defines the interface for writing telemetry data
type TelemetryWriter interface {
	WriteTraces(ctx context.Context, traces []otlp.TraceData) error
	WriteMetrics(ctx context.Context, sums []otlp.MetricSumData, gauges []otlp.MetricGaugeData, histograms []otlp.MetricHistogramData) error
	WriteLogs(ctx context.Context, logs []otlp.LogData) error
}

// WorkItemType represents the type of work item
type WorkItemType int

const (
	WorkItemTypeTraces WorkItemType = iota
	WorkItemTypeMetrics
	WorkItemTypeLogs
)

// WorkItem represents a single unit of work with raw OTLP bytes
type WorkItem struct {
	Type      WorkItemType
	RawData   []byte // Raw protobuf bytes
	Timestamp time.Time
}

// Pool represents a simple worker pool
type Pool struct {
	queue         chan WorkItem
	shutdown      chan struct{}
	wg            sync.WaitGroup
	writer        TelemetryWriter
	parser        *parser.OTLPParser
	enricher      *processor.Enricher
	logger        *zap.Logger
	queueSize     int
	submitTimeout time.Duration
}

// NewPool creates a new worker pool with a single worker
func NewPool(queueSize int, writer TelemetryWriter, agentService services.AgentService, logger *zap.Logger) *Pool {
	return &Pool{
		queue:         make(chan WorkItem, queueSize),
		shutdown:      make(chan struct{}),
		writer:        writer,
		parser:        parser.NewOTLPParser(logger),
		enricher:      processor.NewEnricher(agentService, logger),
		logger:        logger,
		queueSize:     queueSize,
		submitTimeout: 100 * time.Millisecond, // Timeout for submitting to queue
	}
}

// Start starts the worker pool
func (p *Pool) Start() {
	p.logger.Info("Starting worker pool", zap.Int("workers", 1), zap.Int("queue_size", p.queueSize))
	p.wg.Add(1)
	go p.worker()
}

// Stop gracefully stops the worker pool
func (p *Pool) Stop(timeout time.Duration) error {
	p.logger.Info("Stopping worker pool", zap.Duration("timeout", timeout))

	// Signal shutdown
	close(p.shutdown)

	// Wait for worker to finish with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		p.logger.Info("Worker pool stopped gracefully")
		return nil
	case <-time.After(timeout):
		p.logger.Warn("Worker pool shutdown timeout", zap.Int("remaining_items", len(p.queue)))
		return fmt.Errorf("shutdown timeout exceeded")
	}
}

// Submit submits a work item to the queue
func (p *Pool) Submit(item WorkItem) error {
	select {
	case p.queue <- item:
		return nil
	case <-time.After(p.submitTimeout):
		return fmt.Errorf("queue full, submit timeout")
	}
}

// QueueDepth returns the current queue depth
func (p *Pool) QueueDepth() int {
	return len(p.queue)
}

// worker is the main worker goroutine
func (p *Pool) worker() {
	defer p.wg.Done()

	p.logger.Info("Worker started")

	for {
		select {
		case item := <-p.queue:
			p.processItem(item)
		case <-p.shutdown:
			// Drain remaining items
			p.logger.Info("Draining remaining queue items", zap.Int("count", len(p.queue)))
			for {
				select {
				case item := <-p.queue:
					p.processItem(item)
				default:
					p.logger.Info("Worker stopped")
					return
				}
			}
		}
	}
}

// processItem processes a single work item
func (p *Pool) processItem(item WorkItem) {
	start := time.Now()
	ctx := context.Background()

	var err error
	var itemType string

	switch item.Type {
	case WorkItemTypeTraces:
		itemType = "traces"

		// Parse raw bytes
		traces, err := p.parser.ParseTraces(item.RawData)
		if err != nil {
			p.logger.Error("Failed to parse traces", zap.Error(err))
			return
		}

		// Enrich with group information
		p.enricher.EnrichTraces(ctx, traces)

		// Write to storage
		err = p.writer.WriteTraces(ctx, traces)
		p.logger.Debug("Processed traces",
			zap.Int("count", len(traces)),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))

	case WorkItemTypeMetrics:
		itemType = "metrics"

		// Parse raw bytes
		sums, gauges, histograms, err := p.parser.ParseMetrics(item.RawData)
		if err != nil {
			p.logger.Error("Failed to parse metrics", zap.Error(err))
			return
		}

		// Enrich with group information
		p.enricher.EnrichMetrics(ctx, sums, gauges, histograms)

		// Write to storage
		err = p.writer.WriteMetrics(ctx, sums, gauges, histograms)
		p.logger.Debug("Processed metrics",
			zap.Int("sums", len(sums)),
			zap.Int("gauges", len(gauges)),
			zap.Int("histograms", len(histograms)),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))

	case WorkItemTypeLogs:
		itemType = "logs"

		// Parse raw bytes
		logs, err := p.parser.ParseLogs(item.RawData)
		if err != nil {
			p.logger.Error("Failed to parse logs", zap.Error(err))
			return
		}

		// Enrich with group information
		p.enricher.EnrichLogs(ctx, logs)

		// Write to storage
		err = p.writer.WriteLogs(ctx, logs)
		p.logger.Debug("Processed logs",
			zap.Int("count", len(logs)),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err))
	}

	if err != nil {
		p.logger.Error("Failed to process work item",
			zap.String("type", itemType),
			zap.Error(err))
	}
}
