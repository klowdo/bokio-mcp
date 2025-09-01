package tools

import (
	"context"
	"fmt"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAuthTools registers authentication-related MCP tools
func RegisterAuthTools(server *mcp.Server, client *bokio.Client) error {
	// Register bokio_authenticate tool
	if err := server.RegisterTool("bokio_authenticate", mcp.Tool{
		Name: "bokio_authenticate",
		Description: "Start OAuth2 authentication flow with Bokio API. Returns authorization URL for user to visit.",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"state": map[string]interface{}{
					"type": "string",
					"description": "Optional state parameter for OAuth2 flow",
				},
			},
		},
		Handler: createAuthenticateHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_authenticate tool: %w", err)
	}

	// Register bokio_exchange_token tool
	if err := server.RegisterTool("bokio_exchange_token", mcp.Tool{
		Name: "bokio_exchange_token",
		Description: "Exchange authorization code for access token",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"code": map[string]interface{}{
					"type": "string",
					"description": "Authorization code from OAuth2 callback",
					"required": true,
				},
			},
			"required": []string{"code"},
		},
		Handler: createExchangeTokenHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_exchange_token tool: %w", err)
	}

	// Register bokio_get_connections tool
	if err := server.RegisterTool("bokio_get_connections", mcp.Tool{
		Name: "bokio_get_connections",
		Description: "Get current connection information including user details and permissions",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
		Handler: createGetConnectionsHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_get_connections tool: %w", err)
	}

	// Register bokio_check_auth tool
	if err := server.RegisterTool("bokio_check_auth", mcp.Tool{
		Name: "bokio_check_auth",
		Description: "Check if the client is currently authenticated with valid tokens",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{},
		},
		Handler: createCheckAuthHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_check_auth tool: %w", err)
	}

	return nil
}

// createAuthenticateHandler creates the handler for the authenticate tool
func createAuthenticateHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		state := "default"
		if s, ok := params["state"].(string); ok && s != "" {
			state = s
		}

		authURL := client.GetAuthURL(state)

		return map[string]interface{}{
			"success": true,
			"auth_url": authURL,
			"message": "Visit the provided URL to authorize the application",
			"instructions": "After authorization, you will be redirected with a 'code' parameter. Use the bokio_exchange_token tool with this code to complete authentication.",
		}, nil
	}
}

// createExchangeTokenHandler creates the handler for the exchange token tool
func createExchangeTokenHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		code, ok := params["code"].(string)
		if !ok || code == "" {
			return nil, fmt.Errorf("authorization code is required")
		}

		token, err := client.ExchangeCodeForToken(ctx, code)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to exchange code for token: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"message": "Successfully authenticated with Bokio API",
			"token_type": token.TokenType,
			"expires_at": token.Expiry,
			"has_refresh_token": token.RefreshToken != "",
		}, nil
	}
}

// createGetConnectionsHandler creates the handler for the get connections tool
func createGetConnectionsHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		connInfo, err := client.GetConnectionInfo(ctx)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to get connection info: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"connection": map[string]interface{}{
				"user_id": connInfo.UserID,
				"company_id": connInfo.CompanyID,
				"company_name": connInfo.CompanyName,
				"email": connInfo.Email,
				"permissions": connInfo.Permissions,
			},
		}, nil
	}
}

// createCheckAuthHandler creates the handler for the check auth tool
func createCheckAuthHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		isAuthenticated := client.IsAuthenticated()
		
		response := map[string]interface{}{
			"authenticated": isAuthenticated,
		}

		if isAuthenticated {
			token := client.GetToken()
			if token != nil {
				response["token_type"] = token.TokenType
				response["expires_at"] = token.Expiry
				response["has_refresh_token"] = token.RefreshToken != ""
			}
		}

		return response, nil
	}
}