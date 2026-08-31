package tools

import (
	"context"

	"github.com/google/uuid"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/klowdo/bokio-mcp/bokio/generated/general"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// CompanyRefParams identifies a company
type CompanyRefParams struct {
	CompanyID string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
}

// AccountGetParams defines parameters for fetching a single account
type AccountGetParams struct {
	CompanyID string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Account   int32  `json:"account" jsonschema:"Account number from the chart of accounts (e.g. 1930)"`
}

// FiscalYearGetParams defines parameters for fetching a single fiscal year
type FiscalYearGetParams struct {
	CompanyID    string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	FiscalYearID string `json:"fiscal_year_id" jsonschema:"Fiscal year UUID"`
}

// JournalEntryCommentsListParams defines parameters for listing journal entry comments
type JournalEntryCommentsListParams struct {
	CompanyID      string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	JournalEntryID string `json:"journal_entry_id" jsonschema:"Journal entry UUID"`
}

// JournalEntryCommentRefParams identifies a single journal entry comment
type JournalEntryCommentRefParams struct {
	CompanyID      string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	JournalEntryID string `json:"journal_entry_id" jsonschema:"Journal entry UUID"`
	CommentID      string `json:"comment_id" jsonschema:"Comment UUID"`
}

// JournalEntryCommentCreateParams defines parameters for commenting on a journal entry
type JournalEntryCommentCreateParams struct {
	CompanyID      string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	JournalEntryID string `json:"journal_entry_id" jsonschema:"Journal entry UUID"`
	Content        string `json:"content" jsonschema:"Comment text"`
}

// JournalEntryCommentUpdateParams defines parameters for editing a journal entry comment
type JournalEntryCommentUpdateParams struct {
	CompanyID      string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	JournalEntryID string `json:"journal_entry_id" jsonschema:"Journal entry UUID"`
	CommentID      string `json:"comment_id" jsonschema:"Comment UUID to update"`
	Content        string `json:"content" jsonschema:"New comment text"`
}

// ConnectionsListParams defines parameters for listing integration connections
type ConnectionsListParams struct {
	Page     *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize *int32  `json:"page_size,omitempty" jsonschema:"Items per page, max 100 (optional)"`
	TenantID *string `json:"tenant_id,omitempty" jsonschema:"Filter by tenant UUID (optional)"`
}

// ConnectionRefParams identifies a single connection
type ConnectionRefParams struct {
	ConnectionID string `json:"connection_id" jsonschema:"Connection UUID"`
}

// RegisterCompanyInfoTools registers company information, account, fiscal year, comment and connection tools
func RegisterCompanyInfoTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_company_information_get",
		Description: "Get company information such as name, organisation number and VAT settings",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args CompanyRefParams) (*mcp.CallToolResult, any, error) {
		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetCompanyInformationWithId(ctx, companyUUID)

		return renderAPIResponse("get company information", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_accounts_get",
		Description: "Get a single account from the chart of accounts by its number",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args AccountGetParams) (*mcp.CallToolResult, any, error) {
		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetAccountWithNumber(ctx, companyUUID, args.Account)

		return renderAPIResponse("get account", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_fiscal_years_get",
		Description: "Get a single fiscal year by id",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args FiscalYearGetParams) (*mcp.CallToolResult, any, error) {
		companyUUID, fiscalYearUUID, errResult := resolveEntityRef(args.CompanyID, args.FiscalYearID, "fiscal year")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetFiscalYearWithId(ctx, companyUUID, fiscalYearUUID)

		return renderAPIResponse("get fiscal year", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_journal_entries_comments_list",
		Description: "List comments on a journal entry",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args JournalEntryCommentsListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, journalEntryUUID, errResult := resolveEntityRef(args.CompanyID, args.JournalEntryID, "journal entry")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetJournalEntryComments(ctx, companyUUID, journalEntryUUID)

		return renderAPIResponse("list journal entry comments", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_journal_entries_comments_get",
		Description: "Get a single comment on a journal entry",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args JournalEntryCommentRefParams) (*mcp.CallToolResult, any, error) {
		companyUUID, journalEntryUUID, commentUUID, errResult := resolveNestedRef(args.CompanyID, args.JournalEntryID, args.CommentID, "journal entry", "comment")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetJournalEntryComment(ctx, companyUUID, journalEntryUUID, commentUUID)

		return renderAPIResponse("get journal entry comment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_journal_entries_comments_create",
		Description: "Add a comment to a journal entry",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args JournalEntryCommentCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, journalEntryUUID, errResult := resolveEntityRef(args.CompanyID, args.JournalEntryID, "journal entry")
		if errResult != nil {
			return errResult, nil, nil
		}

		if args.Content == "" {
			return toolError("Comment content is required"), nil, nil
		}

		resp, err := client.CompanyClient.PostJournalEntryComment(ctx, companyUUID, journalEntryUUID, company.JournalEntryCommentWrite{
			Content: args.Content,
		})

		return renderAPIResponse("create journal entry comment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_journal_entries_comments_update",
		Description: "Edit a comment on a journal entry",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args JournalEntryCommentUpdateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, journalEntryUUID, commentUUID, errResult := resolveNestedRef(args.CompanyID, args.JournalEntryID, args.CommentID, "journal entry", "comment")
		if errResult != nil {
			return errResult, nil, nil
		}

		if args.Content == "" {
			return toolError("Comment content is required"), nil, nil
		}

		resp, err := client.CompanyClient.PutJournalEntryComment(ctx, companyUUID, journalEntryUUID, commentUUID, company.JournalEntryCommentWrite{
			Content: args.Content,
		})

		return renderAPIResponse("update journal entry comment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_journal_entries_comments_delete",
		Description: "Delete a comment from a journal entry",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args JournalEntryCommentRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, journalEntryUUID, commentUUID, errResult := resolveNestedRef(args.CompanyID, args.JournalEntryID, args.CommentID, "journal entry", "comment")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteJournalEntryComment(ctx, companyUUID, journalEntryUUID, commentUUID)

		return renderAPIResponse("delete journal entry comment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_connections_list",
		Description: "List the companies this integration token is connected to",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ConnectionsListParams) (*mcp.CallToolResult, any, error) {
		params := &general.GetConnectionsParams{
			Page:     args.Page,
			PageSize: args.PageSize,
		}

		if args.TenantID != nil {
			tenantUUID, err := uuid.Parse(*args.TenantID)
			if err != nil {
				return toolError("Invalid tenant ID format: %v", err), nil, nil
			}
			params.TenantId = &tenantUUID
		}

		resp, err := client.GeneralClient.GetConnections(ctx, params)

		return renderAPIResponse("list connections", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_connections_get",
		Description: "Get a single integration connection by id",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ConnectionRefParams) (*mcp.CallToolResult, any, error) {
		connectionUUID, err := uuid.Parse(args.ConnectionID)
		if err != nil {
			return toolError("Invalid connection ID format: %v", err), nil, nil
		}

		resp, err := client.GeneralClient.GetConnectionById(ctx, connectionUUID)

		return renderAPIResponse("get connection", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_connections_delete",
		Description: "Revoke an integration connection",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ConnectionRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		connectionUUID, err := uuid.Parse(args.ConnectionID)
		if err != nil {
			return toolError("Invalid connection ID format: %v", err), nil, nil
		}

		resp, err := client.GeneralClient.DeleteConnectionById(ctx, connectionUUID)

		return renderAPIResponse("delete connection", resp, err)
	})

	return nil
}
