package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterCustomerTools registers customer-related MCP tools
func RegisterCustomerTools(server *mcp.Server, client *bokio.Client) error {
	// Register bokio_list_customers tool
	if err := server.RegisterTool("bokio_list_customers", mcp.Tool{
		Name: "bokio_list_customers",
		Description: "List customers with optional filtering and pagination",
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
				"search": map[string]interface{}{
					"type": "string",
					"description": "Search customers by name or email",
				},
			},
		},
		Handler: createListCustomersHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_list_customers tool: %w", err)
	}

	// Register bokio_get_customer tool
	if err := server.RegisterTool("bokio_get_customer", mcp.Tool{
		Name: "bokio_get_customer",
		Description: "Get a specific customer by ID",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
					"description": "Customer ID",
				},
			},
			"required": []string{"id"},
		},
		Handler: createGetCustomerHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_get_customer tool: %w", err)
	}

	// Register bokio_create_customer tool
	if err := server.RegisterTool("bokio_create_customer", mcp.Tool{
		Name: "bokio_create_customer",
		Description: "Create a new customer",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
					"description": "Customer name",
				},
				"email": map[string]interface{}{
					"type": "string",
					"format": "email",
					"description": "Customer email address",
				},
				"phone": map[string]interface{}{
					"type": "string",
					"description": "Customer phone number",
				},
				"organization_number": map[string]interface{}{
					"type": "string",
					"description": "Organization number",
				},
				"vat_number": map[string]interface{}{
					"type": "string",
					"description": "VAT number",
				},
				"address": map[string]interface{}{
					"type": "object",
					"description": "Customer address",
					"properties": map[string]interface{}{
						"street": map[string]interface{}{
							"type": "string",
							"description": "Street address",
						},
						"postal_code": map[string]interface{}{
							"type": "string",
							"description": "Postal code",
						},
						"city": map[string]interface{}{
							"type": "string",
							"description": "City",
						},
						"country": map[string]interface{}{
							"type": "string",
							"description": "Country",
						},
					},
				},
				"payment_terms": map[string]interface{}{
					"type": "integer",
					"description": "Payment terms in days",
					"minimum": 0,
				},
			},
			"required": []string{"name"},
		},
		Handler: createCreateCustomerHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_create_customer tool: %w", err)
	}

	// Register bokio_update_customer tool
	if err := server.RegisterTool("bokio_update_customer", mcp.Tool{
		Name: "bokio_update_customer",
		Description: "Update an existing customer",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
					"description": "Customer ID",
				},
				"name": map[string]interface{}{
					"type": "string",
					"description": "Customer name",
				},
				"email": map[string]interface{}{
					"type": "string",
					"format": "email",
					"description": "Customer email address",
				},
				"phone": map[string]interface{}{
					"type": "string",
					"description": "Customer phone number",
				},
				"organization_number": map[string]interface{}{
					"type": "string",
					"description": "Organization number",
				},
				"vat_number": map[string]interface{}{
					"type": "string",
					"description": "VAT number",
				},
				"address": map[string]interface{}{
					"type": "object",
					"description": "Customer address",
					"properties": map[string]interface{}{
						"street": map[string]interface{}{
							"type": "string",
							"description": "Street address",
						},
						"postal_code": map[string]interface{}{
							"type": "string",
							"description": "Postal code",
						},
						"city": map[string]interface{}{
							"type": "string",
							"description": "City",
						},
						"country": map[string]interface{}{
							"type": "string",
							"description": "Country",
						},
					},
				},
				"payment_terms": map[string]interface{}{
					"type": "integer",
					"description": "Payment terms in days",
					"minimum": 0,
				},
			},
			"required": []string{"id"},
		},
		Handler: createUpdateCustomerHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_update_customer tool: %w", err)
	}

	return nil
}

// createListCustomersHandler creates the handler for the list customers tool
func createListCustomersHandler(client *bokio.Client) mcp.ToolHandler {
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
		
		if search, ok := params["search"].(string); ok && search != "" {
			queryParams["search"] = search
		}

		// Construct URL with query parameters
		path := "/customers"
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
				"error": fmt.Sprintf("Failed to list customers: %v", err),
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var customerList bokio.ListResponse[bokio.Customer]
		if err := json.Unmarshal(resp.Body(), &customerList); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": customerList.Data,
			"pagination": customerList.Meta,
		}, nil
	}
}

