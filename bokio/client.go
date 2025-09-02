// Package bokio provides a client for the Bokio API with OAuth2 authentication support.
package bokio

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"golang.org/x/oauth2"
)

// Client represents the Bokio API client with OAuth2 support
type Client struct {
	// HTTP client configuration
	httpClient *resty.Client
	baseURL    string

	// OAuth2 configuration
	oauth2Config *oauth2.Config
	clientID     string
	clientSecret string

	// Token management
	tokenMutex   sync.RWMutex
	accessToken  string
	refreshToken string
	tokenExpiry  time.Time
	tenantID     string
	tenantType   string

	// Rate limiting
	rateLimiter chan struct{}

	// Logging
	logger Logger
	
	// Security
	readOnly bool // When true, prevents all write operations
}

// Config holds the configuration for the Bokio API client
type Config struct {
	// OAuth2 credentials
	ClientID     string
	ClientSecret string

	// API configuration
	BaseURL     string
	RedirectURI string
	Scopes      []string

	// Client configuration
	Timeout     time.Duration
	MaxRetries  int
	RateLimit   int // requests per second
	UserAgent   string

	// Logging
	Logger Logger
	
	// Security
	ReadOnly bool // When true, prevents all write operations
}

// Logger interface for customizable logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// DefaultLogger provides a simple logger implementation
type DefaultLogger struct{}

func (l *DefaultLogger) Debug(msg string, fields ...interface{}) {
	log.Printf("[DEBUG] %s %v", msg, fields)
}

func (l *DefaultLogger) Info(msg string, fields ...interface{}) {
	log.Printf("[INFO] %s %v", msg, fields)
}

func (l *DefaultLogger) Warn(msg string, fields ...interface{}) {
	log.Printf("[WARN] %s %v", msg, fields)
}

func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	log.Printf("[ERROR] %s %v", msg, fields)
}

// TokenResponse represents the OAuth2 token response from Bokio API
type TokenResponse struct {
	TenantID     string `json:"tenant_id"`
	TenantType   string `json:"tenant_type"`
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// APIError represents an error response from the Bokio API
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("Bokio API error %d: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("Bokio API error %d: %s", e.Code, e.Message)
}

// DefaultConfig returns a default configuration for the Bokio client
func DefaultConfig() *Config {
	return &Config{
		BaseURL:    "https://api.bokio.se",
		Scopes:     []string{"accounting", "invoices"},
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RateLimit:  10, // 10 requests per second
		UserAgent:  "Bokio-MCP-Client/1.0",
		Logger:     &DefaultLogger{},
	}
}

// NewClient creates a new Bokio API client with the given configuration
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.ClientID == "" || config.ClientSecret == "" {
		return nil, fmt.Errorf("client ID and client secret are required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.bokio.se"
	}

	if config.Logger == nil {
		config.Logger = &DefaultLogger{}
	}

	// Initialize OAuth2 configuration
	oauth2Config := &oauth2.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  config.BaseURL + "/authorize",
			TokenURL: config.BaseURL + "/token",
		},
		RedirectURL: config.RedirectURI,
		Scopes:      config.Scopes,
	}

	// Initialize HTTP client
	httpClient := resty.New()
	httpClient.SetBaseURL(config.BaseURL)
	httpClient.SetTimeout(config.Timeout)
	httpClient.SetRetryCount(config.MaxRetries)
	httpClient.SetHeader("User-Agent", config.UserAgent)

	// Add retry condition
	httpClient.AddRetryCondition(func(r *resty.Response, err error) bool {
		return r.StatusCode() >= 500 || r.StatusCode() == 429
	})

	// Add debug logging if enabled
	httpClient.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		config.Logger.Debug("API Request", "method", req.Method, "url", req.URL, "headers", req.Header)
		return nil
	})

	httpClient.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		config.Logger.Debug("API Response", "status", resp.StatusCode(), "time", resp.Time())
		return nil
	})

	// Initialize rate limiter
	var rateLimiter chan struct{}
	if config.RateLimit > 0 {
		rateLimiter = make(chan struct{}, config.RateLimit)
		// Fill the rate limiter buffer
		for i := 0; i < config.RateLimit; i++ {
			rateLimiter <- struct{}{}
		}
		// Start rate limiter goroutine
		go func() {
			ticker := time.NewTicker(time.Second / time.Duration(config.RateLimit))
			defer ticker.Stop()
			for range ticker.C {
				select {
				case rateLimiter <- struct{}{}:
				default:
				}
			}
		}()
	}

	client := &Client{
		httpClient:   httpClient,
		baseURL:      config.BaseURL,
		oauth2Config: oauth2Config,
		clientID:     config.ClientID,
		clientSecret: config.ClientSecret,
		rateLimiter:  rateLimiter,
		logger:       config.Logger,
		readOnly:     config.ReadOnly,
	}

	return client, nil
}

