package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/api"
	"github.com/getlawrence/lawrence-oss/internal/config"
	"github.com/getlawrence/lawrence-oss/internal/metrics"
	"github.com/getlawrence/lawrence-oss/internal/opamp"
	"github.com/getlawrence/lawrence-oss/internal/otlp/receiver"
	"github.com/getlawrence/lawrence-oss/internal/services"
	"github.com/getlawrence/lawrence-oss/internal/storage/applicationstore"
	"github.com/getlawrence/lawrence-oss/internal/storage/telemetrystore"
	"github.com/getlawrence/lawrence-oss/internal/utils"
)

const (
	appName = "Lawrence OSS"
	version = "0.1.0"
)

func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "lawrence",
		Short: "Lawrence OSS - OpenTelemetry observability platform",
		Long: `Lawrence OSS is a comprehensive observability platform that provides:
- OpenTelemetry data collection and processing
- Agent management via OpAMP protocol
- Real-time telemetry analysis
- Modern web interface for monitoring and management`,
		RunE: runLawrence,
	}

	// Add subcommands
	rootCmd.AddCommand(versionCommand())
	rootCmd.AddCommand(configCommand())

	// Add flags
	rootCmd.PersistentFlags().String("config", "./lawrence.yaml", "Path to configuration file")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "json", "Log format (json, console)")

	// Bind flags to viper
	_ = viper.BindPFlags(rootCmd.PersistentFlags())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runLawrence(cmd *cobra.Command, args []string) error {
	// Load configuration
	configPath := viper.GetString("config")
	config, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger, err := utils.NewLogger(config.Logging.Level, config.Logging.Format)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer func() { _ = logger.Sync() }()

	logger.Info("Starting Lawrence OSS",
		zap.String("version", version),
		zap.String("config", configPath))

	// Initialize storage factories
	logger.Info("Initializing storage layer")

	// Create application store using meta factory
	appStoreFactory, err := applicationstore.NewFactoryFromAppConfig(config)
	if err != nil {
		logger.Fatal("Failed to create application store factory", zap.Error(err))
	}

	// Initialize the factory
	if err := appStoreFactory.Initialize(logger); err != nil {
		logger.Fatal("Failed to initialize application store factory", zap.Error(err))
	}

	// Create application store
	appStore, err := appStoreFactory.CreateApplicationStore()
	if err != nil {
		logger.Fatal("Failed to create application store", zap.Error(err))
	}

	// Ensure application store factory is properly closed on shutdown
	defer func() {
		if err := appStoreFactory.Close(); err != nil {
			logger.Error("Failed to close application store factory", zap.Error(err))
		}
	}()

	// Create telemetry store using meta factory
	telemetryStoreFactory, err := telemetrystore.NewFactoryFromAppConfig(config)
	if err != nil {
		logger.Fatal("Failed to create telemetry store factory", zap.Error(err))
	}

	// Initialize the factory
	if err := telemetryStoreFactory.Initialize(logger); err != nil {
		logger.Fatal("Failed to initialize telemetry store factory", zap.Error(err))
	}

	// Create telemetry reader
	telemetryReader, err := telemetryStoreFactory.CreateTelemetryReader()
	if err != nil {
		logger.Fatal("Failed to create telemetry reader", zap.Error(err))
	}

	// Create writer adapter for OTLP receivers (handles both sync and async writes)
	telemetryWriter, err := telemetryStoreFactory.CreateTelemetryWriter()
	if err != nil {
		logger.Fatal("Failed to create telemetry writer", zap.Error(err))
	}

	// Ensure telemetry store factory is properly closed on shutdown
	defer func() {
		if err := telemetryStoreFactory.Close(); err != nil {
			logger.Error("Failed to close telemetry store factory", zap.Error(err))
		}
	}()

	// Initialize metrics
	logger.Info("Initializing metrics")
	registry := prometheus.NewRegistry()
	metricsFactory := metrics.NewPrometheusFactory("lawrence", registry)
	opampMetrics := metrics.NewOpAMPMetrics(metricsFactory)
	otlpMetrics := metrics.NewOTLPMetrics(metricsFactory)

	// Initialize OpAMP components
	logger.Info("Initializing OpAMP server")
	agents := opamp.NewAgents(logger)

	// Determine which OTLP endpoints to offer to agents
	// If agent_*_endpoint is configured, use it; otherwise use the receiver endpoint
	agentGRPCEndpoint := config.OTLP.AgentGRPCEndpoint
	if agentGRPCEndpoint == "" {
		agentGRPCEndpoint = config.OTLP.GRPCEndpoint
	}
	agentHTTPEndpoint := config.OTLP.AgentHTTPEndpoint
	if agentHTTPEndpoint == "" {
		agentHTTPEndpoint = config.OTLP.HTTPEndpoint
	}

	// Initialize service layer
	logger.Info("Initializing service layer")

	// Create agent service (without config sender initially to break circular dependency)
	agentService := services.NewAgentService(appStore, logger)

	// Create OpAMP server with agent service (for persistence)
	opampServer, err := opamp.NewServer(agents, agentService, opampMetrics, agentGRPCEndpoint, agentHTTPEndpoint, logger)
	if err != nil {
		logger.Fatal("Failed to create OpAMP server", zap.Error(err))
	}

	// Create config sender and inject into agent service (breaks circular dependency)
	configSender := opamp.NewConfigSender(agents, logger)
	agentService.(*services.AgentServiceImpl).SetConfigSender(configSender)

	// Create telemetry query service
	telemetryService := services.NewTelemetryQueryService(telemetryReader, agentService, logger)

	// Start OpAMP server
	if err := opampServer.Start(config.Server.OpAMPPort); err != nil {
		logger.Fatal("Failed to start OpAMP server", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = opampServer.Stop(ctx)
	}()

	// Initialize OTLP receivers
	grpcServer, err := receiver.NewGRPCServer(4317, telemetryWriter, otlpMetrics, logger)
	if err != nil {
		logger.Fatal("Failed to create gRPC server", zap.Error(err))
	}
	if err := grpcServer.Start(); err != nil {
		logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = grpcServer.Stop(ctx)
	}()

	httpServer, err := receiver.NewHTTPServer(4318, telemetryWriter, otlpMetrics, logger)
	if err != nil {
		logger.Fatal("Failed to create HTTP server", zap.Error(err))
	}
	if err := httpServer.Start(); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpServer.Stop(ctx)
	}()

	// Initialize HTTP API server
	logger.Info("Initializing HTTP API server")
	apiServer := api.NewServer(agentService, telemetryService, logger)

	// Start API server in a goroutine
	go func() {
		if err := apiServer.Start(fmt.Sprintf("%d", config.Server.HTTPPort)); err != nil {
			logger.Fatal("Failed to start API server", zap.Error(err))
		}
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = apiServer.Stop(ctx)
	}()

	// Start background services
	go startRollupGenerator(telemetryService, config, logger)
	go startCleanupTask(telemetryService, config, logger)

	logger.Info("Lawrence OSS is running",
		zap.Int("opamp_port", config.Server.OpAMPPort),
		zap.Int("otlp_grpc_port", 4317),
		zap.Int("otlp_http_port", 4318),
		zap.Int("api_port", config.Server.HTTPPort))

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down Lawrence OSS...")
	return nil
}

