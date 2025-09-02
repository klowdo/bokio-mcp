// Package bokio provides authenticated access to Bokio API using generated clients
package bokio

import (
	"fmt"
	"net/http"
	"os"

	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/klowdo/bokio-mcp/bokio/generated/general"
)

// AuthClient wraps generated API clients with Bearer token authentication
type AuthClient struct {
	CompanyClient *company.Client
	GeneralClient *general.Client
	token         string
	baseURL       string
	readOnly      bool
}

// Config holds the simple configuration for the auth client
type Config struct {
	IntegrationToken string
	BaseURL          string
	ReadOnly         bool
}

// NewAuthClient creates a new authenticated client using generated clients
func NewAuthClient(config *Config) (*AuthClient, error) {
	if config.IntegrationToken == "" {
		return nil, fmt.Errorf("BOKIO_INTEGRATION_TOKEN is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.bokio.se"
	}

	// Create authenticated HTTP client
	httpClient := &authenticatedHTTPClient{token: config.IntegrationToken}

	// Create generated clients with authentication
	companyClient, err := company.NewClient(config.BaseURL, company.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create company client: %w", err)
	}

	generalClient, err := general.NewClient(config.BaseURL, general.WithHTTPClient(httpClient))
	if err != nil {
		return nil, fmt.Errorf("failed to create general client: %w", err)
	}

	return &AuthClient{
		CompanyClient: companyClient,
		GeneralClient: generalClient,
		token:         config.IntegrationToken,
		baseURL:       config.BaseURL,
		readOnly:      config.ReadOnly,
	}, nil
}

// LoadConfigFromEnv loads configuration from environment variables
func LoadConfigFromEnv() *Config {
	return &Config{
		IntegrationToken: os.Getenv("BOKIO_INTEGRATION_TOKEN"),
		BaseURL:          getEnvWithDefault("BOKIO_BASE_URL", "https://api.bokio.se"),
		ReadOnly:         os.Getenv("BOKIO_READ_ONLY") == "true",
	}
}

// authenticatedHTTPClient adds Bearer token authentication to all requests
type authenticatedHTTPClient struct {
	token string
}

// Do implements the HttpRequestDoer interface by adding Bearer token authentication
func (c *authenticatedHTTPClient) Do(req *http.Request) (*http.Response, error) {
	// Add Bearer token to all requests
	req.Header.Set("Authorization", "Bearer "+c.token)

	// Use default HTTP client
	return http.DefaultClient.Do(req)
}

// GetToken returns the current authentication token
func (ac *AuthClient) GetToken() string {
	return ac.token
}

// GetBaseURL returns the base URL for the API
func (ac *AuthClient) GetBaseURL() string {
	return ac.baseURL
}

// IsAuthenticated returns true if the client has an authentication token
func (ac *AuthClient) IsAuthenticated() bool {
	return ac.token != ""
}

// GetConfig returns the current configuration including read-only mode
func (ac *AuthClient) GetConfig() *Config {
	return &Config{
		IntegrationToken: ac.token,
		BaseURL:          ac.baseURL,
		ReadOnly:         ac.readOnly,
	}
}

// IsReadOnly returns true if the client is in read-only mode
func (ac *AuthClient) IsReadOnly() bool {
	return ac.readOnly
}

// getEnvWithDefault returns the value of an environment variable or a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
