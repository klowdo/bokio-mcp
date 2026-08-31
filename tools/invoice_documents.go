package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"mime/multipart"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// InvoiceRefParams identifies a single invoice
type InvoiceRefParams struct {
	CompanyID string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	InvoiceID string `json:"invoice_id" jsonschema:"Invoice UUID"`
}

// InvoiceAttachmentsListParams defines parameters for listing invoice attachments
type InvoiceAttachmentsListParams struct {
	CompanyID string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	InvoiceID string  `json:"invoice_id" jsonschema:"Invoice UUID"`
	Page      *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32  `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
	Query     *string `json:"query,omitempty" jsonschema:"Filter query (optional)"`
}

// InvoiceAttachmentRefParams identifies a single invoice attachment
type InvoiceAttachmentRefParams struct {
	CompanyID    string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	InvoiceID    string `json:"invoice_id" jsonschema:"Invoice UUID"`
	AttachmentID string `json:"attachment_id" jsonschema:"Attachment UUID"`
}

// InvoicePaymentRefParams identifies a single invoice payment
type InvoicePaymentRefParams struct {
	CompanyID string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	InvoiceID string `json:"invoice_id" jsonschema:"Invoice UUID"`
	PaymentID string `json:"payment_id" jsonschema:"Payment UUID"`
}

// InvoicePaymentCreateParams defines parameters for recording a payment on an invoice
type InvoicePaymentCreateParams struct {
	CompanyID                string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	InvoiceID                string  `json:"invoice_id" jsonschema:"Invoice UUID the payment belongs to"`
	Date                     string  `json:"date" jsonschema:"Payment date in YYYY-MM-DD format"`
	SumBaseCurrency          float64 `json:"sum_base_currency" jsonschema:"Paid amount in the company base currency"`
	BookkeepingAccountNumber int32   `json:"bookkeeping_account_number" jsonschema:"Account the payment was made to (e.g. 1930)"`
}

// InvoiceSettlementRefParams identifies a single invoice settlement
type InvoiceSettlementRefParams struct {
	CompanyID    string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	InvoiceID    string `json:"invoice_id" jsonschema:"Invoice UUID"`
	SettlementID string `json:"settlement_id" jsonschema:"Settlement UUID"`
}

// InvoiceSettlementCreateParams defines parameters for creating an invoice settlement
type InvoiceSettlementCreateParams struct {
	CompanyID string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	InvoiceID string `json:"invoice_id" jsonschema:"Invoice UUID the settlement belongs to"`
	Type      string `json:"type" jsonschema:"Settlement type: bankFee, currency or paymentServiceFee"`
	BodyJSON  string `json:"details_json" jsonschema:"Settlement details as a JSON object, matching the settlement type (e.g. {\"amount\":12.5,\"date\":\"2026-01-15\"})"`
}

// RegisterInvoiceDocumentTools registers invoice document, attachment, payment and settlement tools
func RegisterInvoiceDocumentTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_download",
		Description: "Download an invoice as a PDF, returned base64 encoded",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceRefParams) (*mcp.CallToolResult, any, error) {
		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DownloadInvoice(ctx, companyUUID, invoiceUUID)

		return renderBinaryResponse("download invoice", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_publish",
		Description: "Publish a draft invoice, making it final and bookkept",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.PostInvoicesInvoiceIdPublish(ctx, companyUUID, invoiceUUID)

		return renderAPIResponse("publish invoice", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_record",
		Description: "Record (bookkeep) an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.RecordInvoiceWithId(ctx, companyUUID, invoiceUUID)

		return renderAPIResponse("record invoice", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_delete",
		Description: "Delete an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteInvoiceWithId(ctx, companyUUID, invoiceUUID)

		return renderAPIResponse("delete invoice", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_credit",
		Description: "Create a credit note for an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.PostInvoicesInvoiceIdCredit(ctx, companyUUID, invoiceUUID)

		return renderAPIResponse("credit invoice", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_attachments_list",
		Description: "List attachments on an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceAttachmentsListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetInvoiceAttachments(ctx, companyUUID, invoiceUUID, &company.GetInvoiceAttachmentsParams{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Query,
		})

		return renderAPIResponse("list invoice attachments", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_attachments_get",
		Description: "Get metadata for a single invoice attachment",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceAttachmentRefParams) (*mcp.CallToolResult, any, error) {
		companyUUID, invoiceUUID, attachmentUUID, errResult := resolveNestedRef(args.CompanyID, args.InvoiceID, args.AttachmentID, "invoice", "attachment")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetInvoiceAttachment(ctx, companyUUID, invoiceUUID, attachmentUUID)

		return renderAPIResponse("get invoice attachment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_attachments_download",
		Description: "Download an invoice attachment, returned base64 encoded",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceAttachmentRefParams) (*mcp.CallToolResult, any, error) {
		companyUUID, invoiceUUID, attachmentUUID, errResult := resolveNestedRef(args.CompanyID, args.InvoiceID, args.AttachmentID, "invoice", "attachment")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DownloadInvoiceAttachment(ctx, companyUUID, invoiceUUID, attachmentUUID)

		return renderBinaryResponse("download invoice attachment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_attachments_delete",
		Description: "Delete an attachment from an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceAttachmentRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, attachmentUUID, errResult := resolveNestedRef(args.CompanyID, args.InvoiceID, args.AttachmentID, "invoice", "attachment")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteInvoiceAttachment(ctx, companyUUID, invoiceUUID, attachmentUUID)

		return renderAPIResponse("delete invoice attachment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_payments_get",
		Description: "Get a single payment recorded against an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoicePaymentRefParams) (*mcp.CallToolResult, any, error) {
		companyUUID, invoiceUUID, paymentUUID, errResult := resolveNestedRef(args.CompanyID, args.InvoiceID, args.PaymentID, "invoice", "payment")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetInvoicePayment(ctx, companyUUID, invoiceUUID, paymentUUID)

		return renderAPIResponse("get invoice payment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_payments_create",
		Description: "Record a payment against an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoicePaymentCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		date, err := parseAPIDate(args.Date)
		if err != nil {
			return toolError("Invalid date: %v", err), nil, nil
		}

		resp, err := client.CompanyClient.PostInvoicePayment(ctx, companyUUID, invoiceUUID, company.InvoicePayment{
			Date:                     date,
			SumBaseCurrency:          args.SumBaseCurrency,
			BookkeepingAccountNumber: args.BookkeepingAccountNumber,
		})

		return renderAPIResponse("create invoice payment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_payments_record",
		Description: "Record (bookkeep) a payment on an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoicePaymentRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, paymentUUID, errResult := resolveNestedRef(args.CompanyID, args.InvoiceID, args.PaymentID, "invoice", "payment")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.RecordInvoicePaymentWithId(ctx, companyUUID, invoiceUUID, paymentUUID)

		return renderAPIResponse("record invoice payment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_payments_delete",
		Description: "Delete a payment recorded against an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoicePaymentRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, paymentUUID, errResult := resolveNestedRef(args.CompanyID, args.InvoiceID, args.PaymentID, "invoice", "payment")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteInvoicePayment(ctx, companyUUID, invoiceUUID, paymentUUID)

		return renderAPIResponse("delete invoice payment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_settlements_get",
		Description: "Get a single settlement on an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceSettlementRefParams) (*mcp.CallToolResult, any, error) {
		companyUUID, invoiceUUID, settlementUUID, errResult := resolveNestedRef(args.CompanyID, args.InvoiceID, args.SettlementID, "invoice", "settlement")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetInvoiceSettlement(ctx, companyUUID, invoiceUUID, settlementUUID)

		return renderAPIResponse("get invoice settlement", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_settlements_create",
		Description: "Create a settlement on an invoice (currency difference, bank fee or payment service fee)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceSettlementCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		settlementType := company.InvoiceSettlementWriteType(args.Type)
		if !settlementType.Valid() {
			return toolError("Settlement type must be one of bankFee, currency or paymentServiceFee"), nil, nil
		}

		settlement := company.InvoiceSettlementWrite{Type: settlementType}
		if err := settlement.InvoiceSettlementDetails.UnmarshalJSON([]byte(args.BodyJSON)); err != nil {
			return toolError("Invalid details_json: %v", err), nil, nil
		}

		resp, err := client.CompanyClient.PostInvoiceSettlement(ctx, companyUUID, invoiceUUID, settlement)

		return renderAPIResponse("create invoice settlement", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_settlements_record",
		Description: "Record (bookkeep) a settlement on an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceSettlementRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, settlementUUID, errResult := resolveNestedRef(args.CompanyID, args.InvoiceID, args.SettlementID, "invoice", "settlement")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.RecordInvoiceSettlementWithId(ctx, companyUUID, invoiceUUID, settlementUUID)

		return renderAPIResponse("record invoice settlement", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_settlements_delete",
		Description: "Delete a settlement on an invoice",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceSettlementRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, settlementUUID, errResult := resolveNestedRef(args.CompanyID, args.InvoiceID, args.SettlementID, "invoice", "settlement")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteInvoiceSettlement(ctx, companyUUID, invoiceUUID, settlementUUID)

		return renderAPIResponse("delete invoice settlement", resp, err)
	})

	return nil
}

// InvoiceAttachmentCreateParams defines parameters for attaching a file to an invoice
type InvoiceAttachmentCreateParams struct {
	CompanyID   string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	InvoiceID   string `json:"invoice_id" jsonschema:"Invoice UUID to attach the file to"`
	FileName    string `json:"file_name" jsonschema:"Name of the file, including extension"`
	FileContent string `json:"file_content" jsonschema:"Base64 encoded file content (jpeg, png or pdf, max 4 MB)"`
}

// RegisterInvoiceAttachmentUploadTool registers the multipart invoice attachment upload tool
func RegisterInvoiceAttachmentUploadTool(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_invoices_attachments_create",
		Description: "Attach a file to an invoice. Attached files appear as extra pages on the final invoice.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args InvoiceAttachmentCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, invoiceUUID, errResult := resolveInvoiceRef(args.CompanyID, args.InvoiceID)
		if errResult != nil {
			return errResult, nil, nil
		}

		if args.FileName == "" {
			return toolError("file_name is required"), nil, nil
		}

		fileData, err := base64.StdEncoding.DecodeString(args.FileContent)
		if err != nil {
			return toolError("Invalid base64 file_content: %v", err), nil, nil
		}

		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		fileWriter, err := writer.CreateFormFile("file", args.FileName)
		if err != nil {
			return toolError("Failed to create multipart file field: %v", err), nil, nil
		}

		if _, err := fileWriter.Write(fileData); err != nil {
			return toolError("Failed to write file data: %v", err), nil, nil
		}

		if err := writer.Close(); err != nil {
			return toolError("Failed to close multipart writer: %v", err), nil, nil
		}

		contentType := writer.FormDataContentType()
		resp, err := client.CompanyClient.PostInvoiceAttachmentWithBody(ctx, companyUUID, invoiceUUID,
			&company.PostInvoiceAttachmentParams{ContentType: contentType}, contentType, &body)

		return renderAPIResponse("create invoice attachment", resp, err)
	})

	return nil
}
