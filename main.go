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

	// Initialize Bokio API client
	bokioClient, err := bokio.NewClient(&config.BokioConfig)
	if err != nil {
		return fmt.Errorf("failed to create Bokio client: %w", err)
	}

	// Create MCP server
	server := mcp.NewServer(serverName, serverVersion, nil)

	// Register tools with the server
	if err := tools.RegisterAuthTools(server, bokioClient); err != nil {
		return fmt.Errorf("failed to register auth tools: %w", err)
	}

	if err := tools.RegisterInvoiceTools(server, bokioClient); err != nil {
		return fmt.Errorf("failed to register invoice tools: %w", err)
	}

	if err := tools.RegisterCustomerTools(server, bokioClient); err != nil {
		return fmt.Errorf("failed to register customer tools: %w", err)
	}

	if err := tools.RegisterJournalTools(server, bokioClient); err != nil {
		return fmt.Errorf("failed to register journal tools: %w", err)
	}

	if err := tools.RegisterUploadTools(server, bokioClient); err != nil {
		return fmt.Errorf("failed to register upload tools: %w", err)
	}

	slog.Info("Starting Bokio MCP server",
		"name", serverName,
		"version", serverVersion,
		"bokio_base_url", config.BokioConfig.BaseURL,
		"read_only_mode", config.ReadOnly)

	// Create and start the MCP server with stdio transport
	transport := mcp.NewStdioTransport()
	return server.Run(ctx, transport)
}

// Config holds all application configuration
type Config struct {
	BokioConfig bokio.Config
	ReadOnly    bool
}

// loadConfig loads configuration from environment variables
func loadConfig() (*Config, error) {
	// Parse read-only mode
	readOnly := os.Getenv("BOKIO_READ_ONLY") == "true"

	bokioConfig := bokio.Config{
		BaseURL:      getEnvWithDefault("BOKIO_BASE_URL", "https://api.bokio.se"),
		ClientID:     os.Getenv("BOKIO_CLIENT_ID"),
		ClientSecret: os.Getenv("BOKIO_CLIENT_SECRET"),
		RedirectURI:  getEnvWithDefault("BOKIO_REDIRECT_URL", "http://localhost:8080/callback"),
		ReadOnly:     readOnly,
	}

	// Validate required configuration
	if bokioConfig.ClientID == "" {
		return nil, fmt.Errorf("BOKIO_CLIENT_ID environment variable is required")
	}
	if bokioConfig.ClientSecret == "" {
		return nil, fmt.Errorf("BOKIO_CLIENT_SECRET environment variable is required")
	}

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
