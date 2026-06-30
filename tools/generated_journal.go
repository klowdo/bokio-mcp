package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// GeneratedJournalParams defines parameters for the generated journal tool
type GeneratedJournalParams struct {
	CompanyID string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	Page      *int32 `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32 `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
}

// GeneratedJournalResult defines the result using generated clients only
type GeneratedJournalResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RegisterGeneratedJournalTools registers journal tools using ONLY generated API clients
func RegisterGeneratedJournalTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_journal_entries_list",
		Description: "List journal entries for a company with optional pagination",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args GeneratedJournalParams) (*mcp.CallToolResult, GeneratedJournalResult, error) {
		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, GeneratedJournalResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, GeneratedJournalResult{}, nil
		}

		genParams := &company.GetJournalentryParams{
			Page:     args.Page,
			PageSize: args.PageSize,
		}

		resp, err := client.CompanyClient.GetJournalentry(ctx, companyUUID, genParams)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to list journal entries: %v", err),
					},
				},
			}, GeneratedJournalResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, GeneratedJournalResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, GeneratedJournalResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved journal entries\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
				},
			},
		}, GeneratedJournalResult{}, nil
	})

	return nil
}
