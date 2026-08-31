package tools

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func readOnlyClientSession(t *testing.T) *mcp.ClientSession {
	t.Helper()

	client, err := bokio.NewAuthClient(&bokio.Config{
		IntegrationToken: "test-token",
		BaseURL:          "https://api.bokio.se/v1",
		ReadOnly:         true,
	})
	require.NoError(t, err)

	server := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	require.NoError(t, RegisterLedgerTools(server, client))
	require.NoError(t, RegisterSupplierTools(server, client))

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := server.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = serverSession.Close() })

	clientSession, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).
		Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = clientSession.Close() })

	return clientSession
}

func TestWriteToolsBlockedInReadOnlyMode(t *testing.T) {
	const companyID = "ea9ee4dd-fae3-4aec-a7db-6fc9cc1f8135"
	const entityID = "a419cf69-db6f-4de9-992c-b1a60942a443"

	tests := []struct {
		name string
		args map[string]any
	}{
		{
			name: "bokio_journal_entries_create",
			args: map[string]any{
				"company_id": companyID,
				"date":       "2026-01-15",
				"title":      "Office supplies",
				"items": []map[string]any{
					{"account": 5410, "debit": 500.0},
					{"account": 1930, "credit": 500.0},
				},
			},
		},
		{
			name: "bokio_journal_entries_reverse",
			args: map[string]any{"company_id": companyID, "journal_entry_id": entityID},
		},
		{
			name: "bokio_suppliers_create",
			args: map[string]any{"company_id": companyID, "name": "Acme AB"},
		},
		{
			name: "bokio_supplier_invoices_create",
			args: map[string]any{
				"company_id":   companyID,
				"supplier_id":  entityID,
				"invoice_date": "2026-01-15",
				"due_date":     "2026-02-15",
				"total_amount": 1250.0,
			},
		},
		{
			name: "bokio_supplier_invoices_update",
			args: map[string]any{
				"company_id":          companyID,
				"supplier_invoice_id": entityID,
				"supplier_id":         entityID,
				"invoice_date":        "2026-01-15",
				"due_date":            "2026-02-15",
				"total_amount":        1250.0,
			},
		},
	}

	session := readOnlyClientSession(t)
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tt.name, Arguments: tt.args})
			require.NoError(t, err)
			require.True(t, result.IsError, "%s must refuse to run in read-only mode", tt.name)

			text, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok)
			assert.Contains(t, strings.ToLower(text.Text), "read-only mode")
		})
	}
}

func TestParseAPIDate(t *testing.T) {
	date, err := parseAPIDate("2026-01-15")
	require.NoError(t, err)
	assert.Equal(t, "2026-01-15", date.Format("2006-01-02"))

	_, err = parseAPIDate("15/01/2026")
	assert.Error(t, err)
}

func TestSupplierInvoiceCreateParamsUnmarshal(t *testing.T) {
	input := `{"company_id":"c1","supplier_id":"s1","invoice_date":"2026-01-15","due_date":"2026-02-15","total_amount":1250.5,"invoice_number":"INV-1"}`

	var got SupplierInvoiceCreateParams
	require.NoError(t, json.Unmarshal([]byte(input), &got))

	assert.Equal(t, SupplierInvoiceCreateParams{
		CompanyID:     "c1",
		SupplierID:    "s1",
		InvoiceDate:   "2026-01-15",
		DueDate:       "2026-02-15",
		TotalAmount:   1250.5,
		InvoiceNumber: stringPtr("INV-1"),
	}, got)
}

func TestJournalEntryCreateParamsUnmarshal(t *testing.T) {
	input := `{"company_id":"c1","date":"2026-01-15","title":"Sale","items":[{"account":1930,"debit":100},{"account":3001,"credit":100}]}`

	var got JournalEntryCreateParams
	require.NoError(t, json.Unmarshal([]byte(input), &got))

	require.Len(t, got.Items, 2)
	assert.Equal(t, int32(1930), got.Items[0].Account)
	assert.Equal(t, 100.0, *got.Items[0].Debit)
	assert.Nil(t, got.Items[0].Credit)
	assert.Equal(t, 100.0, *got.Items[1].Credit)
}

func TestReadToolsReturnAPIContent(t *testing.T) {
	const companyID = "ea9ee4dd-fae3-4aec-a7db-6fc9cc1f8135"
	const entityID = "a419cf69-db6f-4de9-992c-b1a60942a443"

	var requestedPaths []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPaths = append(requestedPaths, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"marker":"real-api-payload"}`))
	}))
	defer server.Close()

	client, err := bokio.NewAuthClient(&bokio.Config{
		IntegrationToken: "test-token",
		BaseURL:          server.URL,
	})
	require.NoError(t, err)

	mcpServer := mcp.NewServer(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	require.NoError(t, RegisterLedgerTools(mcpServer, client))
	require.NoError(t, RegisterSupplierTools(mcpServer, client))

	ctx := context.Background()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	serverSession, err := mcpServer.Connect(ctx, serverTransport, nil)
	require.NoError(t, err)
	defer serverSession.Close()

	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil).
		Connect(ctx, clientTransport, nil)
	require.NoError(t, err)
	defer session.Close()

	tests := []struct {
		name string
		args map[string]any
	}{
		{"bokio_fiscal_years_list", map[string]any{"company_id": companyID}},
		{"bokio_chart_of_accounts_list", map[string]any{"company_id": companyID}},
		{"bokio_bank_payments_list", map[string]any{"company_id": companyID}},
		{"bokio_bank_payments_get", map[string]any{"company_id": companyID, "bank_payment_id": entityID}},
		{"bokio_journal_entries_get", map[string]any{"company_id": companyID, "journal_entry_id": entityID}},
		{"bokio_suppliers_list", map[string]any{"company_id": companyID}},
		{"bokio_suppliers_get", map[string]any{"company_id": companyID, "supplier_id": entityID}},
		{"bokio_supplier_invoices_list", map[string]any{"company_id": companyID}},
		{"bokio_supplier_invoices_get", map[string]any{"company_id": companyID, "supplier_invoice_id": entityID}},
		{"bokio_sie_download", map[string]any{"company_id": companyID, "fiscal_year_id": entityID}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := session.CallTool(ctx, &mcp.CallToolParams{Name: tt.name, Arguments: tt.args})
			require.NoError(t, err)
			require.False(t, result.IsError)

			text, ok := result.Content[0].(*mcp.TextContent)
			require.True(t, ok)

			want := "real-api-payload"
			if tt.name == "bokio_sie_download" {
				want = base64.StdEncoding.EncodeToString([]byte(`{"marker":"real-api-payload"}`))
			}
			assert.Contains(t, text.Text, want, "%s must return the API response, not an empty struct", tt.name)
		})
	}

	assert.Len(t, requestedPaths, len(tests))
}
