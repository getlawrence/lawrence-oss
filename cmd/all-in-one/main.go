package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/getlawrence/lawrence-oss/internal/api"
	"github.com/getlawrence/lawrence-oss/internal/app"
	"github.com/getlawrence/lawrence-oss/internal/opamp"
	"github.com/getlawrence/lawrence-oss/internal/otlp/receiver"
	"github.com/getlawrence/lawrence-oss/internal/storage"
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
	rootCmd.AddCommand(envCommand())
	rootCmd.AddCommand(statusCommand())
	rootCmd.AddCommand(configCommand())

	// Add flags
	rootCmd.PersistentFlags().String("config", "./lawrence.yaml", "Path to configuration file")
	rootCmd.PersistentFlags().String("log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().String("log-format", "json", "Log format (json, console)")

	// Bind flags to viper
	viper.BindPFlags(rootCmd.PersistentFlags())

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func runLawrence(cmd *cobra.Command, args []string) error {
	// Load configuration
	configPath := viper.GetString("config")
	config, err := app.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize logger
	logger, err := utils.NewLogger(config.Logging.Level, config.Logging.Format)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}
	defer logger.Sync()

	logger.Info("Starting Lawrence OSS",
		zap.String("version", version),
		zap.String("config", configPath))

	// Initialize storage container
	logger.Info("Initializing storage layer")
	storageContainer, err := storage.NewContainer(
		config.Storage.App.Path,
		config.Storage.Telemetry.Path,
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to initialize storage", zap.Error(err))
	}
	defer storageContainer.Close()

	// Create telemetry writer for OTLP receivers
	writer := storageContainer.GetTelemetryWriter()
	if writer == nil {
		logger.Fatal("Failed to get telemetry writer from storage container")
	}

	// Cast to the correct type
	telemetryWriter, ok := writer.(receiver.TelemetryWriter)
	if !ok {
		logger.Fatal("Telemetry writer does not implement TelemetryWriter interface")
	}

	// Initialize OpAMP server
	logger.Info("Initializing OpAMP server")
	agents := opamp.NewAgents(logger)
	opampServer, err := opamp.NewServer(agents, logger)
	if err != nil {
		logger.Fatal("Failed to create OpAMP server", zap.Error(err))
	}
	if err := opampServer.Start(config.Server.OpAMPPort); err != nil {
		logger.Fatal("Failed to start OpAMP server", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		opampServer.Stop(ctx)
	}()

	// Initialize OTLP receivers
	grpcServer, err := receiver.NewGRPCServer(4317, telemetryWriter, logger)
	if err != nil {
		logger.Fatal("Failed to create gRPC server", zap.Error(err))
	}
	if err := grpcServer.Start(); err != nil {
		logger.Fatal("Failed to start gRPC server", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		grpcServer.Stop(ctx)
	}()

	httpServer, err := receiver.NewHTTPServer(4318, telemetryWriter, logger)
	if err != nil {
		logger.Fatal("Failed to create HTTP server", zap.Error(err))
	}
	if err := httpServer.Start(); err != nil {
		logger.Fatal("Failed to start HTTP server", zap.Error(err))
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		httpServer.Stop(ctx)
	}()

	// Initialize HTTP API server
	logger.Info("Initializing HTTP API server")
	apiServer := api.NewServer(storageContainer, logger)

	// Start API server in a goroutine
	go func() {
		if err := apiServer.Start(fmt.Sprintf("%d", config.Server.HTTPPort)); err != nil {
			logger.Fatal("Failed to start API server", zap.Error(err))
		}
	}()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		apiServer.Stop(ctx)
	}()

	// Start background services
	go startRollupGenerator(storageContainer, config, logger)
	go startCleanupTask(storageContainer, config, logger)

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

// envCommand returns the environment subcommand
func envCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "env",
		Short: "Print environment variables",
		Run: func(cmd *cobra.Command, args []string) {
			for _, env := range os.Environ() {
				fmt.Println(env)
			}
		},
	}
}

// statusCommand returns the status subcommand
func statusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Print service status",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement health check
			fmt.Println("Service status: Running")
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
			_, err := app.LoadConfig(configPath)
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
func startRollupGenerator(container *storage.Container, config *app.Config, logger *zap.Logger) {
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
		if err := generateRollup(ctx, container, "1m", now, logger); err != nil {
			logger.Error("Failed to generate 1m rollup", zap.Error(err))
		}

		if now.Minute()%5 == 0 {
			if err := generateRollup(ctx, container, "5m", now, logger); err != nil {
				logger.Error("Failed to generate 5m rollup", zap.Error(err))
			}
		}

		if now.Minute() == 0 {
			if err := generateRollup(ctx, container, "1h", now, logger); err != nil {
				logger.Error("Failed to generate 1h rollup", zap.Error(err))
			}
		}
	}
}

// generateRollup generates a single rollup for the given interval
func generateRollup(ctx context.Context, container *storage.Container, interval string, now time.Time, logger *zap.Logger) error {
	// TODO: Implement rollup generation
	logger.Debug("Generating rollup", zap.String("interval", interval))
	return nil
}

// startCleanupTask periodically cleans up old data
func startCleanupTask(container *storage.Container, config *app.Config, logger *zap.Logger) {
	logger.Info("Starting cleanup task")
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		retention := 24 * time.Hour // TODO: Parse from config

		if err := container.Telemetry.CleanupOldData(ctx, retention); err != nil {
			logger.Error("Failed to cleanup old telemetry data", zap.Error(err))
		} else {
			logger.Debug("Cleaned up old telemetry data", zap.Duration("retention", retention))
		}
	}
}
