package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ListInvoicesParams defines the parameters for listing invoices
type ListInvoicesParams struct {
	Page       *int    `json:"page,omitempty"`
	PerPage    *int    `json:"per_page,omitempty"`
	Status     *string `json:"status,omitempty"`
	CustomerID *string `json:"customer_id,omitempty"`
}

// ListInvoicesResult defines the result of listing invoices
type ListInvoicesResult struct {
	Success    bool                   `json:"success"`
	Data       []bokio.Invoice        `json:"data,omitempty"`
	Pagination interface{}            `json:"pagination,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// GetInvoiceParams defines the parameters for getting an invoice
type GetInvoiceParams struct {
	ID string `json:"id"`
}

// GetInvoiceResult defines the result of getting an invoice
type GetInvoiceResult struct {
	Success bool           `json:"success"`
	Data    *bokio.Invoice `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// InvoiceItemParams defines the parameters for an invoice item
type InvoiceItemParams struct {
	Description string           `json:"description"`
	Quantity    float64          `json:"quantity"`
	UnitPrice   MoneyParams      `json:"unit_price"`
	VATRate     float64          `json:"vat_rate"`
}

// MoneyParams defines the parameters for money values
type MoneyParams struct {
	Amount   float64 `json:"amount"`
	Currency *string `json:"currency,omitempty"`
}

// CreateInvoiceParams defines the parameters for creating an invoice
type CreateInvoiceParams struct {
	CustomerID  string              `json:"customer_id"`
	Date        *string             `json:"date,omitempty"`
	DueDate     *string             `json:"due_date,omitempty"`
	Description *string             `json:"description,omitempty"`
	Items       []InvoiceItemParams `json:"items"`
}

// CreateInvoiceResult defines the result of creating an invoice
type CreateInvoiceResult struct {
	Success bool           `json:"success"`
	Data    *bokio.Invoice `json:"data,omitempty"`
	Message string         `json:"message,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// UpdateInvoiceParams defines the parameters for updating an invoice
type UpdateInvoiceParams struct {
	ID          string  `json:"id"`
	CustomerID  *string `json:"customer_id,omitempty"`
	Date        *string `json:"date,omitempty"`
	DueDate     *string `json:"due_date,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateInvoiceResult defines the result of updating an invoice
type UpdateInvoiceResult struct {
	Success bool           `json:"success"`
	Data    *bokio.Invoice `json:"data,omitempty"`
	Message string         `json:"message,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// RegisterInvoiceTools registers invoice-related MCP tools
func RegisterInvoiceTools(server *mcp.Server, client *bokio.Client) error {
	// Register bokio_list_invoices tool
	listInvoicesTool := mcp.NewServerTool[ListInvoicesParams, ListInvoicesResult](
		"bokio_list_invoices",
		"List invoices with optional filtering and pagination",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ListInvoicesParams]) (*mcp.CallToolResultFor[ListInvoicesResult], error) {
			if !client.IsAuthenticated() {
				return &mcp.CallToolResultFor[ListInvoicesResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Not authenticated. Use bokio_authenticate first.",
						},
					},
				}, nil
			}

			// Build query parameters
			queryParams := make(map[string]string)
			
			if params.Arguments.Page != nil {
				queryParams["page"] = fmt.Sprintf("%d", *params.Arguments.Page)
			}
			
			if params.Arguments.PerPage != nil {
				queryParams["per_page"] = fmt.Sprintf("%d", *params.Arguments.PerPage)
			}
			
			if params.Arguments.Status != nil && *params.Arguments.Status != "" {
				queryParams["status"] = *params.Arguments.Status
			}
			
			if params.Arguments.CustomerID != nil && *params.Arguments.CustomerID != "" {
				queryParams["customer_id"] = *params.Arguments.CustomerID
			}

			// Construct URL with query parameters
			path := "/invoices"
			if len(queryParams) > 0 {
				path += "?"
				first := true
				for key, value := range queryParams {
					if !first {
						path += "&"
					}
					path += key + "=" + value
					first = false
				}
			}

			resp, err := client.GET(ctx, path)
			if err != nil {
				return &mcp.CallToolResultFor[ListInvoicesResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to list invoices: %v", err),
						},
					},
				}, nil
			}

			if resp.StatusCode() != http.StatusOK {
				return &mcp.CallToolResultFor[ListInvoicesResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
						},
					},
				}, nil
			}

			var invoiceList bokio.InvoicesResponse
			if err := json.Unmarshal(resp.Body(), &invoiceList); err != nil {
				return &mcp.CallToolResultFor[ListInvoicesResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to parse response: %v", err),
						},
					},
				}, nil
			}

			return &mcp.CallToolResultFor[ListInvoicesResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Found %d invoices", len(invoiceList.Items)),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("page",
				mcp.Description("Page number for pagination (default: 1)"),
				),
			mcp.Property("per_page",
				mcp.Description("Number of items per page (default: 25, max: 100)"),
					),
			mcp.Property("status",
				mcp.Description("Filter by invoice status"),
					),
			mcp.Property("customer_id",
				mcp.Description("Filter by customer ID"),
			),
		),
	)
	
	server.AddTools(listInvoicesTool)

	// Register bokio_get_invoice tool
	getInvoiceTool := mcp.NewServerTool[GetInvoiceParams, GetInvoiceResult](
		"bokio_get_invoice",
		"Get a specific invoice by ID",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[GetInvoiceParams]) (*mcp.CallToolResultFor[GetInvoiceResult], error) {
			if !client.IsAuthenticated() {
				return &mcp.CallToolResultFor[GetInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Not authenticated. Use bokio_authenticate first.",
						},
					},
				}, nil
			}

			id := params.Arguments.ID
			if id == "" {
				return &mcp.CallToolResultFor[GetInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Invoice ID is required",
						},
					},
				}, fmt.Errorf("invoice ID is required")
			}

			resp, err := client.GET(ctx, "/invoices/"+id)
			if err != nil {
				return &mcp.CallToolResultFor[GetInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to get invoice: %v", err),
						},
					},
				}, nil
			}

			if resp.StatusCode() == http.StatusNotFound {
				return &mcp.CallToolResultFor[GetInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Invoice not found",
						},
					},
				}, nil
			}

			if resp.StatusCode() != http.StatusOK {
				return &mcp.CallToolResultFor[GetInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
						},
					},
				}, nil
			}

			var invoice bokio.Invoice
			if err := json.Unmarshal(resp.Body(), &invoice); err != nil {
				return &mcp.CallToolResultFor[GetInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to parse response: %v", err),
						},
					},
				}, nil
			}

			return &mcp.CallToolResultFor[GetInvoiceResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invoice: %s (ID: %s)", invoice.InvoiceNumber, invoice.ID),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("id",
				mcp.Description("Invoice ID"),
				mcp.Required(true),
			),
		),
	)
	
	server.AddTools(getInvoiceTool)

	// Register bokio_create_invoice tool
	createInvoiceTool := mcp.NewServerTool[CreateInvoiceParams, CreateInvoiceResult](
		"bokio_create_invoice",
		"Create a new invoice",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateInvoiceParams]) (*mcp.CallToolResultFor[CreateInvoiceResult], error) {
			if !client.IsAuthenticated() {
				return &mcp.CallToolResultFor[CreateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Not authenticated. Use bokio_authenticate first.",
						},
					},
				}, nil
			}

			// Parse and validate the request
			request, err := parseCreateInvoiceRequestFromParams(params.Arguments)
			if err != nil {
				return &mcp.CallToolResultFor[CreateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid request: %v", err),
						},
					},
				}, fmt.Errorf("invalid request: %w", err)
			}

			resp, err := client.POST(ctx, "/invoices", request)
			if err != nil {
				return &mcp.CallToolResultFor[CreateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to create invoice: %v", err),
						},
					},
				}, nil
			}

			if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
				return &mcp.CallToolResultFor[CreateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
						},
					},
				}, nil
			}

			var invoice bokio.Invoice
			if err := json.Unmarshal(resp.Body(), &invoice); err != nil {
				return &mcp.CallToolResultFor[CreateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to parse response: %v", err),
						},
					},
				}, nil
			}

			return &mcp.CallToolResultFor[CreateInvoiceResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invoice created successfully: %s (ID: %s)", invoice.InvoiceNumber, invoice.ID),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("customer_id",
				mcp.Description("ID of the customer"),
				mcp.Required(true),
			),
			mcp.Property("date",
				mcp.Description("Invoice date (YYYY-MM-DD)"),
					),
			mcp.Property("due_date",
				mcp.Description("Due date (YYYY-MM-DD)"),
					),
			mcp.Property("description",
				mcp.Description("Invoice description"),
			),
			mcp.Property("items",
				mcp.Description("Invoice line items"),
				mcp.Required(true),
			),
		),
	)
	
	server.AddTools(createInvoiceTool)

	// Register bokio_update_invoice tool
	updateInvoiceTool := mcp.NewServerTool[UpdateInvoiceParams, UpdateInvoiceResult](
		"bokio_update_invoice",
		"Update an existing invoice",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[UpdateInvoiceParams]) (*mcp.CallToolResultFor[UpdateInvoiceResult], error) {
			if !client.IsAuthenticated() {
				return &mcp.CallToolResultFor[UpdateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Not authenticated. Use bokio_authenticate first.",
						},
					},
				}, nil
			}

			id := params.Arguments.ID
			if id == "" {
				return &mcp.CallToolResultFor[UpdateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Invoice ID is required",
						},
					},
				}, fmt.Errorf("invoice ID is required")
			}

			// Parse update request (only include non-nil fields)
			updateRequest := buildUpdateInvoiceRequest(params.Arguments)

			resp, err := client.PATCH(ctx, "/invoices/"+id, updateRequest)
			if err != nil {
				return &mcp.CallToolResultFor[UpdateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to update invoice: %v", err),
						},
					},
				}, nil
			}

			if resp.StatusCode() == http.StatusNotFound {
				return &mcp.CallToolResultFor[UpdateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Invoice not found",
						},
					},
				}, nil
			}

			if resp.StatusCode() != http.StatusOK {
				return &mcp.CallToolResultFor[UpdateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
						},
					},
				}, nil
			}

			var invoice bokio.Invoice
			if err := json.Unmarshal(resp.Body(), &invoice); err != nil {
				return &mcp.CallToolResultFor[UpdateInvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to parse response: %v", err),
						},
					},
				}, nil
			}

			return &mcp.CallToolResultFor[UpdateInvoiceResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invoice updated successfully: %s (ID: %s)", invoice.InvoiceNumber, invoice.ID),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("id",
				mcp.Description("Invoice ID"),
				mcp.Required(true),
			),
			mcp.Property("customer_id",
				mcp.Description("ID of the customer"),
			),
			mcp.Property("date",
				mcp.Description("Invoice date (YYYY-MM-DD)"),
					),
			mcp.Property("due_date",
				mcp.Description("Due date (YYYY-MM-DD)"),
					),
			mcp.Property("description",
				mcp.Description("Invoice description"),
			),
		),
	)
	
	server.AddTools(updateInvoiceTool)

	return nil
}





// parseCreateInvoiceRequestFromParams parses the new typed parameters into a CreateInvoiceRequest
func parseCreateInvoiceRequestFromParams(params CreateInvoiceParams) (*bokio.CreateInvoiceRequest, error) {
	if params.CustomerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	if len(params.Items) == 0 {
		return nil, fmt.Errorf("items are required")
	}

	items := make([]bokio.InvoiceItem, len(params.Items))
	for i, item := range params.Items {
		if item.Description == "" {
			return nil, fmt.Errorf("item description is required at index %d", i)
		}

		items[i] = bokio.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			Price:       item.UnitPrice.Amount,
			VatRate:     item.VATRate,
		}
	}

	request := &bokio.CreateInvoiceRequest{
		CustomerID: params.CustomerID,
		Items:      items,
	}

	// Optional fields
	if params.Date != nil {
		// In a real implementation, parse the date string to time.Time
		// For now, we'll leave it as nil and let the API handle it
	}

	if params.DueDate != nil {
		// In a real implementation, parse the date string to time.Time
		// For now, we'll leave it as nil and let the API handle it
	}

	if params.Description != nil {
		request.Notes = *params.Description
	}

	return request, nil
}

// buildUpdateInvoiceRequest builds an update request from typed parameters
func buildUpdateInvoiceRequest(params UpdateInvoiceParams) map[string]interface{} {
	updateRequest := make(map[string]interface{})
	
	if params.CustomerID != nil && *params.CustomerID != "" {
		updateRequest["customer_id"] = *params.CustomerID
	}
	
	if params.Date != nil && *params.Date != "" {
		updateRequest["date"] = *params.Date
	}
	
	if params.DueDate != nil && *params.DueDate != "" {
		updateRequest["due_date"] = *params.DueDate
	}
	
	if params.Description != nil {
		updateRequest["description"] = *params.Description
	}

	return updateRequest
}