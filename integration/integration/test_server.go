// Copyright (c) 2024 Lawrence OSS Contributors
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/getlawrence/lawrence-oss/internal/api"
	"github.com/getlawrence/lawrence-oss/internal/metrics"
	"github.com/getlawrence/lawrence-oss/internal/opamp"
	"github.com/getlawrence/lawrence-oss/internal/otlp/processor"
	"github.com/getlawrence/lawrence-oss/internal/otlp/receiver"
	"github.com/getlawrence/lawrence-oss/internal/services"
	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore"
	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/memory"
	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore/sqlite"
	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore"
	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore/duckdb"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// TestServer represents a test instance of Lawrence OSS for integration testing
type TestServer struct {
	// Configuration
	HTTPPort     int
	OpAMPPort    int
	OTLPGRPCPort int
	OTLPHTTPPort int

	// Storage
	appStoreFactory       applicationstore.ApplicationStoreFactory
	telemetryStoreFactory telemetrystore.TelemetryStoreFactory
	telemetryReader       telemetrystore.Reader
	telemetryWriter       telemetrystore.Writer

	// Services
	agentService     services.AgentService
	telemetryService services.TelemetryQueryService

	// Servers
	apiServer   *api.Server
	opampServer *opamp.Server
	grpcServer  *receiver.GRPCServer
	httpServer  *receiver.HTTPServer

	// Metrics
	opampMetrics *metrics.OpAMPMetrics
	otlpMetrics  *metrics.OTLPMetrics

	// Utilities
	logger  *zap.Logger
	baseURL string
	tempDir string
	t       *testing.T
}

// NewTestServer creates a new test server instance
func NewTestServer(t *testing.T, useMemory bool) *TestServer {
	logger := zap.NewNop()

	// Create temp directory for test databases
	tempDir := t.TempDir()

	ts := &TestServer{
		HTTPPort:     findFreePort(),
		OpAMPPort:    findFreePort(),
		OTLPGRPCPort: findFreePort(),
		OTLPHTTPPort: findFreePort(),
		logger:       logger,
		tempDir:      tempDir,
		t:            t,
	}

	ts.baseURL = fmt.Sprintf("http://localhost:%d", ts.HTTPPort)

	// Initialize storage
	if useMemory {
		ts.initMemoryStorage()
	} else {
		ts.initDatabaseStorage()
	}

	// Initialize metrics
	ts.initMetrics()

	// Initialize services
	ts.initServices()

	// Initialize servers
	ts.initServers()

	return ts
}

// initMemoryStorage initializes in-memory storage (fastest for tests)
func (ts *TestServer) initMemoryStorage() {
	// Application store
	appFactory := memory.NewFactory()
	if err := appFactory.Initialize(ts.logger); err != nil {
		ts.t.Fatalf("Failed to initialize memory app store: %v", err)
	}
	ts.appStoreFactory = appFactory

	// For telemetry, use a temp file for DuckDB
	telemetryDBPath := filepath.Join(ts.tempDir, "telemetry-mem.db")
	telemetryFactory := duckdb.NewFactory(telemetryDBPath)
	if err := telemetryFactory.Initialize(ts.logger); err != nil {
		ts.t.Fatalf("Failed to initialize memory telemetry store: %v", err)
	}
	ts.telemetryStoreFactory = telemetryFactory

	var err error
	ts.telemetryReader, err = telemetryFactory.CreateTelemetryReader()
	if err != nil {
		ts.t.Fatalf("Failed to create telemetry reader: %v", err)
	}

	ts.telemetryWriter, err = telemetryFactory.CreateTelemetryWriter()
	if err != nil {
		ts.t.Fatalf("Failed to create telemetry writer: %v", err)
	}
}

// initMetrics initializes metrics components
func (ts *TestServer) initMetrics() {
	registry := prometheus.NewRegistry()
	metricsFactory := metrics.NewPrometheusFactory("lawrence", registry)
	ts.opampMetrics = metrics.NewOpAMPMetrics(metricsFactory)
	ts.otlpMetrics = metrics.NewOTLPMetrics(metricsFactory)
}

// initDatabaseStorage initializes file-based storage
func (ts *TestServer) initDatabaseStorage() {
	appDBPath := filepath.Join(ts.tempDir, "app.db")
	telemetryDBPath := filepath.Join(ts.tempDir, "telemetry.db")

	// Application store
	appFactory := sqlite.NewFactory(appDBPath)
	if err := appFactory.Initialize(ts.logger); err != nil {
		ts.t.Fatalf("Failed to initialize SQLite app store: %v", err)
	}
	ts.appStoreFactory = appFactory

	// Telemetry store
	telemetryFactory := duckdb.NewFactory(telemetryDBPath)
	if err := telemetryFactory.Initialize(ts.logger); err != nil {
		ts.t.Fatalf("Failed to initialize DuckDB telemetry store: %v", err)
	}
	ts.telemetryStoreFactory = telemetryFactory

	var err error
	ts.telemetryReader, err = telemetryFactory.CreateTelemetryReader()
	if err != nil {
		ts.t.Fatalf("Failed to create telemetry reader: %v", err)
	}

	ts.telemetryWriter, err = telemetryFactory.CreateTelemetryWriter()
	if err != nil {
		ts.t.Fatalf("Failed to create telemetry writer: %v", err)
	}
}