// createGetCustomerHandler creates the handler for the get customer tool
func createGetCustomerHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		id, ok := params["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("customer ID is required")
		}

		resp, err := client.Get(ctx, "/customers/"+id)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to get customer: %v", err),
			}, nil
		}

		if resp.StatusCode() == http.StatusNotFound {
			return map[string]interface{}{
				"success": false,
				"error": "Customer not found",
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var customer bokio.Customer
		if err := json.Unmarshal(resp.Body(), &customer); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": customer,
		}, nil
	}
}

// createCreateCustomerHandler creates the handler for the create customer tool
func createCreateCustomerHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		// Parse and validate the request
		request, err := parseCreateCustomerRequest(params)
		if err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}

		resp, err := client.Post(ctx, "/customers", request)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to create customer: %v", err),
			}, nil
		}

		if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var customer bokio.Customer
		if err := json.Unmarshal(resp.Body(), &customer); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": customer,
			"message": "Customer created successfully",
		}, nil
	}
}

// createUpdateCustomerHandler creates the handler for the update customer tool
func createUpdateCustomerHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		id, ok := params["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("customer ID is required")
		}

		// Parse update request (only include provided fields)
		updateRequest := make(map[string]interface{})
		
		if name, ok := params["name"].(string); ok && name != "" {
			updateRequest["name"] = name
		}
		
		if email, ok := params["email"].(string); ok {
			updateRequest["email"] = email
		}
		
		if phone, ok := params["phone"].(string); ok {
			updateRequest["phone"] = phone
		}
		
		if orgNumber, ok := params["organization_number"].(string); ok {
			updateRequest["organization_number"] = orgNumber
		}
		
		if vatNumber, ok := params["vat_number"].(string); ok {
			updateRequest["vat_number"] = vatNumber
		}
		
		if paymentTerms, ok := params["payment_terms"]; ok {
			updateRequest["payment_terms"] = paymentTerms
		}

		if addressRaw, ok := params["address"].(map[string]interface{}); ok {
			address := &bokio.Address{}
			if street, ok := addressRaw["street"].(string); ok {
				address.Street = street
			}
			if postalCode, ok := addressRaw["postal_code"].(string); ok {
				address.PostalCode = postalCode
			}
			if city, ok := addressRaw["city"].(string); ok {
				address.City = city
			}
			if country, ok := addressRaw["country"].(string); ok {
				address.Country = country
			}
			updateRequest["address"] = address
		}

		resp, err := client.Patch(ctx, "/customers/"+id, updateRequest)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to update customer: %v", err),
			}, nil
		}

		if resp.StatusCode() == http.StatusNotFound {
			return map[string]interface{}{
				"success": false,
				"error": "Customer not found",
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var customer bokio.Customer
		if err := json.Unmarshal(resp.Body(), &customer); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": customer,
			"message": "Customer updated successfully",
		}, nil
	}
}

// parseCreateCustomerRequest parses the parameters into a CreateCustomerRequest
func parseCreateCustomerRequest(params map[string]interface{}) (*bokio.CreateCustomerRequest, error) {
	name, ok := params["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("customer name is required")
	}

	request := &bokio.CreateCustomerRequest{
		Name: name,
	}

	// Optional fields
	if email, ok := params["email"].(string); ok {
		request.Email = email
	}

	if phone, ok := params["phone"].(string); ok {
		request.Phone = phone
	}

	if orgNumber, ok := params["organization_number"].(string); ok {
		request.OrganizationNumber = orgNumber
	}

	if vatNumber, ok := params["vat_number"].(string); ok {
		request.VATNumber = vatNumber
	}

	if paymentTerms, ok := params["payment_terms"].(float64); ok {
		request.PaymentTerms = int(paymentTerms)
	} else if paymentTerms, ok := params["payment_terms"].(int); ok {
		request.PaymentTerms = paymentTerms
	}

	if addressRaw, ok := params["address"].(map[string]interface{}); ok {
		address := &bokio.Address{}
		if street, ok := addressRaw["street"].(string); ok {
			address.Street = street
		}
		if postalCode, ok := addressRaw["postal_code"].(string); ok {
			address.PostalCode = postalCode
		}
		if city, ok := addressRaw["city"].(string); ok {
			address.City = city
		}
		if country, ok := addressRaw["country"].(string); ok {
			address.Country = country
		}
		request.Address = address
	}

	return request, nil
}