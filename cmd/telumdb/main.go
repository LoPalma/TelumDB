package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/telumdb/telumdb/internal/config"
	"github.com/telumdb/telumdb/internal/server"
	"github.com/telumdb/telumdb/pkg/storage"
	"go.uber.org/zap"
)

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var (
		configFile  = flag.String("config", "", "Path to configuration file")
		showHelp    = flag.Bool("help", false, "Show help message")
		showVersion = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showHelp {
		printHelp()
		return
	}

	if *showVersion {
		printVersion()
		return
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	logger.Info("Starting TelumDB",
		zap.String("version", version),
		zap.String("commit", commit),
		zap.String("build_date", date),
	)

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	// Initialize storage engine
	storageEngine, err := storage.New(cfg.Storage)
	if err != nil {
		logger.Fatal("Failed to initialize storage engine", zap.Error(err))
	}

	// Create server
	srv, err := server.New(cfg, storageEngine, logger)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := srv.Start(ctx); err != nil {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during server shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}

func printHelp() {
	fmt.Printf(`TelumDB - The World's First Hybrid General-Purpose + AI Tensor Database

Usage:
  telumdb [options]

Options:
  -config string     Path to configuration file (default: config.yaml)
  -help              Show this help message
  -version           Show version information

Environment Variables:
  TELUMDB_CONFIG_FILE    Path to configuration file
  TELUMDB_DATA_DIR       Data directory path
  TELUMDB_LOG_LEVEL      Log level (debug, info, warn, error)

Examples:
  telumdb                                    # Start with default config
  telumdb -config /etc/telumdb/config.yaml   # Start with custom config
  telumdb -version                           # Show version

For more information, visit: https://github.com/telumdb/telumdb
`)
}

func printVersion() {
	fmt.Printf(`TelumDB %s
Commit: %s
Built: %s
Go Version: %s
OS/Arch: %s/%s

Copyright 2024 TelumDB Contributors
License: Apache License 2.0
`,
		version,
		commit,
		date,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}