// IsReadOnly returns true if the client is configured in read-only mode
func (c *Client) IsReadOnly() bool {
	return c.readOnly
}

// validateWriteOperation checks if write operations are allowed
func (c *Client) validateWriteOperation(operation string) error {
	if c.readOnly {
		return fmt.Errorf("operation '%s' not allowed in read-only mode. Set BOKIO_READ_ONLY=false to enable write operations", operation)
	}
	return nil
}

// GetAuthorizationURL returns the URL for OAuth2 authorization
func (c *Client) GetAuthorizationURL(state string) string {
	return c.oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCodeForToken exchanges an authorization code for access and refresh tokens
func (c *Client) ExchangeCodeForToken(ctx context.Context, code string) error {
	c.logger.Info("Exchanging authorization code for tokens")

	// Prepare the request body
	data := url.Values{
		"grant_type":   {"authorization_code"},
		"code":         {code},
		"redirect_uri": {c.oauth2Config.RedirectURL},
	}

	// Create basic auth header
	authHeader := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))

	resp, err := c.httpClient.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Authorization", "Basic "+authHeader).
		SetBody(data.Encode()).
		Post("/token")

	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return c.handleAPIError(resp)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	// Store tokens
	c.tokenMutex.Lock()
	c.accessToken = tokenResp.AccessToken
	c.refreshToken = tokenResp.RefreshToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	c.tenantID = tokenResp.TenantID
	c.tenantType = tokenResp.TenantType
	c.tokenMutex.Unlock()

	c.logger.Info("Successfully obtained access token", "tenant_id", c.tenantID, "expires_in", tokenResp.ExpiresIn)
	return nil
}

// RefreshAccessToken refreshes the access token using the refresh token
func (c *Client) RefreshAccessToken(ctx context.Context) error {
	c.tokenMutex.RLock()
	refreshToken := c.refreshToken
	c.tokenMutex.RUnlock()

	if refreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	c.logger.Info("Refreshing access token")

	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))

	resp, err := c.httpClient.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Authorization", "Basic "+authHeader).
		SetBody(data.Encode()).
		Post("/token")

	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return c.handleAPIError(resp)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	// Update stored tokens
	c.tokenMutex.Lock()
	c.accessToken = tokenResp.AccessToken
	if tokenResp.RefreshToken != "" {
		c.refreshToken = tokenResp.RefreshToken
	}
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	c.tokenMutex.Unlock()

	c.logger.Info("Successfully refreshed access token")
	return nil
}

