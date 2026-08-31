package tools

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// FiscalYearsListParams defines parameters for listing fiscal years
type FiscalYearsListParams struct {
	CompanyID string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Page      *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32  `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
	Query     *string `json:"query,omitempty" jsonschema:"Filter query on startDate or endDate (optional)"`
}

// ChartOfAccountsListParams defines parameters for listing the chart of accounts
type ChartOfAccountsListParams struct {
	CompanyID string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Query     *string `json:"query,omitempty" jsonschema:"Filter query on account, name or accountType (optional)"`
}

// SieDownloadParams defines parameters for downloading a SIE file
type SieDownloadParams struct {
	CompanyID    string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	FiscalYearID string `json:"fiscal_year_id" jsonschema:"Fiscal year UUID (from bokio_fiscal_years_list)"`
}

// BankPaymentsListParams defines parameters for listing bank payments
type BankPaymentsListParams struct {
	CompanyID string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Page      *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32  `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
	Query     *string `json:"query,omitempty" jsonschema:"Filter query on amount or status (optional)"`
}

// BankPaymentGetParams defines parameters for fetching a single bank payment
type BankPaymentGetParams struct {
	CompanyID     string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	BankPaymentID string `json:"bank_payment_id" jsonschema:"Bank payment UUID"`
}

// JournalEntryGetParams defines parameters for fetching a single journal entry
type JournalEntryGetParams struct {
	CompanyID      string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	JournalEntryID string `json:"journal_entry_id" jsonschema:"Journal entry UUID"`
}

// JournalEntryItemParams defines a single debit or credit line of a journal entry
type JournalEntryItemParams struct {
	Account int32    `json:"account" jsonschema:"Account number from the chart of accounts (e.g. 1930)"`
	Debit   *float64 `json:"debit,omitempty" jsonschema:"Debit amount (set either debit or credit)"`
	Credit  *float64 `json:"credit,omitempty" jsonschema:"Credit amount (set either debit or credit)"`
}

// JournalEntryCreateParams defines parameters for creating a journal entry
type JournalEntryCreateParams struct {
	CompanyID string                   `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Date      string                   `json:"date" jsonschema:"Entry date in YYYY-MM-DD format"`
	Title     string                   `json:"title" jsonschema:"Title describing the journal entry"`
	Items     []JournalEntryItemParams `json:"items" jsonschema:"Line items; total debit must equal total credit"`
}

// JournalEntryReverseParams defines parameters for reversing a journal entry
type JournalEntryReverseParams struct {
	CompanyID      string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	JournalEntryID string `json:"journal_entry_id" jsonschema:"Journal entry UUID to reverse"`
}

func readOnlyError() *mcp.CallToolResult {
	return toolError("Operation not allowed in read-only mode")
}

func parseAPIDate(value string) (openapi_types.Date, error) {
	parsed, err := time.Parse(time.DateOnly, value)
	if err != nil {
		return openapi_types.Date{}, fmt.Errorf("date must be in YYYY-MM-DD format: %w", err)
	}
	return openapi_types.Date{Time: parsed}, nil
}

// RegisterLedgerTools registers fiscal year, SIE, chart of accounts, bank payment and journal entry tools
func RegisterLedgerTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_fiscal_years_list",
		Description: "List fiscal years for a company. The fiscal year id is required to download a SIE file.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args FiscalYearsListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetFiscalYears(ctx, companyUUID, &company.GetFiscalYearsParams{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Query,
		})

		return renderAPIResponse("list fiscal years", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_chart_of_accounts_list",
		Description: "List the chart of accounts for a company. Use it to pick account numbers when creating journal entries.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ChartOfAccountsListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetChartOfAccounts(ctx, companyUUID, &company.GetChartOfAccountsParams{
			Query: args.Query,
		})

		return renderAPIResponse("list chart of accounts", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_sie_download",
		Description: "Download the SIE file for a fiscal year, containing the full general ledger with account balances. Returned base64 encoded.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args SieDownloadParams) (*mcp.CallToolResult, any, error) {
		companyUUID, fiscalYearUUID, errResult := resolveEntityRef(args.CompanyID, args.FiscalYearID, "fiscal year")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DownloadSieFile(ctx, companyUUID, fiscalYearUUID)
		if err != nil {
			return toolError("Failed to download SIE file: %v", err), nil, nil
		}
		defer resp.Body.Close()

		fileContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return toolError("Failed to read SIE file content: %v", err), nil, nil
		}

		if resp.StatusCode != http.StatusOK {
			return toolError("API returned status %d: %s", resp.StatusCode, fileContent), nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully downloaded SIE file\n\nFiscal Year: %s\nContent-Type: %s\nFile Size: %d bytes\nStatus: %d\n\nBase64 Content: %s",
						fiscalYearUUID, resp.Header.Get("Content-Type"), len(fileContent), resp.StatusCode,
						base64.StdEncoding.EncodeToString(fileContent)),
				},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_bank_payments_list",
		Description: "List bank payments (bank transactions) for a company with optional pagination and filtering",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args BankPaymentsListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetBankPaymentsLimited(ctx, companyUUID, &company.GetBankPaymentsLimitedParams{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Query,
		})

		return renderAPIResponse("list bank payments", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_bank_payments_get",
		Description: "Get a single bank payment by id",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args BankPaymentGetParams) (*mcp.CallToolResult, any, error) {
		companyUUID, bankPaymentUUID, errResult := resolveEntityRef(args.CompanyID, args.BankPaymentID, "bank payment")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetBankPaymentsBankPaymentId(ctx, companyUUID, bankPaymentUUID)

		return renderAPIResponse("get bank payment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_journal_entries_get",
		Description: "Get a single journal entry by id, including its line items",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args JournalEntryGetParams) (*mcp.CallToolResult, any, error) {
		companyUUID, journalEntryUUID, errResult := resolveEntityRef(args.CompanyID, args.JournalEntryID, "journal entry")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetJournalentriesJournalId(ctx, companyUUID, journalEntryUUID)

		return renderAPIResponse("get journal entry", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_journal_entries_create",
		Description: "Create a journal entry. Total debit must equal total credit.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args JournalEntryCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		if len(args.Items) == 0 {
			return toolError("At least one line item is required"), nil, nil
		}

		date, err := parseAPIDate(args.Date)
		if err != nil {
			return toolError("Invalid date: %v", err), nil, nil
		}

		items := make([]company.JournalEntryItem, 0, len(args.Items))
		for _, item := range args.Items {
			if item.Debit == nil && item.Credit == nil {
				return toolError("Line item for account %d must set debit or credit", item.Account), nil, nil
			}
			items = append(items, company.JournalEntryItem{
				Account: &item.Account,
				Debit:   item.Debit,
				Credit:  item.Credit,
			})
		}

		entry := company.JournalEntry{
			Date:  &date,
			Items: items,
		}
		if args.Title != "" {
			entry.Title = &args.Title
		}

		resp, err := client.CompanyClient.PostJournalentry(ctx, companyUUID, entry)

		return renderAPIResponse("create journal entry", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_journal_entries_reverse",
		Description: "Reverse an existing journal entry, creating a counter entry that cancels it",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args JournalEntryReverseParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, journalEntryUUID, errResult := resolveEntityRef(args.CompanyID, args.JournalEntryID, "journal entry")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.ReverseJournalentry(ctx, companyUUID, journalEntryUUID)

		return renderAPIResponse("reverse journal entry", resp, err)
	})

	return nil
}
