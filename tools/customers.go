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

// CustomersListResult defines the result for listing customers
type CustomersListResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
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

// CustomerCreateResult defines the result for creating a customer
type CustomerCreateResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// CustomerGetParams defines parameters for getting a customer
type CustomerGetParams struct {
	CompanyID  string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	CustomerID string `json:"customer_id" jsonschema:"Customer UUID"`
}

// CustomerGetResult defines the result for getting a customer
type CustomerGetResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
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

// CustomerUpdateResult defines the result for updating a customer
type CustomerUpdateResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// RegisterCustomerTools registers customer-related MCP tools using generated API clients
func RegisterCustomerTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_customers_list",
		Description: "List customers for a company with optional pagination and filtering",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CustomersListParams) (*mcp.CallToolResult, CustomersListResult, error) {
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
			}, CustomersListResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, CustomersListResult{}, nil
		}

		genParams := &company.GetCustomerParams{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Search,
		}

		resp, err := client.CompanyClient.GetCustomer(ctx, companyUUID, genParams)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to list customers: %v", err),
					},
				},
			}, CustomersListResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, CustomersListResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, CustomersListResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved customers\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
				},
			},
		}, CustomersListResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_customers_create",
		Description: "Create a new customer for a company",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CustomerCreateParams) (*mcp.CallToolResult, CustomerCreateResult, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Operation not allowed in read-only mode",
					},
				},
			}, CustomerCreateResult{}, nil
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
			}, CustomerCreateResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, CustomerCreateResult{}, nil
		}

		if args.Name == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer name is required",
					},
				},
			}, CustomerCreateResult{}, nil
		}

		customerType := company.CustomerType(args.Type)
		if customerType != company.Company && customerType != company.Private {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer type must be 'company' or 'private'",
					},
				},
			}, CustomerCreateResult{}, nil
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
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to create customer: %v", err),
					},
				},
			}, CustomerCreateResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, CustomerCreateResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, CustomerCreateResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully created customer\n\nCompany: %s\nCustomer: %s\nStatus: %d\nResponse: %v", companyIDStr, args.Name, resp.StatusCode, responseData),
				},
			},
		}, CustomerCreateResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_customers_get",
		Description: "Get a specific customer by ID",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CustomerGetParams) (*mcp.CallToolResult, CustomerGetResult, error) {
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
			}, CustomerGetResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, CustomerGetResult{}, nil
		}

		if args.CustomerID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer ID is required",
					},
				},
			}, CustomerGetResult{}, nil
		}

		customerUUID, err := uuid.Parse(args.CustomerID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid customer ID format: %v", err),
					},
				},
			}, CustomerGetResult{}, nil
		}

		resp, err := client.CompanyClient.GetCustomersCustomerId(ctx, companyUUID, customerUUID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to get customer: %v", err),
					},
				},
			}, CustomerGetResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer not found",
					},
				},
			}, CustomerGetResult{}, nil
		}

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, CustomerGetResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, CustomerGetResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved customer\n\nCompany: %s\nCustomer ID: %s\nStatus: %d\nResponse: %v", companyIDStr, args.CustomerID, resp.StatusCode, responseData),
				},
			},
		}, CustomerGetResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_customers_update",
		Description: "Update an existing customer",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CustomerUpdateParams) (*mcp.CallToolResult, CustomerUpdateResult, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Operation not allowed in read-only mode",
					},
				},
			}, CustomerUpdateResult{}, nil
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
			}, CustomerUpdateResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, CustomerUpdateResult{}, nil
		}

		if args.CustomerID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer ID is required",
					},
				},
			}, CustomerUpdateResult{}, nil
		}

		customerUUID, err := uuid.Parse(args.CustomerID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid customer ID format: %v", err),
					},
				},
			}, CustomerUpdateResult{}, nil
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
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Customer type must be 'company' or 'private'",
						},
					},
				}, CustomerUpdateResult{}, nil
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
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to update customer: %v", err),
					},
				},
			}, CustomerUpdateResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Customer not found",
					},
				},
			}, CustomerUpdateResult{}, nil
		}

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, CustomerUpdateResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, CustomerUpdateResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully updated customer\n\nCompany: %s\nCustomer ID: %s\nStatus: %d\nResponse: %v", companyIDStr, args.CustomerID, resp.StatusCode, responseData),
				},
			},
		}, CustomerUpdateResult{}, nil
	})

	return nil
}