// AuthenticateClientCredentials authenticates using client credentials for General API
func (c *Client) AuthenticateClientCredentials(ctx context.Context) error {
	c.logger.Info("Authenticating with client credentials")

	data := url.Values{
		"grant_type": {"client_credentials"},
	}

	authHeader := base64.StdEncoding.EncodeToString([]byte(c.clientID + ":" + c.clientSecret))

	resp, err := c.httpClient.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Authorization", "Basic "+authHeader).
		SetBody(data.Encode()).
		Post("/token")

	if err != nil {
		return fmt.Errorf("failed to authenticate with client credentials: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return c.handleAPIError(resp)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(resp.Body(), &tokenResp); err != nil {
		return fmt.Errorf("failed to parse token response: %w", err)
	}

	// Store access token (no refresh token for client credentials)
	c.tokenMutex.Lock()
	c.accessToken = tokenResp.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
	c.tenantID = tokenResp.TenantID
	c.tenantType = tokenResp.TenantType
	c.tokenMutex.Unlock()

	c.logger.Info("Successfully authenticated with client credentials")
	return nil
}

// ensureValidToken ensures we have a valid access token, refreshing if necessary
func (c *Client) ensureValidToken(ctx context.Context) error {
	c.tokenMutex.RLock()
	hasToken := c.accessToken != ""
	isExpired := time.Now().Add(5 * time.Minute).After(c.tokenExpiry) // Refresh 5 minutes early
	hasRefreshToken := c.refreshToken != ""
	c.tokenMutex.RUnlock()

	if !hasToken {
		return fmt.Errorf("no access token available, please authenticate first")
	}

	if isExpired && hasRefreshToken {
		return c.RefreshAccessToken(ctx)
	}

	return nil
}

// makeRequest performs a rate-limited HTTP request with proper authentication
func (c *Client) makeRequest(ctx context.Context, method, path string, body interface{}) (*resty.Response, error) {
	// Check read-only mode for write operations
	if method != "GET" && method != "HEAD" && method != "OPTIONS" {
		if err := c.validateWriteOperation(fmt.Sprintf("%s %s", method, path)); err != nil {
			return nil, err
		}
	}
	
	// Rate limiting
	if c.rateLimiter != nil {
		select {
		case <-c.rateLimiter:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// Ensure we have a valid token
	if err := c.ensureValidToken(ctx); err != nil {
		return nil, err
	}

	// Get current access token
	c.tokenMutex.RLock()
	accessToken := c.accessToken
	c.tokenMutex.RUnlock()

	req := c.httpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+accessToken).
		SetHeader("Content-Type", "application/json")

	if body != nil {
		req.SetBody(body)
	}

	var resp *resty.Response
	var err error

	switch strings.ToUpper(method) {
	case "GET":
		resp, err = req.Get(path)
	case "POST":
		resp, err = req.Post(path)
	case "PUT":
		resp, err = req.Put(path)
	case "DELETE":
		resp, err = req.Delete(path)
	case "PATCH":
		resp, err = req.Patch(path)
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	// Handle API errors
	if resp.StatusCode() >= 400 {
		return resp, c.handleAPIError(resp)
	}

	return resp, nil
}

// GET performs a GET request to the specified path
func (c *Client) GET(ctx context.Context, path string) (*resty.Response, error) {
	return c.makeRequest(ctx, "GET", path, nil)
}

// POST performs a POST request to the specified path with the given body
func (c *Client) POST(ctx context.Context, path string, body interface{}) (*resty.Response, error) {
	return c.makeRequest(ctx, "POST", path, body)
}

// PUT performs a PUT request to the specified path with the given body
func (c *Client) PUT(ctx context.Context, path string, body interface{}) (*resty.Response, error) {
	return c.makeRequest(ctx, "PUT", path, body)
}

// DELETE performs a DELETE request to the specified path
func (c *Client) DELETE(ctx context.Context, path string) (*resty.Response, error) {
	return c.makeRequest(ctx, "DELETE", path, nil)
}

// PATCH performs a PATCH request to the specified path with the given body
func (c *Client) PATCH(ctx context.Context, path string, body interface{}) (*resty.Response, error) {
	return c.makeRequest(ctx, "PATCH", path, body)
}

// handleAPIError processes API error responses and returns a structured error
func (c *Client) handleAPIError(resp *resty.Response) error {
	var apiError APIError
	
	// Try to parse the error response
	if err := json.Unmarshal(resp.Body(), &apiError); err != nil {
		// If we can't parse the error, create a generic one
		apiError = APIError{
			Code:    resp.StatusCode(),
			Message: "API request failed",
			Details: string(resp.Body()),
		}
	}

	// If no code was set, use the HTTP status code
	if apiError.Code == 0 {
		apiError.Code = resp.StatusCode()
	}

	c.logger.Error("API error", "status", resp.StatusCode(), "error", apiError.Message)
	return &apiError
}

// GetTenantInfo returns the current tenant information
func (c *Client) GetTenantInfo() (tenantID, tenantType string) {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.tenantID, c.tenantType
}

// IsAuthenticated returns whether the client has a valid access token
func (c *Client) IsAuthenticated() bool {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.accessToken != "" && time.Now().Before(c.tokenExpiry)
}

// SetTokens manually sets the access and refresh tokens (useful for token persistence)
func (c *Client) SetTokens(accessToken, refreshToken string, expiresAt time.Time) {
	c.tokenMutex.Lock()
	defer c.tokenMutex.Unlock()
	c.accessToken = accessToken
	c.refreshToken = refreshToken
	c.tokenExpiry = expiresAt
}

// GetTokens returns the current tokens (useful for token persistence)
func (c *Client) GetTokens() (accessToken, refreshToken string, expiresAt time.Time) {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.accessToken, c.refreshToken, c.tokenExpiry
}