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

// InvoicePaymentsListParams defines parameters for listing invoice payments
type InvoicePaymentsListParams struct {
	CompanyID string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	InvoiceID string `json:"invoice_id" jsonschema:"Invoice UUID to get payments for"`
	Page      *int32 `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32 `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
}

// InvoiceSettlementsListParams defines parameters for listing invoice settlements
type InvoiceSettlementsListParams struct {
	CompanyID string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	InvoiceID string `json:"invoice_id" jsonschema:"Invoice UUID to get settlements for"`
	Page      *int32 `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32 `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
}

func toolError(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}},
	}
}

func resolveInvoiceRef(companyID, invoiceID string) (uuid.UUID, uuid.UUID, *mcp.CallToolResult) {
	if companyID == "" {
		companyID = os.Getenv("BOKIO_COMPANY_ID")
	}
	if companyID == "" {
		return uuid.Nil, uuid.Nil, toolError("Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)")
	}

	companyUUID, err := uuid.Parse(companyID)
	if err != nil {
		return uuid.Nil, uuid.Nil, toolError("Invalid company ID format: %v", err)
	}

	invoiceUUID, err := uuid.Parse(invoiceID)
	if err != nil {
		return uuid.Nil, uuid.Nil, toolError("Invalid invoice ID format: %v", err)
	}

	return companyUUID, invoiceUUID, nil
}

func renderAPIResponse(label string, resp *http.Response, err error) (*mcp.CallToolResult, any, error) {
	if err != nil {
		return toolError("Failed to retrieve %s: %v", label, err), nil, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return toolError("API returned status %d", resp.StatusCode), nil, nil
	}

	var responseData json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return toolError("Failed to decode response: %v", err), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("✅ Successfully retrieved %s\n\nStatus: %d\nResponse: %s", label, resp.StatusCode, responseData),
			},
		},
	}, nil, nil
}

// RegisterPaymentTools registers invoice payment and settlement tools using generated API clients
func RegisterPaymentTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_payments_list",
		Description: "List payments recorded against an invoice. Use the latest payment date to determine when an invoice was actually paid.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoicePaymentsListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetInvoicePayments(ctx, companyUUID, invoiceUUID, &company.GetInvoicePaymentsParams{
			Page:     args.Page,
			PageSize: args.PageSize,
		})

		return renderAPIResponse("invoice payments", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_settlements_list",
		Description: "List settlements on an invoice (currency differences, write-offs). Settlements clear the receivable without money moving, so they are not payment dates.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceSettlementsListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetInvoiceSettlements(ctx, companyUUID, invoiceUUID, &company.GetInvoiceSettlementsParams{
			Page:     args.Page,
			PageSize: args.PageSize,
		})

		return renderAPIResponse("invoice settlements", resp, err)
	})

	return nil
}
