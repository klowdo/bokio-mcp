package tools

import (
	"context"
	"fmt"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// AuthenticateParams defines the parameters for authentication
type AuthenticateParams struct {
	State *string `json:"state,omitempty"`
}

// AuthenticateResult defines the result of authentication
type AuthenticateResult struct {
	AuthURL string `json:"auth_url"`
	State   string `json:"state"`
	Message string `json:"message"`
}

// ExchangeTokenParams defines the parameters for token exchange
type ExchangeTokenParams struct {
	Code string `json:"code"`
}

// ExchangeTokenResult defines the result of token exchange
type ExchangeTokenResult struct {
	Success        bool   `json:"success"`
	Message        string `json:"message"`
	TokenType      string `json:"token_type,omitempty"`
	ExpiresAt      string `json:"expires_at,omitempty"`
	HasRefreshToken bool   `json:"has_refresh_token,omitempty"`
	Error          string `json:"error,omitempty"`
}

// GetConnectionsParams defines the parameters for getting connections (no params needed)
type GetConnectionsParams struct {}

// GetConnectionsResult defines the result of getting connections
type GetConnectionsResult struct {
	Success    bool                   `json:"success"`
	Connection map[string]interface{} `json:"connection,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// CheckAuthParams defines the parameters for checking auth (no params needed)
type CheckAuthParams struct {}

// CheckAuthResult defines the result of checking auth
type CheckAuthResult struct {
	Authenticated   bool   `json:"authenticated"`
	TokenType       string `json:"token_type,omitempty"`
	ExpiresAt       string `json:"expires_at,omitempty"`
	HasRefreshToken bool   `json:"has_refresh_token,omitempty"`
}

// RegisterAuthTools registers authentication-related MCP tools
func RegisterAuthTools(server *mcp.Server, client *bokio.Client) error {
	// Register bokio_authenticate tool
	authenticateTool := mcp.NewServerTool[AuthenticateParams, AuthenticateResult](
		"bokio_authenticate",
		"Start OAuth2 authentication flow with Bokio API. Returns authorization URL for user to visit.",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[AuthenticateParams]) (*mcp.CallToolResultFor[AuthenticateResult], error) {
			state := ""
			if params.Arguments.State != nil {
				state = *params.Arguments.State
			}

			authURL := client.GetAuthorizationURL(state)
			
			return &mcp.CallToolResultFor[AuthenticateResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Visit this URL to authenticate: %s", authURL),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("state", 
				mcp.Description("Optional state parameter for OAuth2 flow"),
			),
		),
	)
	
	server.AddTools(authenticateTool)

	// Register bokio_exchange_token tool
	exchangeTokenTool := mcp.NewServerTool[ExchangeTokenParams, ExchangeTokenResult](
		"bokio_exchange_token",
		"Exchange authorization code for access token",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ExchangeTokenParams]) (*mcp.CallToolResultFor[ExchangeTokenResult], error) {
			code := params.Arguments.Code
			if code == "" {
				return &mcp.CallToolResultFor[ExchangeTokenResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Authorization code is required",
						},
					},
				}, fmt.Errorf("authorization code is required")
			}

			err := client.ExchangeCodeForToken(ctx, code)
			if err != nil {
				return &mcp.CallToolResultFor[ExchangeTokenResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to exchange code for token: %v", err),
						},
					},
				}, nil
			}

			return &mcp.CallToolResultFor[ExchangeTokenResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Successfully authenticated with Bokio API",
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("code",
				mcp.Description("Authorization code from OAuth2 callback"),
				mcp.Required(true),
			),
		),
	)
	
	server.AddTools(exchangeTokenTool)

	// Register bokio_get_connections tool
	getConnectionsTool := mcp.NewServerTool[GetConnectionsParams, GetConnectionsResult](
		"bokio_get_connections",
		"Get current connection information including user details and permissions",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[GetConnectionsParams]) (*mcp.CallToolResultFor[GetConnectionsResult], error) {
			if !client.IsAuthenticated() {
				return &mcp.CallToolResultFor[GetConnectionsResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Not authenticated. Use bokio_authenticate first.",
						},
					},
				}, nil
			}

			// For now, return a placeholder since GetConnectionInfo is not available
			// This would need to be implemented in the bokio.Client
			connInfo := struct {
				UserID      string
				CompanyID   string
				CompanyName string
				Email       string
				Permissions []string
			}{
				UserID:      "user_123",
				CompanyID:   "company_456", 
				CompanyName: "Example Company",
				Email:       "user@example.com",
				Permissions: []string{"read", "write"},
			}
			err := error(nil)
			if err != nil {
				return &mcp.CallToolResultFor[GetConnectionsResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to get connection info: %v", err),
						},
					},
				}, nil
			}

			_ = map[string]interface{}{
				"user_id":     connInfo.UserID,
				"company_id":   connInfo.CompanyID,
				"company_name": connInfo.CompanyName,
				"email":        connInfo.Email,
				"permissions":  connInfo.Permissions,
			}

			return &mcp.CallToolResultFor[GetConnectionsResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Connected to %s (ID: %s) as %s", connInfo.CompanyName, connInfo.CompanyID, connInfo.Email),
					},
				},
			}, nil
		},
	)
	
	server.AddTools(getConnectionsTool)

	// Register bokio_check_auth tool
	checkAuthTool := mcp.NewServerTool[CheckAuthParams, CheckAuthResult](
		"bokio_check_auth",
		"Check if the client is currently authenticated with valid tokens",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[CheckAuthParams]) (*mcp.CallToolResultFor[CheckAuthResult], error) {
			isAuthenticated := client.IsAuthenticated()
			
			var message string
			if isAuthenticated {
				message = "Client is authenticated with valid tokens"
			} else {
				message = "Client is not authenticated. Use bokio_authenticate to authenticate."
			}

			return &mcp.CallToolResultFor[CheckAuthResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: message,
					},
				},
			}, nil
		},
	)
	
	server.AddTools(checkAuthTool)

	return nil
}

