package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterInvoiceTools registers invoice-related MCP tools
func RegisterInvoiceTools(server *mcp.Server, client *bokio.Client) error {
	// Register bokio_list_invoices tool
	if err := server.RegisterTool("bokio_list_invoices", mcp.Tool{
		Name: "bokio_list_invoices",
		Description: "List invoices with optional filtering and pagination",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page": map[string]interface{}{
					"type": "integer",
					"description": "Page number for pagination (default: 1)",
					"minimum": 1,
				},
				"per_page": map[string]interface{}{
					"type": "integer",
					"description": "Number of items per page (default: 25, max: 100)",
					"minimum": 1,
					"maximum": 100,
				},
				"status": map[string]interface{}{
					"type": "string",
					"description": "Filter by invoice status",
					"enum": []string{"draft", "sent", "paid", "overdue", "cancelled"},
				},
				"customer_id": map[string]interface{}{
					"type": "string",
					"description": "Filter by customer ID",
				},
			},
		},
		Handler: createListInvoicesHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_list_invoices tool: %w", err)
	}

	// Register bokio_get_invoice tool
	if err := server.RegisterTool("bokio_get_invoice", mcp.Tool{
		Name: "bokio_get_invoice",
		Description: "Get a specific invoice by ID",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
					"description": "Invoice ID",
				},
			},
			"required": []string{"id"},
		},
		Handler: createGetInvoiceHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_get_invoice tool: %w", err)
	}

	// Register bokio_create_invoice tool
	if err := server.RegisterTool("bokio_create_invoice", mcp.Tool{
		Name: "bokio_create_invoice",
		Description: "Create a new invoice",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"customer_id": map[string]interface{}{
					"type": "string",
					"description": "ID of the customer",
				},
				"date": map[string]interface{}{
					"type": "string",
					"format": "date",
					"description": "Invoice date (YYYY-MM-DD)",
				},
				"due_date": map[string]interface{}{
					"type": "string",
					"format": "date",
					"description": "Due date (YYYY-MM-DD)",
				},
				"description": map[string]interface{}{
					"type": "string",
					"description": "Invoice description",
				},
				"items": map[string]interface{}{
					"type": "array",
					"description": "Invoice line items",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"description": map[string]interface{}{
								"type": "string",
								"description": "Item description",
							},
							"quantity": map[string]interface{}{
								"type": "number",
								"description": "Quantity",
								"minimum": 0,
							},
							"unit_price": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"amount": map[string]interface{}{
										"type": "number",
										"description": "Price amount",
									},
									"currency": map[string]interface{}{
										"type": "string",
										"description": "Currency code",
										"default": "SEK",
									},
								},
								"required": []string{"amount"},
							},
							"vat_rate": map[string]interface{}{
								"type": "number",
								"description": "VAT rate as decimal (e.g., 0.25 for 25%)",
								"minimum": 0,
								"maximum": 1,
							},
						},
						"required": []string{"description", "quantity", "unit_price", "vat_rate"},
					},
				},
			},
			"required": []string{"customer_id", "items"},
		},
		Handler: createCreateInvoiceHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_create_invoice tool: %w", err)
	}

	// Register bokio_update_invoice tool
	if err := server.RegisterTool("bokio_update_invoice", mcp.Tool{
		Name: "bokio_update_invoice",
		Description: "Update an existing invoice",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
					"description": "Invoice ID",
				},
				"customer_id": map[string]interface{}{
					"type": "string",
					"description": "ID of the customer",
				},
				"date": map[string]interface{}{
					"type": "string",
					"format": "date",
					"description": "Invoice date (YYYY-MM-DD)",
				},
				"due_date": map[string]interface{}{
					"type": "string",
					"format": "date",
					"description": "Due date (YYYY-MM-DD)",
				},
				"description": map[string]interface{}{
					"type": "string",
					"description": "Invoice description",
				},
			},
			"required": []string{"id"},
		},
		Handler: createUpdateInvoiceHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_update_invoice tool: %w", err)
	}

	return nil
}

// createListInvoicesHandler creates the handler for the list invoices tool
func createListInvoicesHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		// Build query parameters
		queryParams := make(map[string]string)
		
		if page, ok := params["page"]; ok {
			queryParams["page"] = fmt.Sprintf("%v", page)
		}
		
		if perPage, ok := params["per_page"]; ok {
			queryParams["per_page"] = fmt.Sprintf("%v", perPage)
		}
		
		if status, ok := params["status"].(string); ok && status != "" {
			queryParams["status"] = status
		}
		
		if customerID, ok := params["customer_id"].(string); ok && customerID != "" {
			queryParams["customer_id"] = customerID
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

		resp, err := client.Get(ctx, path)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to list invoices: %v", err),
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var invoiceList bokio.ListResponse[bokio.Invoice]
		if err := json.Unmarshal(resp.Body(), &invoiceList); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": invoiceList.Data,
			"pagination": invoiceList.Meta,
		}, nil
	}
}

