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
	CompanyID string `json:"company_id"`
	Page      *int32 `json:"page,omitempty"`
	PageSize  *int32 `json:"page_size,omitempty"`
}

// GeneratedJournalResult defines the result using generated clients only
type GeneratedJournalResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RegisterGeneratedJournalTools registers journal tools using ONLY generated API clients
func RegisterGeneratedJournalTools(server *mcp.Server, client *bokio.AuthClient) error {
	// Tool to list journal entries using generated client
	listJournalTool := mcp.NewServerTool[GeneratedJournalParams, GeneratedJournalResult](
		"bokio_journal_entries_list",
		"List journal entries for a company with optional pagination",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[GeneratedJournalParams]) (*mcp.CallToolResultFor[GeneratedJournalResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[GeneratedJournalResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
						},
					},
				}, nil
			}

			// Parse company UUID
			companyUUID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return &mcp.CallToolResultFor[GeneratedJournalResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Create parameters for the generated client
			genParams := &company.GetJournalentryParams{
				Page:     params.Arguments.Page,
				PageSize: params.Arguments.PageSize,
			}

			// Call the generated client method
			resp, err := client.CompanyClient.GetJournalentry(ctx, companyUUID, genParams)
			if err != nil {
				return &mcp.CallToolResultFor[GeneratedJournalResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to list journal entries: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[GeneratedJournalResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
						},
					},
				}, nil
			}

			// Parse response body as generic interface
			var responseData interface{}
			if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
				return &mcp.CallToolResultFor[GeneratedJournalResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[GeneratedJournalResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully retrieved journal entries\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("page",
				mcp.Description("Page number (optional)"),
			),
			mcp.Property("page_size",
				mcp.Description("Items per page (optional)"),
			),
		),
	)

	server.AddTools(listJournalTool)
	return nil
}
