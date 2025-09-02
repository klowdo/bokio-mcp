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
	CompanyID string  `json:"company_id"`
	Page      *int32  `json:"page,omitempty"`
	PageSize  *int32  `json:"page_size,omitempty"`
	Search    *string `json:"search,omitempty"`
}

// CustomersListResult defines the result for listing customers
type CustomersListResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CustomerCreateParams defines parameters for creating a customer
type CustomerCreateParams struct {
	CompanyID          string  `json:"company_id"`
	Name               string  `json:"name"`
	Email              *string `json:"email,omitempty"`
	Phone              *string `json:"phone,omitempty"`
	OrganizationNumber *string `json:"organization_number,omitempty"`
	VatNumber          *string `json:"vat_number,omitempty"`
	Type               string  `json:"type"` // "company" or "private"
	PaymentTerms       *int    `json:"payment_terms,omitempty"`
}

// CustomerCreateResult defines the result for creating a customer
type CustomerCreateResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CustomerGetParams defines parameters for getting a customer
type CustomerGetParams struct {
	CompanyID  string `json:"company_id"`
	CustomerID string `json:"customer_id"`
}

// CustomerGetResult defines the result for getting a customer
type CustomerGetResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// CustomerUpdateParams defines parameters for updating a customer
type CustomerUpdateParams struct {
	CompanyID          string  `json:"company_id"`
	CustomerID         string  `json:"customer_id"`
	Name               *string `json:"name,omitempty"`
	Email              *string `json:"email,omitempty"`
	Phone              *string `json:"phone,omitempty"`
	OrganizationNumber *string `json:"organization_number,omitempty"`
	VatNumber          *string `json:"vat_number,omitempty"`
	Type               *string `json:"type,omitempty"` // "company" or "private"
	PaymentTerms       *int    `json:"payment_terms,omitempty"`
}

