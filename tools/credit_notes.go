package tools

import (
	"context"
	"encoding/json"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CreditNotesListParams defines parameters for listing credit notes
type CreditNotesListParams struct {
	CompanyID string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Page      *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32  `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
	Query     *string `json:"query,omitempty" jsonschema:"Filter query (optional)"`
}

// CreditNoteRefParams identifies a single credit note
type CreditNoteRefParams struct {
	CompanyID    string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	CreditNoteID string `json:"credit_note_id" jsonschema:"Credit note UUID"`
}

// CreditNoteUpdateParams defines parameters for updating a credit note
type CreditNoteUpdateParams struct {
	CompanyID    string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	CreditNoteID string `json:"credit_note_id" jsonschema:"Credit note UUID to update"`
	CreditNote   string `json:"credit_note_json" jsonschema:"Full credit note as a JSON object, in the shape returned by bokio_credit_notes_get"`
}

// RegisterCreditNoteTools registers credit note tools
func RegisterCreditNoteTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_credit_notes_list",
		Description: "List credit notes for a company with optional pagination and filtering",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreditNotesListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.ListCreditNotesV1(ctx, companyUUID, &company.ListCreditNotesV1Params{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Query,
		})

		return renderAPIResponse("list credit notes", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_credit_notes_get",
		Description: "Get a single credit note by id",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreditNoteRefParams) (*mcp.CallToolResult, any, error) {
		companyUUID, creditNoteUUID, errResult := resolveEntityRef(args.CompanyID, args.CreditNoteID, "credit note")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetCreditNoteV1(ctx, companyUUID, creditNoteUUID)

		return renderAPIResponse("get credit note", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_credit_notes_download",
		Description: "Download a credit note as a PDF, returned base64 encoded",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreditNoteRefParams) (*mcp.CallToolResult, any, error) {
		companyUUID, creditNoteUUID, errResult := resolveEntityRef(args.CompanyID, args.CreditNoteID, "credit note")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DownloadCreditNote(ctx, companyUUID, creditNoteUUID)

		return renderBinaryResponse("download credit note", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_credit_notes_update",
		Description: "Update a draft credit note. Fetch it with bokio_credit_notes_get first, then send the modified object.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreditNoteUpdateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, creditNoteUUID, errResult := resolveEntityRef(args.CompanyID, args.CreditNoteID, "credit note")
		if errResult != nil {
			return errResult, nil, nil
		}

		var creditNote company.CreditNote
		if err := json.Unmarshal([]byte(args.CreditNote), &creditNote); err != nil {
			return toolError("Invalid credit_note_json: %v", err), nil, nil
		}

		resp, err := client.CompanyClient.UpdateCreditNoteV1(ctx, companyUUID, creditNoteUUID, creditNote)

		return renderAPIResponse("update credit note", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_credit_notes_publish",
		Description: "Publish a draft credit note",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreditNoteRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, creditNoteUUID, errResult := resolveEntityRef(args.CompanyID, args.CreditNoteID, "credit note")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.PublishCreditNoteV1(ctx, companyUUID, creditNoteUUID)

		return renderAPIResponse("publish credit note", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_credit_notes_record",
		Description: "Record (bookkeep) a credit note",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreditNoteRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, creditNoteUUID, errResult := resolveEntityRef(args.CompanyID, args.CreditNoteID, "credit note")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.RecordCreditNoteWithId(ctx, companyUUID, creditNoteUUID)

		return renderAPIResponse("record credit note", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_credit_notes_delete",
		Description: "Delete a credit note",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CreditNoteRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, creditNoteUUID, errResult := resolveEntityRef(args.CompanyID, args.CreditNoteID, "credit note")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteCreditNoteWithId(ctx, companyUUID, creditNoteUUID)

		return renderAPIResponse("delete credit note", resp, err)
	})

	return nil
}
