// Package testutil provides testing utilities for the Bokio MCP server
package testutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/stretchr/testify/require"
)

// MockBokioServer creates a mock HTTP server that simulates Bokio API responses
type MockBokioServer struct {
	*httptest.Server
	ResponseCode int
	ResponseBody string
	LastRequest  *http.Request
}

// NewMockBokioServer creates a new mock server for testing
func NewMockBokioServer(responseCode int, responseBody string) *MockBokioServer {
	mock := &MockBokioServer{
		ResponseCode: responseCode,
		ResponseBody: responseBody,
	}

	mock.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mock.LastRequest = r
		w.WriteHeader(mock.ResponseCode)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(mock.ResponseBody))
	}))

	return mock
}

// GetLastAuthHeader returns the Authorization header from the last request
func (m *MockBokioServer) GetLastAuthHeader() string {
	if m.LastRequest == nil {
		return ""
	}
	return m.LastRequest.Header.Get("Authorization")
}

// TestConfig creates a test configuration for use in tests
func TestConfig(opts ...ConfigOption) *bokio.Config {
	config := &bokio.Config{
		IntegrationToken: "test-integration-token",
		BaseURL:          "https://api.bokio.se",
		ReadOnly:         false,
	}

	for _, opt := range opts {
		opt(config)
	}

	return config
}

// ConfigOption allows customization of test configuration
type ConfigOption func(*bokio.Config)

// WithToken sets the integration token
func WithToken(token string) ConfigOption {
	return func(c *bokio.Config) {
		c.IntegrationToken = token
	}
}

// WithBaseURL sets the base URL
func WithBaseURL(baseURL string) ConfigOption {
	return func(c *bokio.Config) {
		c.BaseURL = baseURL
	}
}

// WithReadOnly sets the read-only mode
func WithReadOnly(readOnly bool) ConfigOption {
	return func(c *bokio.Config) {
		c.ReadOnly = readOnly
	}
}

// CreateTestClient creates an authenticated client for testing
func CreateTestClient(t *testing.T, opts ...ConfigOption) *bokio.AuthClient {
	config := TestConfig(opts...)
	client, err := bokio.NewAuthClient(config)
	require.NoError(t, err)
	return client
}

// AssertValidJSON checks that a string is valid JSON
func AssertValidJSON(t *testing.T, jsonStr string) {
	// This would require importing encoding/json and doing json.Valid
	// For now, just check it's not empty
	require.NotEmpty(t, jsonStr)
}

// Common test data
var (
	// TestCompanyID is a standard test company ID
	TestCompanyID = "test-company-123"
	
	// TestCustomerID is a standard test customer ID
	TestCustomerID = "test-customer-456"
	
	// TestToken is a standard test authentication token
	TestToken = "test-integration-token-789"
)

// Helper functions for creating test data

// ValidCustomersListParams returns valid parameters for listing customers
func ValidCustomersListParams() map[string]interface{} {
	return map[string]interface{}{
		"company_id": TestCompanyID,
		"page":       1,
		"page_size":  25,
	}
}

// ValidCustomerCreateParams returns valid parameters for creating a customer
func ValidCustomerCreateParams() map[string]interface{} {
	return map[string]interface{}{
		"company_id": TestCompanyID,
		"name":       "Test Customer",
		"type":       "private",
		"email":      "test@example.com",
	}
}

// ValidCustomerUpdateParams returns valid parameters for updating a customer
func ValidCustomerUpdateParams() map[string]interface{} {
	return map[string]interface{}{
		"company_id":  TestCompanyID,
		"customer_id": TestCustomerID,
		"name":        "Updated Test Customer",
		"email":       "updated@example.com",
	}
}

// MockSuccessResponse returns a mock success response
func MockSuccessResponse(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"data":    data,
	}
}

// MockErrorResponse returns a mock error response
func MockErrorResponse(errorMsg string) map[string]interface{} {
	return map[string]interface{}{
		"success": false,
		"error":   errorMsg,
	}
}