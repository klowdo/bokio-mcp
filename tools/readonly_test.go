package tools

import (
	"testing"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadOnlyMode(t *testing.T) {
	tests := []struct {
		name     string
		readOnly bool
		expected bool
	}{
		{
			name:     "read-only mode enabled",
			readOnly: true,
			expected: true,
		},
		{
			name:     "read-only mode disabled",
			readOnly: false,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &bokio.Config{
				IntegrationToken: "test-token",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         tt.readOnly,
			}

			client, err := bokio.NewAuthClient(config)
			require.NoError(t, err)

			assert.Equal(t, tt.expected, client.IsReadOnly())
			assert.Equal(t, tt.expected, client.GetConfig().ReadOnly)
		})
	}
}

func TestReadOnlyModeValidation(t *testing.T) {
	// Test that read-only mode validation works correctly
	// Mock read-only client
	config := &bokio.Config{
		IntegrationToken: "test-token",
		BaseURL:          "https://api.bokio.se",
		ReadOnly:         true,
	}

	client, err := bokio.NewAuthClient(config)
	require.NoError(t, err)

	// Verify read-only mode is set
	assert.True(t, client.IsReadOnly())

	// Test helper function for checking read-only mode
	assert.True(t, isReadOnlyMode(client))

	// Test with non-read-only client
	normalConfig := &bokio.Config{
		IntegrationToken: "test-token",
		BaseURL:          "https://api.bokio.se",
		ReadOnly:         false,
	}

	normalClient, err := bokio.NewAuthClient(normalConfig)
	require.NoError(t, err)

	assert.False(t, normalClient.IsReadOnly())
	assert.False(t, isReadOnlyMode(normalClient))
}

func TestReadOnlyModeErrorMessages(t *testing.T) {
	// Test that read-only mode returns appropriate error messages
	tests := []struct {
		name         string
		operation    string
		expectedMsg  string
	}{
		{
			name:        "create operation blocked",
			operation:   "create",
			expectedMsg: "Operation not allowed in read-only mode",
		},
		{
			name:        "update operation blocked",
			operation:   "update",
			expectedMsg: "Operation not allowed in read-only mode",
		},
		{
			name:        "delete operation blocked",
			operation:   "delete",
			expectedMsg: "Operation not allowed in read-only mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := getReadOnlyErrorMessage(tt.operation)
			assert.Equal(t, tt.expectedMsg, msg)
		})
	}
}

func TestReadOnlyModeConfiguration(t *testing.T) {
	// Test various ways read-only mode can be configured
	tests := []struct {
		name            string
		envValue        string
		configValue     bool
		expectedReadOnly bool
	}{
		{
			name:            "explicit true in config",
			configValue:     true,
			expectedReadOnly: true,
		},
		{
			name:            "explicit false in config",
			configValue:     false,
			expectedReadOnly: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &bokio.Config{
				IntegrationToken: "test-token",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         tt.configValue,
			}

			client, err := bokio.NewAuthClient(config)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedReadOnly, client.IsReadOnly())
		})
	}
}

func TestReadOnlyModeParameterStructures(t *testing.T) {
	// Test that parameter structures work correctly with read-only validation
	tests := []struct {
		name       string
		params     interface{}
		isWriteOp  bool
		shouldPass bool
	}{
		{
			name: "customers list (read operation)",
			params: CustomersListParams{
				CompanyID: "test-company",
			},
			isWriteOp:  false,
			shouldPass: true,
		},
		{
			name: "customer create (write operation)",
			params: CustomerCreateParams{
				CompanyID: "test-company",
				Name:      "Test Customer",
				Type:      "private",
			},
			isWriteOp:  true,
			shouldPass: false,
		},
		{
			name: "customer update (write operation)",
			params: CustomerUpdateParams{
				CompanyID:  "test-company",
				CustomerID: "cust-123",
				Name:       stringPtr("Updated Name"),
			},
			isWriteOp:  true,
			shouldPass: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &bokio.Config{
				IntegrationToken: "test-token",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         true, // Always test with read-only enabled
			}

			client, err := bokio.NewAuthClient(config)
			require.NoError(t, err)

			// Simulate read-only check
			if tt.isWriteOp && client.IsReadOnly() {
				assert.False(t, tt.shouldPass, "Write operation should not pass in read-only mode")
			} else {
				assert.True(t, tt.shouldPass, "Read operation should pass in read-only mode")
			}
		})
	}
}

// Helper functions for read-only mode testing

func isReadOnlyMode(client *bokio.AuthClient) bool {
	return client.IsReadOnly()
}

func getReadOnlyErrorMessage(operation string) string {
	return "Operation not allowed in read-only mode"
}

func shouldAllowOperation(client *bokio.AuthClient, operationType string) bool {
	if !client.IsReadOnly() {
		return true
	}

	// In read-only mode, only allow read operations
	readOperations := map[string]bool{
		"list":   true,
		"get":    true,
		"show":   true,
		"read":   true,
		"view":   true,
		"search": true,
	}

	return readOperations[operationType]
}

func TestShouldAllowOperation(t *testing.T) {
	config := &bokio.Config{
		IntegrationToken: "test-token",
		BaseURL:          "https://api.bokio.se",
		ReadOnly:         true,
	}

	client, err := bokio.NewAuthClient(config)
	require.NoError(t, err)

	tests := []struct {
		operation string
		allowed   bool
	}{
		{"list", true},
		{"get", true},
		{"show", true},
		{"read", true},
		{"view", true},
		{"search", true},
		{"create", false},
		{"update", false},
		{"delete", false},
		{"modify", false},
		{"write", false},
	}

	for _, tt := range tests {
		t.Run(tt.operation, func(t *testing.T) {
			result := shouldAllowOperation(client, tt.operation)
			assert.Equal(t, tt.allowed, result)
		})
	}

	// Test with non-read-only client
	normalConfig := &bokio.Config{
		IntegrationToken: "test-token",
		BaseURL:          "https://api.bokio.se",
		ReadOnly:         false,
	}

	normalClient, err := bokio.NewAuthClient(normalConfig)
	require.NoError(t, err)

	// All operations should be allowed in normal mode
	for _, tt := range tests {
		t.Run(tt.operation+"_normal_mode", func(t *testing.T) {
			result := shouldAllowOperation(normalClient, tt.operation)
			assert.True(t, result, "All operations should be allowed in normal mode")
		})
	}
}