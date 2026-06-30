package tools

import (
	"bytes"
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

// InvoiceListParams defines parameters for listing invoices
type InvoiceListParams struct {
	CompanyID string  `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	Page      *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32  `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
	Query     *string `json:"query,omitempty" jsonschema:"Optional query to filter the data set (optional)"`
}

// InvoiceCreateParams defines parameters for creating invoices
type InvoiceCreateParams struct {
	CompanyID string      `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	Invoice   interface{} `json:"invoice" jsonschema:"Invoice data object to create"`
}

// InvoiceGetParams defines parameters for getting a specific invoice
type InvoiceGetParams struct {
	CompanyID string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	InvoiceID string `json:"invoice_id" jsonschema:"Invoice UUID to retrieve"`
}

// InvoiceUpdateParams defines parameters for updating invoices
type InvoiceUpdateParams struct {
	CompanyID string      `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	InvoiceID string      `json:"invoice_id" jsonschema:"Invoice UUID to update"`
	Invoice   interface{} `json:"invoice" jsonschema:"Invoice data object with updates"`
}

// InvoiceLineItemsListParams defines parameters for listing invoice line items
type InvoiceLineItemsListParams struct {
	CompanyID string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	InvoiceID string `json:"invoice_id" jsonschema:"Invoice UUID to get line items for"`
}

// InvoiceLineItemsCreateParams defines parameters for creating invoice line items
type InvoiceLineItemsCreateParams struct {
	CompanyID string      `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	InvoiceID string      `json:"invoice_id" jsonschema:"Invoice UUID to add line item to"`
	LineItem  interface{} `json:"line_item" jsonschema:"Line item data object to create"`
}

// InvoiceResult defines the result structure for all invoice operations
type InvoiceResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RegisterInvoiceTools registers all invoice management tools using ONLY generated API clients
func RegisterInvoiceTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_list",
		Description: "List invoices for a company with optional pagination and filtering",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceListParams) (*mcp.CallToolResult, InvoiceResult, error) {
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
			}, InvoiceResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		genParams := &company.GetInvoiceParams{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Query,
		}

		resp, err := client.CompanyClient.GetInvoice(ctx, companyUUID, genParams)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to list invoices: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, InvoiceResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved invoices\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
				},
			},
		}, InvoiceResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_create",
		Description: "Create a new invoice for a company",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceCreateParams) (*mcp.CallToolResult, InvoiceResult, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Operation not allowed in read-only mode",
					},
				},
			}, InvoiceResult{}, nil
		}

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
			}, InvoiceResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		invoiceData, err := json.Marshal(args.Invoice)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid invoice data: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		var invoiceBody company.PostInvoiceJSONRequestBody
		if err := json.Unmarshal(invoiceData, &invoiceBody); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to parse invoice data: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		resp, err := client.CompanyClient.PostInvoice(ctx, companyUUID, invoiceBody)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to create invoice: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, InvoiceResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully created invoice\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
				},
			},
		}, InvoiceResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_get",
		Description: "Get a specific invoice by ID",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceGetParams) (*mcp.CallToolResult, InvoiceResult, error) {
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
			}, InvoiceResult{}, nil
		}

		if args.InvoiceID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invoice ID is required",
					},
				},
			}, InvoiceResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		invoiceUUID, err := uuid.Parse(args.InvoiceID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid invoice ID format: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		resp, err := client.CompanyClient.GetInvoicesInvoiceId(ctx, companyUUID, invoiceUUID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to get invoice: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, InvoiceResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved invoice\n\nCompany: %s\nInvoice: %s\nStatus: %d\nResponse: %v", companyIDStr, args.InvoiceID, resp.StatusCode, responseData),
				},
			},
		}, InvoiceResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_update",
		Description: "Update an existing invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceUpdateParams) (*mcp.CallToolResult, InvoiceResult, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Operation not allowed in read-only mode",
					},
				},
			}, InvoiceResult{}, nil
		}

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
			}, InvoiceResult{}, nil
		}

		if args.InvoiceID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invoice ID is required",
					},
				},
			}, InvoiceResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		invoiceUUID, err := uuid.Parse(args.InvoiceID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid invoice ID format: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		invoiceData, err := json.Marshal(args.Invoice)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid invoice data: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		var invoiceBody company.PutInvoiceJSONRequestBody
		if err := json.Unmarshal(invoiceData, &invoiceBody); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to parse invoice data: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		resp, err := client.CompanyClient.PutInvoice(ctx, companyUUID, invoiceUUID, invoiceBody)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to update invoice: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, InvoiceResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully updated invoice\n\nCompany: %s\nInvoice: %s\nStatus: %d\nResponse: %v", companyIDStr, args.InvoiceID, resp.StatusCode, responseData),
				},
			},
		}, InvoiceResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_line_items_list",
		Description: "List line items for a specific invoice (retrieves invoice details including line items)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceLineItemsListParams) (*mcp.CallToolResult, InvoiceResult, error) {
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
			}, InvoiceResult{}, nil
		}

		if args.InvoiceID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invoice ID is required",
					},
				},
			}, InvoiceResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		invoiceUUID, err := uuid.Parse(args.InvoiceID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid invoice ID format: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		resp, err := client.CompanyClient.GetInvoicesInvoiceId(ctx, companyUUID, invoiceUUID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to get invoice line items: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, InvoiceResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		var lineItems interface{}
		if respMap, ok := responseData.(map[string]interface{}); ok {
			if items, exists := respMap["lineItems"]; exists {
				lineItems = items
			}
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved invoice line items\n\nCompany: %s\nInvoice: %s\nStatus: %d\nLine Items: %v", companyIDStr, args.InvoiceID, resp.StatusCode, lineItems),
				},
			},
		}, InvoiceResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_line_items_create",
		Description: "Create a new line item for an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceLineItemsCreateParams) (*mcp.CallToolResult, InvoiceResult, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Operation not allowed in read-only mode",
					},
				},
			}, InvoiceResult{}, nil
		}

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
			}, InvoiceResult{}, nil
		}

		if args.InvoiceID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Invoice ID is required",
					},
				},
			}, InvoiceResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		invoiceUUID, err := uuid.Parse(args.InvoiceID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid invoice ID format: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		lineItemData, err := json.Marshal(args.LineItem)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid line item data: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		resp, err := client.CompanyClient.PostInvoiceLineItemWithBody(ctx, companyUUID, invoiceUUID, "application/json", bytes.NewReader(lineItemData))
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to create line item: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, InvoiceResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, InvoiceResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully created line item\n\nCompany: %s\nInvoice: %s\nStatus: %d\nResponse: %v", companyIDStr, args.InvoiceID, resp.StatusCode, responseData),
				},
			},
		}, InvoiceResult{}, nil
	})

	return nil
}
