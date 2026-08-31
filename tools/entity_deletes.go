package tools

import (
	"context"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CustomerRefParams identifies a single customer
type CustomerRefParams struct {
	CompanyID  string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	CustomerID string `json:"customer_id" jsonschema:"Customer UUID"`
}

// ItemRefParams identifies a single item
type ItemRefParams struct {
	CompanyID string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	ItemID    string `json:"item_id" jsonschema:"Item UUID"`
}

// SupplierUpdateParams defines parameters for updating a supplier
type SupplierUpdateParams struct {
	CompanyID  string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	SupplierID string  `json:"supplier_id" jsonschema:"Supplier UUID to update"`
	Name       string  `json:"name" jsonschema:"Supplier name"`
	OrgNumber  *string `json:"org_number,omitempty" jsonschema:"Organisation number (optional)"`
	VatNumber  *string `json:"vat_number,omitempty" jsonschema:"VAT number (optional)"`
	Currency   *string `json:"currency,omitempty" jsonschema:"ISO 4217 currency code (optional)"`
	AddressL1  *string `json:"address_line1,omitempty" jsonschema:"Address line 1 (optional)"`
	AddressL2  *string `json:"address_line2,omitempty" jsonschema:"Address line 2 (optional)"`
	PostalCode *string `json:"postal_code,omitempty" jsonschema:"Postal code (optional)"`
	City       *string `json:"city,omitempty" jsonschema:"City (optional)"`
	Country    *string `json:"country,omitempty" jsonschema:"ISO 3166-1 alpha-2 country code (optional)"`
}

// SupplierInvoiceUploadParams attaches an existing upload to a supplier invoice
type SupplierInvoiceUploadParams struct {
	CompanyID         string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	SupplierInvoiceID string `json:"supplier_invoice_id" jsonschema:"Supplier invoice UUID"`
	UploadID          string `json:"upload_id" jsonschema:"UUID of an existing upload to attach"`
}

// RegisterEntityDeleteTools registers customer, item and supplier deletion and update tools
func RegisterEntityDeleteTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_customers_delete",
		Description: "Delete a customer",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CustomerRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, customerUUID, errResult := resolveEntityRef(args.CompanyID, args.CustomerID, "customer")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteCustomer(ctx, companyUUID, customerUUID)

		return renderAPIResponse("delete customer", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_items_delete",
		Description: "Delete an item (article)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ItemRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, itemUUID, errResult := resolveEntityRef(args.CompanyID, args.ItemID, "item")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteItem(ctx, companyUUID, itemUUID)

		return renderAPIResponse("delete item", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_suppliers_update",
		Description: "Update an existing supplier",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SupplierUpdateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, supplierUUID, errResult := resolveEntityRef(args.CompanyID, args.SupplierID, "supplier")
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

		resp, err := client.CompanyClient.PutSupplier(ctx, companyUUID, supplierUUID, supplier)

		return renderAPIResponse("update supplier", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_suppliers_delete",
		Description: "Delete a supplier",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SupplierGetParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, supplierUUID, errResult := resolveEntityRef(args.CompanyID, args.SupplierID, "supplier")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteSupplier(ctx, companyUUID, supplierUUID)

		return renderAPIResponse("delete supplier", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_supplier_invoices_delete",
		Description: "Delete a supplier invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SupplierInvoiceGetParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, errResult := resolveEntityRef(args.CompanyID, args.SupplierInvoiceID, "supplier invoice")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteSupplierInvoice(ctx, companyUUID, invoiceUUID)

		return renderAPIResponse("delete supplier invoice", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_supplier_invoices_upload_attach",
		Description: "Attach an existing upload to a supplier invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SupplierInvoiceUploadParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, uploadUUID, errResult := resolveNestedRef(args.CompanyID, args.SupplierInvoiceID, args.UploadID, "supplier invoice", "upload")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.PostSupplierInvoiceUpload(ctx, companyUUID, invoiceUUID, company.UploadRef{
			Id: &uploadUUID,
		})

		return renderAPIResponse("attach upload to supplier invoice", resp, err)
	})

	return nil
}
