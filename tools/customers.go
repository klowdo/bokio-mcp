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
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// CustomersListParams defines parameters for listing customers
type CustomersListParams struct {
	CompanyID string  `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	Page      *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32  `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
	Search    *string `json:"search,omitempty" jsonschema:"Search customers by name or email (optional)"`
}

// CustomerCreateParams defines parameters for creating a customer
type CustomerCreateParams struct {
	CompanyID          string  `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	Name               string  `json:"name" jsonschema:"Customer name"`
	Email              *string `json:"email,omitempty" jsonschema:"Customer email address (optional)"`
	Phone              *string `json:"phone,omitempty" jsonschema:"Customer phone number (optional)"`
	OrganizationNumber *string `json:"organization_number,omitempty" jsonschema:"Organization number (optional)"`
	VatNumber          *string `json:"vat_number,omitempty" jsonschema:"VAT number (optional)"`
	Type               string  `json:"type" jsonschema:"Customer type: 'company' or 'private'"`
	PaymentTerms       *int    `json:"payment_terms,omitempty" jsonschema:"Payment terms in days (optional)"`
}

// CustomerGetParams defines parameters for getting a customer
type CustomerGetParams struct {
	CompanyID  string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	CustomerID string `json:"customer_id" jsonschema:"Customer UUID"`
}

// CustomerUpdateParams defines parameters for updating a customer
type CustomerUpdateParams struct {
	CompanyID          string  `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	CustomerID         string  `json:"customer_id" jsonschema:"Customer UUID"`
	Name               *string `json:"name,omitempty" jsonschema:"Customer name (optional)"`
	Email              *string `json:"email,omitempty" jsonschema:"Customer email address (optional)"`
	Phone              *string `json:"phone,omitempty" jsonschema:"Customer phone number (optional)"`
	OrganizationNumber *string `json:"organization_number,omitempty" jsonschema:"Organization number (optional)"`
	VatNumber          *string `json:"vat_number,omitempty" jsonschema:"VAT number (optional)"`
	Type               *string `json:"type,omitempty" jsonschema:"Customer type: 'company' or 'private' (optional)"`
	PaymentTerms       *int    `json:"payment_terms,omitempty" jsonschema:"Payment terms in days (optional)"`
}

// RegisterCustomerTools registers customer-related MCP tools using generated API clients
func RegisterCustomerTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_customers_list",
		Description: "List customers for a company with optional pagination and filtering",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CustomersListParams) (*mcp.CallToolResult, any, error) {
		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, nil, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, nil, nil
		}

		genParams := &company.GetCustomerParams{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Search,
		}

		resp, err := client.CompanyClient.GetCustomer(ctx, companyUUID, genParams)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to list customers: %v", err),
					},
				},
			}, nil, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, nil, nil
		}

		var responseData json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved customers\n\nCompany: %s\nStatus: %d\nResponse: %s", companyIDStr, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_customers_create",
		Description: "Create a new customer for a company",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CustomerCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Operation not allowed in read-only mode",
					},
				},
			}, nil, nil
		}

		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, nil, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, nil, nil
		}

		if args.Name == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer name is required",
					},
				},
			}, nil, nil
		}

		customerType := company.CustomerType(args.Type)
		if customerType != company.Company && customerType != company.Private {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer type must be 'company' or 'private'",
					},
				},
			}, nil, nil
		}

		customer := company.Customer{
			Name: args.Name,
			Type: customerType,
		}

		if args.Email != nil || args.Phone != nil {
			contactDetails := []struct {
				Email     *string             `json:"email,omitempty"`
				Id        *openapi_types.UUID `json:"id,omitempty"`
				IsDefault *bool               `json:"isDefault,omitempty"`
				Name      *string             `json:"name,omitempty"`
				Phone     *string             `json:"phone,omitempty"`
			}{{
				Email: args.Email,
				Phone: args.Phone,
			}}
			customer.ContactsDetails = &contactDetails
		}
		if args.OrganizationNumber != nil {
			customer.OrgNumber = args.OrganizationNumber
		}
		if args.VatNumber != nil {
			customer.VatNumber = args.VatNumber
		}
		if args.PaymentTerms != nil {
			paymentTermsStr := fmt.Sprintf("%d", *args.PaymentTerms)
			customer.PaymentTerms = &paymentTermsStr
		}

		resp, err := client.CompanyClient.PostCustomer(ctx, companyUUID, customer)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to create customer: %v", err),
					},
				},
			}, nil, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, nil, nil
		}

		var responseData json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully created customer\n\nCompany: %s\nCustomer: %s\nStatus: %d\nResponse: %s", companyIDStr, args.Name, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_customers_get",
		Description: "Get a specific customer by ID",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CustomerGetParams) (*mcp.CallToolResult, any, error) {
		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, nil, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, nil, nil
		}

		if args.CustomerID == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer ID is required",
					},
				},
			}, nil, nil
		}

		customerUUID, err := uuid.Parse(args.CustomerID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid customer ID format: %v", err),
					},
				},
			}, nil, nil
		}

		resp, err := client.CompanyClient.GetCustomersCustomerId(ctx, companyUUID, customerUUID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to get customer: %v", err),
					},
				},
			}, nil, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer not found",
					},
				},
			}, nil, nil
		}

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, nil, nil
		}

		var responseData json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved customer\n\nCompany: %s\nCustomer ID: %s\nStatus: %d\nResponse: %s", companyIDStr, args.CustomerID, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_customers_update",
		Description: "Update an existing customer",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CustomerUpdateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Operation not allowed in read-only mode",
					},
				},
			}, nil, nil
		}

		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, nil, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, nil, nil
		}

		if args.CustomerID == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer ID is required",
					},
				},
			}, nil, nil
		}

		customerUUID, err := uuid.Parse(args.CustomerID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid customer ID format: %v", err),
					},
				},
			}, nil, nil
		}

		customer := company.Customer{}

		if args.Name != nil {
			customer.Name = *args.Name
		}
		if args.Email != nil || args.Phone != nil {
			contactDetails := []struct {
				Email     *string             `json:"email,omitempty"`
				Id        *openapi_types.UUID `json:"id,omitempty"`
				IsDefault *bool               `json:"isDefault,omitempty"`
				Name      *string             `json:"name,omitempty"`
				Phone     *string             `json:"phone,omitempty"`
			}{{
				Email: args.Email,
				Phone: args.Phone,
			}}
			customer.ContactsDetails = &contactDetails
		}
		if args.OrganizationNumber != nil {
			customer.OrgNumber = args.OrganizationNumber
		}
		if args.VatNumber != nil {
			customer.VatNumber = args.VatNumber
		}
		if args.Type != nil {
			customerType := company.CustomerType(*args.Type)
			if customerType != company.Company && customerType != company.Private {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Customer type must be 'company' or 'private'",
						},
					},
				}, nil, nil
			}
			customer.Type = customerType
		}
		if args.PaymentTerms != nil {
			paymentTermsStr := fmt.Sprintf("%d", *args.PaymentTerms)
			customer.PaymentTerms = &paymentTermsStr
		}

		resp, err := client.CompanyClient.PutCustomer(ctx, companyUUID, customerUUID, customer)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to update customer: %v", err),
					},
				},
			}, nil, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer not found",
					},
				},
			}, nil, nil
		}

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, nil, nil
		}

		var responseData json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully updated customer\n\nCompany: %s\nCustomer ID: %s\nStatus: %d\nResponse: %s", companyIDStr, args.CustomerID, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	return nil
}