// initServices initializes service layer
func (ts *TestServer) initServices() {
	appStore, err := ts.appStoreFactory.CreateApplicationStore()
	if err != nil {
		ts.t.Fatalf("Failed to create app store: %v", err)
	}

	// Create agent service without config sender initially
	ts.agentService = services.NewAgentService(appStore, ts.logger)
}

// initServers initializes all servers
func (ts *TestServer) initServers() {
	// OpAMP Server components
	agents := opamp.NewAgents(ts.logger)
	opampServer, err := opamp.NewServer(agents, ts.agentService, ts.opampMetrics, "localhost:4317", "localhost:4318", ts.logger)
	if err != nil {
		ts.t.Fatalf("Failed to create OpAMP server: %v", err)
	}
	ts.opampServer = opampServer

	// Create config sender (separate concern from AgentService)
	configSender := opamp.NewConfigSender(agents, ts.logger)

	// Create telemetry service
	ts.telemetryService = services.NewTelemetryQueryService(ts.telemetryReader, ts.agentService, ts.logger)

	// API Server
	ts.apiServer = api.NewServer(ts.agentService, ts.telemetryService, configSender, ts.logger)

	// Create processor enricher for telemetry
	enricher := processor.NewEnricher(ts.agentService, ts.logger)

	// OTLP Receivers - use telemetry writer with enricher
	grpcServer, err := receiver.NewGRPCServer(ts.OTLPGRPCPort, ts.telemetryWriter, enricher, ts.otlpMetrics, ts.logger)
	if err != nil {
		ts.t.Fatalf("Failed to create gRPC server: %v", err)
	}
	ts.grpcServer = grpcServer

	httpServer, err := receiver.NewHTTPServer(ts.OTLPHTTPPort, ts.telemetryWriter, enricher, ts.otlpMetrics, ts.logger)
	if err != nil {
		ts.t.Fatalf("Failed to create HTTP server: %v", err)
	}
	ts.httpServer = httpServer
}

// Start starts all servers
func (ts *TestServer) Start() {
	// Start API server
	go func() {
		if err := ts.apiServer.Start(fmt.Sprintf("%d", ts.HTTPPort)); err != nil && err != http.ErrServerClosed {
			ts.t.Logf("API server error: %v", err)
		}
	}()

	// Start OpAMP server
	if err := ts.opampServer.Start(ts.OpAMPPort); err != nil {
		ts.t.Fatalf("Failed to start OpAMP server: %v", err)
	}

	// Start OTLP receivers
	if err := ts.grpcServer.Start(); err != nil {
		ts.t.Fatalf("Failed to start gRPC server: %v", err)
	}

	if err := ts.httpServer.Start(); err != nil {
		ts.t.Fatalf("Failed to start HTTP server: %v", err)
	}

	// Wait for servers to be ready
	ts.WaitForReady()
}

// Stop stops all servers
func (ts *TestServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if ts.apiServer != nil {
		_ = ts.apiServer.Stop(ctx)
	}

	if ts.opampServer != nil {
		_ = ts.opampServer.Stop(ctx)
	}

	if ts.grpcServer != nil {
		_ = ts.grpcServer.Stop(ctx)
	}

	if ts.httpServer != nil {
		_ = ts.httpServer.Stop(ctx)
	}

	// Close storage factories
	if closer, ok := ts.appStoreFactory.(applicationstore.Closer); ok {
		closer.Close()
	}
	if closer, ok := ts.telemetryStoreFactory.(interface{ Close() error }); ok {
		closer.Close()
	}

	// Clean up temp directory
	os.RemoveAll(ts.tempDir)
}

// WaitForReady waits for the server to be ready to accept requests
func (ts *TestServer) WaitForReady() {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(ts.baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	ts.t.Fatal("Server did not become ready in time")
}

// GET makes an HTTP GET request
func (ts *TestServer) GET(path string) (*http.Response, error) {
	return http.Get(ts.baseURL + path)
}

// POST makes an HTTP POST request
func (ts *TestServer) POST(path string, contentType string, body io.Reader) (*http.Response, error) {
	return http.Post(ts.baseURL+path, contentType, body)
}

// DELETE makes an HTTP DELETE request
func (ts *TestServer) DELETE(path string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", ts.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

// findFreePort finds an available port
func findFreePort() int {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}
