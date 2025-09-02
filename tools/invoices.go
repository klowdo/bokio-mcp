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

// InvoiceListParams defines parameters for listing invoices
type InvoiceListParams struct {
	CompanyID string  `json:"company_id"`
	Page      *int32  `json:"page,omitempty"`
	PageSize  *int32  `json:"page_size,omitempty"`
	Query     *string `json:"query,omitempty"`
}

// InvoiceCreateParams defines parameters for creating invoices
type InvoiceCreateParams struct {
	CompanyID string      `json:"company_id"`
	Invoice   interface{} `json:"invoice"`
}

// InvoiceGetParams defines parameters for getting a specific invoice
type InvoiceGetParams struct {
	CompanyID string `json:"company_id"`
	InvoiceID string `json:"invoice_id"`
}

// InvoiceUpdateParams defines parameters for updating invoices
type InvoiceUpdateParams struct {
	CompanyID string      `json:"company_id"`
	InvoiceID string      `json:"invoice_id"`
	Invoice   interface{} `json:"invoice"`
}

// InvoiceLineItemsListParams defines parameters for listing invoice line items
type InvoiceLineItemsListParams struct {
	CompanyID string `json:"company_id"`
	InvoiceID string `json:"invoice_id"`
}

// InvoiceLineItemsCreateParams defines parameters for creating invoice line items
type InvoiceLineItemsCreateParams struct {
	CompanyID string      `json:"company_id"`
	InvoiceID string      `json:"invoice_id"`
	LineItem  interface{} `json:"line_item"`
}