// CustomerUpdateResult defines the result for updating a customer
type CustomerUpdateResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RegisterCustomerTools registers customer-related MCP tools using generated API clients
func RegisterCustomerTools(server *mcp.Server, client *bokio.AuthClient) error {
	// Tool to list customers using generated client
	listCustomersTool := mcp.NewServerTool[CustomersListParams, CustomersListResult](
		"bokio_customers_list",
		"List customers for a company with optional pagination and filtering",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[CustomersListParams]) (*mcp.CallToolResultFor[CustomersListResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[CustomersListResult]{
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
				return &mcp.CallToolResultFor[CustomersListResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Create parameters for the generated client
			genParams := &company.GetCustomerParams{
				Page:     params.Arguments.Page,
				PageSize: params.Arguments.PageSize,
				Query:    params.Arguments.Search,
			}

			// Call the generated client method
			resp, err := client.CompanyClient.GetCustomer(ctx, companyUUID, genParams)
			if err != nil {
				return &mcp.CallToolResultFor[CustomersListResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to list customers: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[CustomersListResult]{
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
				return &mcp.CallToolResultFor[CustomersListResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[CustomersListResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully retrieved customers\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
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
			mcp.Property("search",
				mcp.Description("Search customers by name or email (optional)"),
			),
		),
	)

	// Tool to create a customer using generated client
	createCustomerTool := mcp.NewServerTool[CustomerCreateParams, CustomerCreateResult](
		"bokio_customers_create",
		"Create a new customer for a company",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[CustomerCreateParams]) (*mcp.CallToolResultFor[CustomerCreateResult], error) {
			// Check read-only mode
			if client.GetConfig().ReadOnly {
				return &mcp.CallToolResultFor[CustomerCreateResult]{
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
				return &mcp.CallToolResultFor[CustomerCreateResult]{
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
				return &mcp.CallToolResultFor[CustomerCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Validate required fields
			if params.Arguments.Name == "" {
				return &mcp.CallToolResultFor[CustomerCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Customer name is required",
						},
					},
				}, nil
			}

			// Validate customer type
			customerType := company.CustomerType(params.Arguments.Type)
			if customerType != company.Company && customerType != company.Private {
				return &mcp.CallToolResultFor[CustomerCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Customer type must be 'company' or 'private'",
						},
					},
				}, nil
			}

			// Build customer object from parameters
			customer := company.Customer{
				Name: params.Arguments.Name,
				Type: customerType,
			}

			// Add optional fields
			if params.Arguments.Email != nil || params.Arguments.Phone != nil {
				contactDetails := []struct {
					Email     *string             `json:"email,omitempty"`
					Id        *openapi_types.UUID `json:"id"`
					IsDefault *bool               `json:"isDefault,omitempty"`
					Name      *string             `json:"name,omitempty"`
					Phone     *string             `json:"phone,omitempty"`
				}{{
					Email: params.Arguments.Email,
					Phone: params.Arguments.Phone,
				}}
				customer.ContactsDetails = &contactDetails
			}
			if params.Arguments.OrganizationNumber != nil {
				customer.OrgNumber = params.Arguments.OrganizationNumber
			}
			if params.Arguments.VatNumber != nil {
				customer.VatNumber = params.Arguments.VatNumber
			}
			if params.Arguments.PaymentTerms != nil {
				paymentTermsStr := fmt.Sprintf("%d", *params.Arguments.PaymentTerms)
				customer.PaymentTerms = &paymentTermsStr
			}

			// Call the generated client method
			resp, err := client.CompanyClient.PostCustomer(ctx, companyUUID, customer)
			if err != nil {
				return &mcp.CallToolResultFor[CustomerCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to create customer: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[CustomerCreateResult]{
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
				return &mcp.CallToolResultFor[CustomerCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[CustomerCreateResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully created customer\n\nCompany: %s\nCustomer: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.Name, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("name",
				mcp.Description("Customer name"),
				mcp.Required(true),
			),
			mcp.Property("type",
				mcp.Description("Customer type: 'company' or 'private'"),
				mcp.Required(true),
			),
			mcp.Property("email",
				mcp.Description("Customer email address (optional)"),
			),
			mcp.Property("phone",
				mcp.Description("Customer phone number (optional)"),
			),
			mcp.Property("organization_number",
				mcp.Description("Organization number (optional)"),
			),
			mcp.Property("vat_number",
				mcp.Description("VAT number (optional)"),
			),
			mcp.Property("payment_terms",
				mcp.Description("Payment terms in days (optional)"),
			),
		),
	)

	// Tool to get a specific customer using generated client
	getCustomerTool := mcp.NewServerTool[CustomerGetParams, CustomerGetResult](
		"bokio_customers_get",
		"Get a specific customer by ID",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[CustomerGetParams]) (*mcp.CallToolResultFor[CustomerGetResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[CustomerGetResult]{
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
				return &mcp.CallToolResultFor[CustomerGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Validate customer ID
			if params.Arguments.CustomerID == "" {
				return &mcp.CallToolResultFor[CustomerGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Customer ID is required",
						},
					},
				}, nil
			}

			// Parse customer UUID
			customerUUID, err := uuid.Parse(params.Arguments.CustomerID)
			if err != nil {
				return &mcp.CallToolResultFor[CustomerGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid customer ID format: %v", err),
						},
					},
				}, nil
			}

			// Call the generated client method
			resp, err := client.CompanyClient.GetCustomersCustomerId(ctx, companyUUID, customerUUID)
			if err != nil {
				return &mcp.CallToolResultFor[CustomerGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to get customer: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode == http.StatusNotFound {
				return &mcp.CallToolResultFor[CustomerGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Customer not found",
						},
					},
				}, nil
			}

			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[CustomerGetResult]{
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
				return &mcp.CallToolResultFor[CustomerGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[CustomerGetResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully retrieved customer\n\nCompany: %s\nCustomer ID: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.CustomerID, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("customer_id",
				mcp.Description("Customer UUID"),
				mcp.Required(true),
			),
		),
	)

	// Tool to update a customer using generated client
	updateCustomerTool := mcp.NewServerTool[CustomerUpdateParams, CustomerUpdateResult](
		"bokio_customers_update",
		"Update an existing customer",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[CustomerUpdateParams]) (*mcp.CallToolResultFor[CustomerUpdateResult], error) {
			// Check read-only mode
			if client.GetConfig().ReadOnly {
				return &mcp.CallToolResultFor[CustomerUpdateResult]{
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
				return &mcp.CallToolResultFor[CustomerUpdateResult]{
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
				return &mcp.CallToolResultFor[CustomerUpdateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Validate customer ID
			if params.Arguments.CustomerID == "" {
				return &mcp.CallToolResultFor[CustomerUpdateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Customer ID is required",
						},
					},
				}, nil
			}

			// Parse customer UUID
			customerUUID, err := uuid.Parse(params.Arguments.CustomerID)
			if err != nil {
				return &mcp.CallToolResultFor[CustomerUpdateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid customer ID format: %v", err),
						},
					},
				}, nil
			}

			// Build customer object from parameters (only include provided fields)
			customer := company.Customer{}

			// Add optional fields only if provided
			if params.Arguments.Name != nil {
				customer.Name = *params.Arguments.Name
			}
			if params.Arguments.Email != nil || params.Arguments.Phone != nil {
				contactDetails := []struct {
					Email     *string             `json:"email,omitempty"`
					Id        *openapi_types.UUID `json:"id"`
					IsDefault *bool               `json:"isDefault,omitempty"`
					Name      *string             `json:"name,omitempty"`
					Phone     *string             `json:"phone,omitempty"`
				}{{
					Email: params.Arguments.Email,
					Phone: params.Arguments.Phone,
				}}
				customer.ContactsDetails = &contactDetails
			}
			if params.Arguments.OrganizationNumber != nil {
				customer.OrgNumber = params.Arguments.OrganizationNumber
			}
			if params.Arguments.VatNumber != nil {
				customer.VatNumber = params.Arguments.VatNumber
			}
			if params.Arguments.Type != nil {
				customerType := company.CustomerType(*params.Arguments.Type)
				if customerType != company.Company && customerType != company.Private {
					return &mcp.CallToolResultFor[CustomerUpdateResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: "Customer type must be 'company' or 'private'",
							},
						},
					}, nil
				}
				customer.Type = customerType
			}
			if params.Arguments.PaymentTerms != nil {
				paymentTermsStr := fmt.Sprintf("%d", *params.Arguments.PaymentTerms)
				customer.PaymentTerms = &paymentTermsStr
			}

			// Call the generated client method
			resp, err := client.CompanyClient.PutCustomer(ctx, companyUUID, customerUUID, customer)
			if err != nil {
				return &mcp.CallToolResultFor[CustomerUpdateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to update customer: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode == http.StatusNotFound {
				return &mcp.CallToolResultFor[CustomerUpdateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Customer not found",
						},
					},
				}, nil
			}

			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[CustomerUpdateResult]{
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
				return &mcp.CallToolResultFor[CustomerUpdateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[CustomerUpdateResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully updated customer\n\nCompany: %s\nCustomer ID: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.CustomerID, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("customer_id",
				mcp.Description("Customer UUID"),
				mcp.Required(true),
			),
			mcp.Property("name",
				mcp.Description("Customer name (optional)"),
			),
			mcp.Property("type",
				mcp.Description("Customer type: 'company' or 'private' (optional)"),
			),
			mcp.Property("email",
				mcp.Description("Customer email address (optional)"),
			),
			mcp.Property("phone",
				mcp.Description("Customer phone number (optional)"),
			),
			mcp.Property("organization_number",
				mcp.Description("Organization number (optional)"),
			),
			mcp.Property("vat_number",
				mcp.Description("VAT number (optional)"),
			),
			mcp.Property("payment_terms",
				mcp.Description("Payment terms in days (optional)"),
			),
		),
	)

	// Register all tools
	server.AddTools(listCustomersTool, createCustomerTool, getCustomerTool, updateCustomerTool)

	return nil
}
