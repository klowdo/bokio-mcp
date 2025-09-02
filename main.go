package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "bokio-mcp"
	serverVersion = "0.1.0"
)

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Info("Received shutdown signal, gracefully shutting down...")
		cancel()
	}()

	if err := run(ctx); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	// Load configuration from environment
	config, err := loadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize Bokio API client using only generated clients
	bokioClient, err := bokio.NewAuthClient(config.BokioConfig)
	if err != nil {
		return fmt.Errorf("failed to create Bokio auth client: %w", err)
	}

	// Create MCP server
	server := mcp.NewServer(serverName, serverVersion, nil)

	// Register tools with the server using ONLY generated API clients

	// Register pure generated journal tools (working demonstration)
	if err := tools.RegisterGeneratedJournalTools(server, bokioClient); err != nil {
		return fmt.Errorf("failed to register generated journal tools: %w", err)
	}

	// TODO: Migrate remaining tools to use generated clients
	// The old tools used manual types that don't exist in the actual API schema
	// They need to be rewritten to use the generated client methods and types

	slog.Info("Starting Bokio MCP server",
		"name", serverName,
		"version", serverVersion,
		"bokio_base_url", config.BokioConfig.BaseURL,
		"auth_method", "Integration Token",
		"authenticated", bokioClient.IsAuthenticated(),
		"read_only_mode", config.ReadOnly)

	// Create and start the MCP server with stdio transport
	transport := mcp.NewStdioTransport()
	return server.Run(ctx, transport)
}

// Config holds all application configuration
type Config struct {
	BokioConfig *bokio.Config
	ReadOnly    bool
}

// loadConfig loads configuration from environment variables
func loadConfig() (*Config, error) {
	// Load simple configuration from environment
	bokioConfig := bokio.LoadConfigFromEnv()

	if bokioConfig.IntegrationToken == "" {
		return nil, fmt.Errorf("BOKIO_INTEGRATION_TOKEN is required")
	}

	// Parse read-only mode
	readOnly := os.Getenv("BOKIO_READ_ONLY") == "true"

	return &Config{
		BokioConfig: bokioConfig,
		ReadOnly:    readOnly,
	}, nil
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