// InvoiceResult defines the result structure for all invoice operations
type InvoiceResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RegisterInvoiceTools registers all invoice management tools using ONLY generated API clients
func RegisterInvoiceTools(server *mcp.Server, client *bokio.AuthClient) error {
	// Tool to list invoices with pagination and filtering
	listInvoicesTool := mcp.NewServerTool[InvoiceListParams, InvoiceResult](
		"bokio_invoices_list",
		"List invoices for a company with optional pagination and filtering",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[InvoiceListParams]) (*mcp.CallToolResultFor[InvoiceResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[InvoiceResult]{
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
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Create parameters for the generated client
			genParams := &company.GetInvoiceParams{
				Page:     params.Arguments.Page,
				PageSize: params.Arguments.PageSize,
				Query:    params.Arguments.Query,
			}

			// Call the generated client method
			resp, err := client.CompanyClient.GetInvoice(ctx, companyUUID, genParams)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to list invoices: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[InvoiceResult]{
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
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[InvoiceResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully retrieved invoices\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
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
			mcp.Property("query",
				mcp.Description("Optional query to filter the data set (optional)"),
			),
		),
	)

	// Tool to create a new invoice
	createInvoiceTool := mcp.NewServerTool[InvoiceCreateParams, InvoiceResult](
		"bokio_invoices_create",
		"Create a new invoice for a company",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[InvoiceCreateParams]) (*mcp.CallToolResultFor[InvoiceResult], error) {
			// Check if client is in read-only mode
			if client.GetConfig().ReadOnly {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Operation not allowed in read-only mode",
						},
					},
				}, nil
			}

			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[InvoiceResult]{
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
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Convert invoice data to proper type
			invoiceData, err := json.Marshal(params.Arguments.Invoice)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid invoice data: %v", err),
						},
					},
				}, nil
			}

			var invoiceBody company.PostInvoiceJSONRequestBody
			if err := json.Unmarshal(invoiceData, &invoiceBody); err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to parse invoice data: %v", err),
						},
					},
				}, nil
			}

			// Call the generated client method
			resp, err := client.CompanyClient.PostInvoice(ctx, companyUUID, invoiceBody)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to create invoice: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[InvoiceResult]{
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
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[InvoiceResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully created invoice\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("invoice",
				mcp.Description("Invoice data object to create"),
				mcp.Required(true),
			),
		),
	)

	// Tool to get a specific invoice by ID
	getInvoiceTool := mcp.NewServerTool[InvoiceGetParams, InvoiceResult](
		"bokio_invoices_get",
		"Get a specific invoice by ID",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[InvoiceGetParams]) (*mcp.CallToolResultFor[InvoiceResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
						},
					},
				}, nil
			}

			if params.Arguments.InvoiceID == "" {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Invoice ID is required",
						},
					},
				}, nil
			}

			// Parse company UUID
			companyUUID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Parse invoice UUID
			invoiceUUID, err := uuid.Parse(params.Arguments.InvoiceID)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid invoice ID format: %v", err),
						},
					},
				}, nil
			}

			// Call the generated client method
			resp, err := client.CompanyClient.GetInvoicesInvoiceId(ctx, companyUUID, invoiceUUID)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to get invoice: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[InvoiceResult]{
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
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[InvoiceResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully retrieved invoice\n\nCompany: %s\nInvoice: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.InvoiceID, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("invoice_id",
				mcp.Description("Invoice UUID to retrieve"),
				mcp.Required(true),
			),
		),
	)

	// Tool to update an invoice
	updateInvoiceTool := mcp.NewServerTool[InvoiceUpdateParams, InvoiceResult](
		"bokio_invoices_update",
		"Update an existing invoice",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[InvoiceUpdateParams]) (*mcp.CallToolResultFor[InvoiceResult], error) {
			// Check if client is in read-only mode
			if client.GetConfig().ReadOnly {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Operation not allowed in read-only mode",
						},
					},
				}, nil
			}

			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
						},
					},
				}, nil
			}

			if params.Arguments.InvoiceID == "" {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Invoice ID is required",
						},
					},
				}, nil
			}

			// Parse company UUID
			companyUUID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Parse invoice UUID
			invoiceUUID, err := uuid.Parse(params.Arguments.InvoiceID)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid invoice ID format: %v", err),
						},
					},
				}, nil
			}

			// Convert invoice data to proper type
			invoiceData, err := json.Marshal(params.Arguments.Invoice)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid invoice data: %v", err),
						},
					},
				}, nil
			}

			var invoiceBody company.PutInvoiceJSONRequestBody
			if err := json.Unmarshal(invoiceData, &invoiceBody); err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to parse invoice data: %v", err),
						},
					},
				}, nil
			}

			// Call the generated client method
			resp, err := client.CompanyClient.PutInvoice(ctx, companyUUID, invoiceUUID, invoiceBody)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to update invoice: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[InvoiceResult]{
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
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[InvoiceResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully updated invoice\n\nCompany: %s\nInvoice: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.InvoiceID, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("invoice_id",
				mcp.Description("Invoice UUID to update"),
				mcp.Required(true),
			),
			mcp.Property("invoice",
				mcp.Description("Invoice data object with updates"),
				mcp.Required(true),
			),
		),
	)

	// Tool to list invoice line items (gets invoice details including line items)
	listLineItemsTool := mcp.NewServerTool[InvoiceLineItemsListParams, InvoiceResult](
		"bokio_invoices_line_items_list",
		"List line items for a specific invoice (retrieves invoice details including line items)",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[InvoiceLineItemsListParams]) (*mcp.CallToolResultFor[InvoiceResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
						},
					},
				}, nil
			}

			if params.Arguments.InvoiceID == "" {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Invoice ID is required",
						},
					},
				}, nil
			}

			// Parse company UUID
			companyUUID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Parse invoice UUID
			invoiceUUID, err := uuid.Parse(params.Arguments.InvoiceID)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid invoice ID format: %v", err),
						},
					},
				}, nil
			}

			// Call the generated client method to get invoice details (including line items)
			resp, err := client.CompanyClient.GetInvoicesInvoiceId(ctx, companyUUID, invoiceUUID)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to get invoice line items: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[InvoiceResult]{
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
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Extract line items from the invoice response
			var lineItems interface{}
			if respMap, ok := responseData.(map[string]interface{}); ok {
				if items, exists := respMap["lineItems"]; exists {
					lineItems = items
				}
			}

			// Return success with line items data
			return &mcp.CallToolResultFor[InvoiceResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully retrieved invoice line items\n\nCompany: %s\nInvoice: %s\nStatus: %d\nLine Items: %v", companyIDStr, params.Arguments.InvoiceID, resp.StatusCode, lineItems),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("invoice_id",
				mcp.Description("Invoice UUID to get line items for"),
				mcp.Required(true),
			),
		),
	)

	// Tool to create a new invoice line item
	createLineItemTool := mcp.NewServerTool[InvoiceLineItemsCreateParams, InvoiceResult](
		"bokio_invoices_line_items_create",
		"Create a new line item for an invoice",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[InvoiceLineItemsCreateParams]) (*mcp.CallToolResultFor[InvoiceResult], error) {
			// Check if client is in read-only mode
			if client.GetConfig().ReadOnly {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Operation not allowed in read-only mode",
						},
					},
				}, nil
			}

			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
						},
					},
				}, nil
			}

			if params.Arguments.InvoiceID == "" {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Invoice ID is required",
						},
					},
				}, nil
			}

			// Parse company UUID
			companyUUID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Parse invoice UUID
			invoiceUUID, err := uuid.Parse(params.Arguments.InvoiceID)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid invoice ID format: %v", err),
						},
					},
				}, nil
			}

			// Convert line item data to proper type
			lineItemData, err := json.Marshal(params.Arguments.LineItem)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid line item data: %v", err),
						},
					},
				}, nil
			}

			var lineItemBody company.PostInvoiceLineItemJSONRequestBody
			if err := json.Unmarshal(lineItemData, &lineItemBody); err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to parse line item data: %v", err),
						},
					},
				}, nil
			}

			// Call the generated client method
			resp, err := client.CompanyClient.PostInvoiceLineItem(ctx, companyUUID, invoiceUUID, lineItemBody)
			if err != nil {
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to create line item: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[InvoiceResult]{
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
				return &mcp.CallToolResultFor[InvoiceResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[InvoiceResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully created line item\n\nCompany: %s\nInvoice: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.InvoiceID, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("invoice_id",
				mcp.Description("Invoice UUID to add line item to"),
				mcp.Required(true),
			),
			mcp.Property("line_item",
				mcp.Description("Line item data object to create"),
				mcp.Required(true),
			),
		),
	)

	// Add all tools to the server
	server.AddTools(
		listInvoicesTool,
		createInvoiceTool,
		getInvoiceTool,
		updateInvoiceTool,
		listLineItemsTool,
		createLineItemTool,
	)

	return nil
}
