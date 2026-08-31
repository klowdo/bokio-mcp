package tools

import (
	"context"

	"github.com/google/uuid"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SuppliersListParams defines parameters for listing suppliers
type SuppliersListParams struct {
	CompanyID string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Page      *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32  `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
	Query     *string `json:"query,omitempty" jsonschema:"Filter query on name or vatNumber (optional)"`
}

// SupplierGetParams defines parameters for fetching a single supplier
type SupplierGetParams struct {
	CompanyID  string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	SupplierID string `json:"supplier_id" jsonschema:"Supplier UUID"`
}

// SupplierCreateParams defines parameters for creating a supplier
type SupplierCreateParams struct {
	CompanyID  string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Name       string  `json:"name" jsonschema:"Supplier name"`
	OrgNumber  *string `json:"org_number,omitempty" jsonschema:"Organisation number (optional)"`
	VatNumber  *string `json:"vat_number,omitempty" jsonschema:"VAT number (optional)"`
	Currency   *string `json:"currency,omitempty" jsonschema:"ISO 4217 currency code, defaults to SEK (optional)"`
	AddressL1  *string `json:"address_line1,omitempty" jsonschema:"Address line 1 (optional)"`
	AddressL2  *string `json:"address_line2,omitempty" jsonschema:"Address line 2 (optional)"`
	PostalCode *string `json:"postal_code,omitempty" jsonschema:"Postal code (optional)"`
	City       *string `json:"city,omitempty" jsonschema:"City (optional)"`
	Country    *string `json:"country,omitempty" jsonschema:"ISO 3166-1 alpha-2 country code, defaults to SE (optional)"`
}

// SupplierInvoicesListParams defines parameters for listing supplier invoices
type SupplierInvoicesListParams struct {
	CompanyID string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Page      *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32  `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
	Query     *string `json:"query,omitempty" jsonschema:"Filter query on supplierRef.id or invoiceNumber (optional)"`
}

// SupplierInvoiceGetParams defines parameters for fetching a single supplier invoice
type SupplierInvoiceGetParams struct {
	CompanyID         string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	SupplierInvoiceID string `json:"supplier_invoice_id" jsonschema:"Supplier invoice UUID"`
}

// SupplierInvoiceCreateParams defines parameters for creating a supplier invoice
type SupplierInvoiceCreateParams struct {
	CompanyID     string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	SupplierID    string  `json:"supplier_id" jsonschema:"Supplier UUID the invoice belongs to"`
	InvoiceDate   string  `json:"invoice_date" jsonschema:"Invoice date in YYYY-MM-DD format"`
	DueDate       string  `json:"due_date" jsonschema:"Due date in YYYY-MM-DD format"`
	TotalAmount   float64 `json:"total_amount" jsonschema:"Total invoice amount"`
	InvoiceNumber *string `json:"invoice_number,omitempty" jsonschema:"Invoice number (optional)"`
	UploadID      *string `json:"upload_id,omitempty" jsonschema:"UUID of an existing upload to attach (optional)"`
}

// SupplierInvoiceUpdateParams defines parameters for updating a supplier invoice
type SupplierInvoiceUpdateParams struct {
	CompanyID         string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	SupplierInvoiceID string  `json:"supplier_invoice_id" jsonschema:"Supplier invoice UUID to update"`
	SupplierID        string  `json:"supplier_id" jsonschema:"Supplier UUID the invoice belongs to"`
	InvoiceDate       string  `json:"invoice_date" jsonschema:"Invoice date in YYYY-MM-DD format"`
	DueDate           string  `json:"due_date" jsonschema:"Due date in YYYY-MM-DD format"`
	TotalAmount       float64 `json:"total_amount" jsonschema:"Total invoice amount"`
	InvoiceNumber     *string `json:"invoice_number,omitempty" jsonschema:"Invoice number (optional)"`
}

// RegisterSupplierTools registers supplier and supplier invoice tools
func RegisterSupplierTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_suppliers_list",
		Description: "List suppliers for a company with optional pagination and filtering",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SuppliersListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetSuppliers(ctx, companyUUID, &company.GetSuppliersParams{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Query,
		})

		return renderAPIResponse("list suppliers", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_suppliers_get",
		Description: "Get a single supplier by id",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SupplierGetParams) (*mcp.CallToolResult, any, error) {
		companyUUID, supplierUUID, errResult := resolveEntityRef(args.CompanyID, args.SupplierID, "supplier")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetSupplierById(ctx, companyUUID, supplierUUID)

		return renderAPIResponse("get supplier", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_suppliers_create",
		Description: "Create a new supplier for a company",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SupplierCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		if args.Name == "" {
			return toolError("Supplier name is required"), nil, nil
		}

		supplier := company.Supplier{
			Name:      args.Name,
			OrgNumber: args.OrgNumber,
			VatNumber: args.VatNumber,
			Currency:  args.Currency,
		}

		if args.AddressL1 != nil || args.AddressL2 != nil || args.PostalCode != nil || args.City != nil || args.Country != nil {
			supplier.Address = &company.SupplierAddress{
				Line1:      args.AddressL1,
				Line2:      args.AddressL2,
				PostalCode: args.PostalCode,
				City:       args.City,
				Country:    args.Country,
			}
		}

		resp, err := client.CompanyClient.PostSupplier(ctx, companyUUID, supplier)

		return renderAPIResponse("create supplier", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_supplier_invoices_list",
		Description: "List supplier invoices for a company with optional pagination and filtering",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SupplierInvoicesListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetSupplierInvoices(ctx, companyUUID, &company.GetSupplierInvoicesParams{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Query,
		})

		return renderAPIResponse("list supplier invoices", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_supplier_invoices_get",
		Description: "Get a single supplier invoice by id",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SupplierInvoiceGetParams) (*mcp.CallToolResult, any, error) {
		companyUUID, invoiceUUID, errResult := resolveEntityRef(args.CompanyID, args.SupplierInvoiceID, "supplier invoice")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetSupplierInvoiceById(ctx, companyUUID, invoiceUUID)

		return renderAPIResponse("get supplier invoice", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_supplier_invoices_create",
		Description: "Create a supplier invoice for an existing supplier",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SupplierInvoiceCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, supplierUUID, errResult := resolveEntityRef(args.CompanyID, args.SupplierID, "supplier")
		if errResult != nil {
			return errResult, nil, nil
		}

		invoiceDate, err := parseAPIDate(args.InvoiceDate)
		if err != nil {
			return toolError("Invalid invoice_date: %v", err), nil, nil
		}

		dueDate, err := parseAPIDate(args.DueDate)
		if err != nil {
			return toolError("Invalid due_date: %v", err), nil, nil
		}

		invoice := company.SupplierInvoiceCreate{
			InvoiceDate:   invoiceDate,
			DueDate:       dueDate,
			TotalAmount:   args.TotalAmount,
			InvoiceNumber: args.InvoiceNumber,
		}
		invoice.SupplierRef.Id = supplierUUID

		if args.UploadID != nil {
			uploadUUID, err := uuid.Parse(*args.UploadID)
			if err != nil {
				return toolError("Invalid upload ID format: %v", err), nil, nil
			}
			invoice.UploadRef = &company.UploadRef{Id: &uploadUUID}
		}

		resp, err := client.CompanyClient.PostSupplierInvoice(ctx, companyUUID, invoice)

		return renderAPIResponse("create supplier invoice", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_supplier_invoices_update",
		Description: "Update an existing supplier invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SupplierInvoiceUpdateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, errResult := resolveEntityRef(args.CompanyID, args.SupplierInvoiceID, "supplier invoice")
		if errResult != nil {
			return errResult, nil, nil
		}

		supplierUUID, err := uuid.Parse(args.SupplierID)
		if err != nil {
			return toolError("Invalid supplier ID format: %v", err), nil, nil
		}

		invoiceDate, err := parseAPIDate(args.InvoiceDate)
		if err != nil {
			return toolError("Invalid invoice_date: %v", err), nil, nil
		}

		dueDate, err := parseAPIDate(args.DueDate)
		if err != nil {
			return toolError("Invalid due_date: %v", err), nil, nil
		}

		invoice := company.SupplierInvoiceUpdate{
			InvoiceDate:   invoiceDate,
			DueDate:       dueDate,
			TotalAmount:   args.TotalAmount,
			InvoiceNumber: args.InvoiceNumber,
		}
		invoice.SupplierRef.Id = supplierUUID

		resp, err := client.CompanyClient.PutSupplierInvoice(ctx, companyUUID, invoiceUUID, invoice)

		return renderAPIResponse("update supplier invoice", resp, err)
	})

	return nil
}
