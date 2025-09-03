package bokio

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		baseURL  string
		readOnly string
		want     *Config
		wantErr  bool
	}{
		{
			name:     "valid config with all values",
			token:    "test-token-123",
			baseURL:  "https://test.api.bokio.se",
			readOnly: "true",
			want: &Config{
				IntegrationToken: "test-token-123",
				BaseURL:          "https://test.api.bokio.se",
				ReadOnly:         true,
			},
		},
		{
			name:     "valid config with defaults",
			token:    "test-token-456",
			baseURL:  "",
			readOnly: "false",
			want: &Config{
				IntegrationToken: "test-token-456",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         false,
			},
		},
		{
			name:     "empty token",
			token:    "",
			baseURL:  "https://api.bokio.se",
			readOnly: "false",
			want: &Config{
				IntegrationToken: "",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			os.Setenv("BOKIO_INTEGRATION_TOKEN", tt.token)
			if tt.baseURL != "" {
				os.Setenv("BOKIO_BASE_URL", tt.baseURL)
			} else {
				os.Unsetenv("BOKIO_BASE_URL")
			}
			os.Setenv("BOKIO_READ_ONLY", tt.readOnly)

			// Clean up environment variables after test
			defer func() {
				os.Unsetenv("BOKIO_INTEGRATION_TOKEN")
				os.Unsetenv("BOKIO_BASE_URL")
				os.Unsetenv("BOKIO_READ_ONLY")
			}()

			got := LoadConfigFromEnv()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewAuthClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: &Config{
				IntegrationToken: "test-token-123",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         false,
			},
			wantErr: false,
		},
		{
			name: "empty token",
			config: &Config{
				IntegrationToken: "",
				BaseURL:          "https://api.bokio.se",
				ReadOnly:         false,
			},
			wantErr: true,
			errMsg:  "BOKIO_INTEGRATION_TOKEN is required",
		},
		{
			name: "empty base URL gets default",
			config: &Config{
				IntegrationToken: "test-token-456",
				BaseURL:          "",
				ReadOnly:         true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewAuthClient(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, client)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, client)

			// Verify client properties
			assert.Equal(t, tt.config.IntegrationToken, client.GetToken())
			assert.Equal(t, tt.config.ReadOnly, client.IsReadOnly())
			assert.True(t, client.IsAuthenticated())

			// Verify default base URL is set when empty
			expectedBaseURL := tt.config.BaseURL
			if expectedBaseURL == "" {
				expectedBaseURL = "https://api.bokio.se"
			}
			assert.Equal(t, expectedBaseURL, client.GetBaseURL())

			// Verify generated clients are created
			assert.NotNil(t, client.CompanyClient)
			assert.NotNil(t, client.GeneralClient)

			// Verify config can be retrieved
			config := client.GetConfig()
			assert.NotNil(t, config)
			assert.Equal(t, tt.config.IntegrationToken, config.IntegrationToken)
			assert.Equal(t, expectedBaseURL, config.BaseURL)
			assert.Equal(t, tt.config.ReadOnly, config.ReadOnly)
		})
	}
}

func TestAuthenticatedHTTPClient(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		expectHeader  bool
		expectedValue string
	}{
		{
			name:          "valid token",
			token:         "test-token-123",
			expectHeader:  true,
			expectedValue: "Bearer test-token-123",
		},
		{
			name:          "empty token",
			token:         "",
			expectHeader:  true,
			expectedValue: "Bearer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server to capture the request
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectHeader {
					authHeader := r.Header.Get("Authorization")
					assert.Equal(t, tt.expectedValue, authHeader)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			// Create authenticated client
			client := &authenticatedHTTPClient{token: tt.token}

			// Create request
			req, err := http.NewRequest("GET", server.URL, nil)
			require.NoError(t, err)

			// Make request
			resp, err := client.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		setEnv       bool
		expected     string
	}{
		{
			name:         "environment variable set",
			key:          "TEST_ENV_VAR",
			defaultValue: "default",
			envValue:     "env-value",
			setEnv:       true,
			expected:     "env-value",
		},
		{
			name:         "environment variable not set",
			key:          "TEST_ENV_VAR_UNSET",
			defaultValue: "default",
			setEnv:       false,
			expected:     "default",
		},
		{
			name:         "environment variable set to empty string",
			key:          "TEST_ENV_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			setEnv:       true,
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			result := getEnvWithDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuthClientMethods(t *testing.T) {
	config := &Config{
		IntegrationToken: "test-token-789",
		BaseURL:          "https://test.bokio.se",
		ReadOnly:         true,
	}

	client, err := NewAuthClient(config)
	require.NoError(t, err)

	// Test GetToken
	assert.Equal(t, "test-token-789", client.GetToken())

	// Test GetBaseURL
	assert.Equal(t, "https://test.bokio.se", client.GetBaseURL())

	// Test IsAuthenticated
	assert.True(t, client.IsAuthenticated())

	// Test IsReadOnly
	assert.True(t, client.IsReadOnly())

	// Test GetConfig
	retrievedConfig := client.GetConfig()
	assert.NotNil(t, retrievedConfig)
	assert.Equal(t, config.IntegrationToken, retrievedConfig.IntegrationToken)
	assert.Equal(t, config.BaseURL, retrievedConfig.BaseURL)
	assert.Equal(t, config.ReadOnly, retrievedConfig.ReadOnly)
}

func TestAuthClientUnauthenticated(t *testing.T) {
	// Create a client with empty token to test unauthenticated state
	// This should fail in NewAuthClient, but let's test the method directly
	client := &AuthClient{
		token: "",
	}

	assert.False(t, client.IsAuthenticated())
	assert.Equal(t, "", client.GetToken())
}

// Test configuration validation edge cases
func TestConfigValidation(t *testing.T) {
	// Test various edge cases for configuration
	tests := []struct {
		name   string
		config *Config
		valid  bool
	}{
		{
			name: "whitespace token",
			config: &Config{
				IntegrationToken: "   ",
				BaseURL:          "https://api.bokio.se",
			},
			valid: true, // Current implementation doesn't trim, so whitespace is valid
		},
		{
			name: "very long token",
			config: &Config{
				IntegrationToken: string(make([]byte, 1000)), // Very long token
				BaseURL:          "https://api.bokio.se",
			},
			valid: true, // Should be valid even if very long
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAuthClient(tt.config)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
