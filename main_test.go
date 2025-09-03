package main

import (
	"os"
	"testing"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		envToken    string
		envBaseURL  string
		envReadOnly string
		expectError bool
		errorMsg    string
		expected    *bokio.Config
	}{
		{
			name:        "valid config with all env vars",
			envToken:    "test-integration-token-123",
			envBaseURL:  "https://test.api.bokio.se",
			envReadOnly: "true",
			expectError: false,
			expected: &bokio.Config{
				IntegrationToken: "test-integration-token-123",
				BaseURL:          "https://test.api.bokio.se",
				ReadOnly:         true,
			},
		},
		{
			name:        "valid config with minimal env vars",
			envToken:    "test-integration-token-456",
			expectError: false,
			expected: &bokio.Config{
				IntegrationToken: "test-integration-token-456",
				BaseURL:          "https://api.bokio.se", // default
				ReadOnly:         false,                  // default
			},
		},
		{
			name:        "missing integration token",
			envToken:    "",
			expectError: true,
			errorMsg:    "BOKIO_INTEGRATION_TOKEN is required",
		},
		{
			name:        "read-only mode variations",
			envToken:    "test-token",
			envReadOnly: "false",
			expectError: false,
			expected: &bokio.Config{
				IntegrationToken: "test-token",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			if tt.envToken != "" {
				os.Setenv("BOKIO_INTEGRATION_TOKEN", tt.envToken)
			} else {
				os.Unsetenv("BOKIO_INTEGRATION_TOKEN")
			}

			if tt.envBaseURL != "" {
				os.Setenv("BOKIO_BASE_URL", tt.envBaseURL)
			} else {
				os.Unsetenv("BOKIO_BASE_URL")
			}

			if tt.envReadOnly != "" {
				os.Setenv("BOKIO_READ_ONLY", tt.envReadOnly)
			} else {
				os.Unsetenv("BOKIO_READ_ONLY")
			}

			// Clean up after test
			defer func() {
				os.Unsetenv("BOKIO_INTEGRATION_TOKEN")
				os.Unsetenv("BOKIO_BASE_URL")
				os.Unsetenv("BOKIO_READ_ONLY")
			}()

			// Test loadConfig function
			config, err := loadConfig()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, config)
			} else {
				require.NoError(t, err)
				require.NotNil(t, config)
				assert.Equal(t, tt.expected.IntegrationToken, config.IntegrationToken)
				assert.Equal(t, tt.expected.BaseURL, config.BaseURL)
				assert.Equal(t, tt.expected.ReadOnly, config.ReadOnly)
			}
		})
	}
}

func TestLoadConfigFromEnvironment(t *testing.T) {
	// Test direct usage of bokio.LoadConfigFromEnv which is used by loadConfig
	tests := []struct {
		name     string
		setupEnv func()
		expected *bokio.Config
	}{
		{
			name: "default configuration",
			setupEnv: func() {
				os.Unsetenv("BOKIO_INTEGRATION_TOKEN")
				os.Unsetenv("BOKIO_BASE_URL")
				os.Unsetenv("BOKIO_READ_ONLY")
			},
			expected: &bokio.Config{
				IntegrationToken: "",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         false,
			},
		},
		{
			name: "custom base URL",
			setupEnv: func() {
				os.Setenv("BOKIO_INTEGRATION_TOKEN", "test-token")
				os.Setenv("BOKIO_BASE_URL", "https://staging.api.bokio.se")
				os.Setenv("BOKIO_READ_ONLY", "false")
			},
			expected: &bokio.Config{
				IntegrationToken: "test-token",
				BaseURL:          "https://staging.api.bokio.se",
				ReadOnly:         false,
			},
		},
		{
			name: "read-only mode enabled",
			setupEnv: func() {
				os.Setenv("BOKIO_INTEGRATION_TOKEN", "readonly-token")
				os.Setenv("BOKIO_READ_ONLY", "true")
			},
			expected: &bokio.Config{
				IntegrationToken: "readonly-token",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupEnv()
			defer func() {
				os.Unsetenv("BOKIO_INTEGRATION_TOKEN")
				os.Unsetenv("BOKIO_BASE_URL")
				os.Unsetenv("BOKIO_READ_ONLY")
			}()

			config := bokio.LoadConfigFromEnv()
			assert.Equal(t, tt.expected, config)
		})
	}
}

func TestConfigurationEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		baseURL  string
		readOnly string
		expected bool // for ReadOnly field
	}{
		{
			name:     "read-only with 'TRUE' (uppercase)",
			token:    "test-token",
			readOnly: "TRUE",
			expected: false, // Only exactly "true" should be true
		},
		{
			name:     "read-only with '1'",
			token:    "test-token",
			readOnly: "1",
			expected: false, // Only exactly "true" should be true
		},
		{
			name:     "read-only with 'yes'",
			token:    "test-token",
			readOnly: "yes",
			expected: false, // Only exactly "true" should be true
		},
		{
			name:     "read-only exactly 'true'",
			token:    "test-token",
			readOnly: "true",
			expected: true,
		},
		{
			name:     "empty read-only env var",
			token:    "test-token",
			readOnly: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("BOKIO_INTEGRATION_TOKEN", tt.token)
			if tt.baseURL != "" {
				os.Setenv("BOKIO_BASE_URL", tt.baseURL)
			}
			os.Setenv("BOKIO_READ_ONLY", tt.readOnly)

			defer func() {
				os.Unsetenv("BOKIO_INTEGRATION_TOKEN")
				os.Unsetenv("BOKIO_BASE_URL")
				os.Unsetenv("BOKIO_READ_ONLY")
			}()

			config, err := loadConfig()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, config.ReadOnly)
		})
	}
}

func TestConfigurationConstants(t *testing.T) {
	// Test that constants are defined correctly
	assert.Equal(t, "bokio-mcp", serverName)
	assert.Equal(t, "0.1.0", serverVersion)
	assert.NotEmpty(t, serverName)
	assert.NotEmpty(t, serverVersion)
}

func TestConfigValidationLogic(t *testing.T) {
	// Test the validation logic in loadConfig function
	tests := []struct {
		name           string
		configOverride *bokio.Config
		expectError    bool
		errorContains  string
	}{
		{
			name: "valid config should pass validation",
			configOverride: &bokio.Config{
				IntegrationToken: "valid-token-123",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         false,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.configOverride != nil {
				// Set environment to match the override
				os.Setenv("BOKIO_INTEGRATION_TOKEN", tt.configOverride.IntegrationToken)
				os.Setenv("BOKIO_BASE_URL", tt.configOverride.BaseURL)
				if tt.configOverride.ReadOnly {
					os.Setenv("BOKIO_READ_ONLY", "true")
				} else {
					os.Setenv("BOKIO_READ_ONLY", "false")
				}

				defer func() {
					os.Unsetenv("BOKIO_INTEGRATION_TOKEN")
					os.Unsetenv("BOKIO_BASE_URL")
					os.Unsetenv("BOKIO_READ_ONLY")
				}()
			}

			config, err := loadConfig()

			if tt.expectError {
				require.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, config)
			}
		})
	}
}

func TestEnvironmentVariablesPrecedence(t *testing.T) {
	// Test that environment variables have correct precedence and defaults

	// Test default values when no env vars are set
	t.Run("defaults", func(t *testing.T) {
		os.Unsetenv("BOKIO_INTEGRATION_TOKEN")
		os.Unsetenv("BOKIO_BASE_URL")
		os.Unsetenv("BOKIO_READ_ONLY")

		defer func() {
			os.Unsetenv("BOKIO_INTEGRATION_TOKEN")
			os.Unsetenv("BOKIO_BASE_URL")
			os.Unsetenv("BOKIO_READ_ONLY")
		}()

		config := bokio.LoadConfigFromEnv()
		assert.Equal(t, "", config.IntegrationToken)
		assert.Equal(t, "https://api.bokio.se", config.BaseURL) // Default
		assert.Equal(t, false, config.ReadOnly)                 // Default
	})

	// Test override of defaults
	t.Run("overrides", func(t *testing.T) {
		os.Setenv("BOKIO_INTEGRATION_TOKEN", "override-token")
		os.Setenv("BOKIO_BASE_URL", "https://override.api.bokio.se")
		os.Setenv("BOKIO_READ_ONLY", "true")

		defer func() {
			os.Unsetenv("BOKIO_INTEGRATION_TOKEN")
			os.Unsetenv("BOKIO_BASE_URL")
			os.Unsetenv("BOKIO_READ_ONLY")
		}()

		config := bokio.LoadConfigFromEnv()
		assert.Equal(t, "override-token", config.IntegrationToken)
		assert.Equal(t, "https://override.api.bokio.se", config.BaseURL)
		assert.Equal(t, true, config.ReadOnly)
	})
}