// versionCommand returns the version subcommand
func versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s v%s\n", appName, version)
		},
	}
}

// configCommand returns the config subcommand
func configCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Print current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			configPath := viper.GetString("config")
			_, err := config.LoadConfig(configPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
				os.Exit(1)
			}
			// TODO: Pretty print configuration
			fmt.Printf("Configuration loaded from: %s\n", configPath)
		},
	}
}

// startRollupGenerator periodically generates rollups for metrics
func startRollupGenerator(telemetryService services.TelemetryQueryService, config *config.Config, logger *zap.Logger) {
	if !config.Rollups.Enabled {
		logger.Info("Rollup generation is disabled")
		return
	}

	logger.Info("Starting rollup generator")
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		now := time.Now()

		// Generate rollups based on time intervals
		if err := generateRollup(ctx, telemetryService, "1m", now, logger); err != nil {
			logger.Error("Failed to generate 1m rollup", zap.Error(err))
		}

		if now.Minute()%5 == 0 {
			if err := generateRollup(ctx, telemetryService, "5m", now, logger); err != nil {
				logger.Error("Failed to generate 5m rollup", zap.Error(err))
			}
		}

		if now.Minute() == 0 {
			if err := generateRollup(ctx, telemetryService, "1h", now, logger); err != nil {
				logger.Error("Failed to generate 1h rollup", zap.Error(err))
			}
		}
	}
}

// generateRollup generates a single rollup for the given interval
func generateRollup(ctx context.Context, telemetryService services.TelemetryQueryService, interval string, now time.Time, logger *zap.Logger) error {
	// TODO: Implement rollup generation
	logger.Debug("Generating rollup", zap.String("interval", interval))
	return nil
}

// startCleanupTask periodically cleans up old data
func startCleanupTask(telemetryService services.TelemetryQueryService, config *config.Config, logger *zap.Logger) {
	logger.Info("Starting cleanup task")
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		retention := 24 * time.Hour // TODO: Parse from config

		if err := telemetryService.CleanupOldData(ctx, retention); err != nil {
			logger.Error("Failed to cleanup old telemetry data", zap.Error(err))
		} else {
			logger.Debug("Cleaned up old telemetry data", zap.Duration("retention", retention))
		}
	}
}