// createGetInvoiceHandler creates the handler for the get invoice tool
func createGetInvoiceHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		id, ok := params["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("invoice ID is required")
		}

		resp, err := client.Get(ctx, "/invoices/"+id)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to get invoice: %v", err),
			}, nil
		}

		if resp.StatusCode() == http.StatusNotFound {
			return map[string]interface{}{
				"success": false,
				"error": "Invoice not found",
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var invoice bokio.Invoice
		if err := json.Unmarshal(resp.Body(), &invoice); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": invoice,
		}, nil
	}
}

// createCreateInvoiceHandler creates the handler for the create invoice tool
func createCreateInvoiceHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		// Parse and validate the request
		request, err := parseCreateInvoiceRequest(params)
		if err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}

		resp, err := client.Post(ctx, "/invoices", request)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to create invoice: %v", err),
			}, nil
		}

		if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var invoice bokio.Invoice
		if err := json.Unmarshal(resp.Body(), &invoice); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": invoice,
			"message": "Invoice created successfully",
		}, nil
	}
}

// createUpdateInvoiceHandler creates the handler for the update invoice tool
func createUpdateInvoiceHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		id, ok := params["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("invoice ID is required")
		}

		// Parse update request (only include non-nil fields)
		updateRequest := make(map[string]interface{})
		
		if customerID, ok := params["customer_id"].(string); ok && customerID != "" {
			updateRequest["customer_id"] = customerID
		}
		
		if date, ok := params["date"].(string); ok && date != "" {
			updateRequest["date"] = date
		}
		
		if dueDate, ok := params["due_date"].(string); ok && dueDate != "" {
			updateRequest["due_date"] = dueDate
		}
		
		if description, ok := params["description"].(string); ok {
			updateRequest["description"] = description
		}

		resp, err := client.Patch(ctx, "/invoices/"+id, updateRequest)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to update invoice: %v", err),
			}, nil
		}

		if resp.StatusCode() == http.StatusNotFound {
			return map[string]interface{}{
				"success": false,
				"error": "Invoice not found",
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var invoice bokio.Invoice
		if err := json.Unmarshal(resp.Body(), &invoice); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": invoice,
			"message": "Invoice updated successfully",
		}, nil
	}
}

// parseCreateInvoiceRequest parses the parameters into a CreateInvoiceRequest
func parseCreateInvoiceRequest(params map[string]interface{}) (*bokio.CreateInvoiceRequest, error) {
	customerID, ok := params["customer_id"].(string)
	if !ok || customerID == "" {
		return nil, fmt.Errorf("customer_id is required")
	}

	itemsRaw, ok := params["items"].([]interface{})
	if !ok || len(itemsRaw) == 0 {
		return nil, fmt.Errorf("items are required")
	}

	items := make([]bokio.InvoiceItem, len(itemsRaw))
	for i, itemRaw := range itemsRaw {
		itemMap, ok := itemRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid item at index %d", i)
		}

		description, ok := itemMap["description"].(string)
		if !ok || description == "" {
			return nil, fmt.Errorf("item description is required at index %d", i)
		}

		quantity, ok := itemMap["quantity"].(float64)
		if !ok {
			// Try to parse as integer
			if qInt, ok := itemMap["quantity"].(int); ok {
				quantity = float64(qInt)
			} else {
				return nil, fmt.Errorf("item quantity is required at index %d", i)
			}
		}

		unitPriceRaw, ok := itemMap["unit_price"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("item unit_price is required at index %d", i)
		}

		amount, ok := unitPriceRaw["amount"].(float64)
		if !ok {
			// Try to parse as integer
			if amtInt, ok := unitPriceRaw["amount"].(int); ok {
				amount = float64(amtInt)
			} else {
				return nil, fmt.Errorf("unit_price amount is required at index %d", i)
			}
		}

		currency, ok := unitPriceRaw["currency"].(string)
		if !ok || currency == "" {
			currency = "SEK" // Default currency
		}

		vatRate, ok := itemMap["vat_rate"].(float64)
		if !ok {
			return nil, fmt.Errorf("item vat_rate is required at index %d", i)
		}

		items[i] = bokio.InvoiceItem{
			Description: description,
			Quantity:    quantity,
			UnitPrice: bokio.Money{
				Amount:   amount,
				Currency: currency,
			},
			VATRate: vatRate,
		}
	}

	request := &bokio.CreateInvoiceRequest{
		CustomerID: customerID,
		Items:      items,
	}

	// Optional fields
	if date, ok := params["date"].(string); ok && date != "" {
		// In a real implementation, parse the date string to time.Time
		// For now, we'll leave it as nil and let the API handle it
	}

	if dueDate, ok := params["due_date"].(string); ok && dueDate != "" {
		// In a real implementation, parse the date string to time.Time
		// For now, we'll leave it as nil and let the API handle it
	}

	if description, ok := params["description"].(string); ok {
		request.Description = description
	}

	return request, nil
}