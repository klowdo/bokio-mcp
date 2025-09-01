package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterJournalTools registers journal entry-related MCP tools
func RegisterJournalTools(server *mcp.Server, client *bokio.Client) error {
	// Register bokio_list_journal_entries tool
	if err := server.RegisterTool("bokio_list_journal_entries", mcp.Tool{
		Name: "bokio_list_journal_entries",
		Description: "List journal entries with optional filtering and pagination",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"page": map[string]interface{}{
					"type": "integer",
					"description": "Page number for pagination (default: 1)",
					"minimum": 1,
				},
				"per_page": map[string]interface{}{
					"type": "integer",
					"description": "Number of items per page (default: 25, max: 100)",
					"minimum": 1,
					"maximum": 100,
				},
				"from_date": map[string]interface{}{
					"type": "string",
					"format": "date",
					"description": "Filter entries from this date (YYYY-MM-DD)",
				},
				"to_date": map[string]interface{}{
					"type": "string",
					"format": "date",
					"description": "Filter entries to this date (YYYY-MM-DD)",
				},
				"account_code": map[string]interface{}{
					"type": "string",
					"description": "Filter by account code",
				},
			},
		},
		Handler: createListJournalEntriesHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_list_journal_entries tool: %w", err)
	}

	// Register bokio_create_journal_entry tool
	if err := server.RegisterTool("bokio_create_journal_entry", mcp.Tool{
		Name: "bokio_create_journal_entry",
		Description: "Create a new journal entry with debit and credit lines",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"date": map[string]interface{}{
					"type": "string",
					"format": "date",
					"description": "Journal entry date (YYYY-MM-DD)",
				},
				"description": map[string]interface{}{
					"type": "string",
					"description": "Journal entry description",
				},
				"reference": map[string]interface{}{
					"type": "string",
					"description": "Optional reference number",
				},
				"lines": map[string]interface{}{
					"type": "array",
					"description": "Journal entry lines (must balance)",
					"minItems": 2,
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"account_code": map[string]interface{}{
								"type": "string",
								"description": "Account code from chart of accounts",
							},
							"description": map[string]interface{}{
								"type": "string",
								"description": "Line description",
							},
							"debit": map[string]interface{}{
								"type": "object",
								"description": "Debit amount (exclusive with credit)",
								"properties": map[string]interface{}{
									"amount": map[string]interface{}{
										"type": "number",
										"minimum": 0,
									},
									"currency": map[string]interface{}{
										"type": "string",
										"default": "SEK",
									},
								},
								"required": []string{"amount"},
							},
							"credit": map[string]interface{}{
								"type": "object",
								"description": "Credit amount (exclusive with debit)",
								"properties": map[string]interface{}{
									"amount": map[string]interface{}{
										"type": "number",
										"minimum": 0,
									},
									"currency": map[string]interface{}{
										"type": "string",
										"default": "SEK",
									},
								},
								"required": []string{"amount"},
							},
						},
						"required": []string{"account_code"},
						"oneOf": []map[string]interface{}{
							{"required": []string{"debit"}},
							{"required": []string{"credit"}},
						},
					},
				},
			},
			"required": []string{"date", "description", "lines"},
		},
		Handler: createCreateJournalEntryHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_create_journal_entry tool: %w", err)
	}

	// Register bokio_reverse_journal_entry tool
	if err := server.RegisterTool("bokio_reverse_journal_entry", mcp.Tool{
		Name: "bokio_reverse_journal_entry",
		Description: "Create a reversing journal entry for an existing entry",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
					"description": "Original journal entry ID to reverse",
				},
				"date": map[string]interface{}{
					"type": "string",
					"format": "date",
					"description": "Date for the reversing entry (YYYY-MM-DD)",
				},
				"description": map[string]interface{}{
					"type": "string",
					"description": "Optional description for the reversing entry",
				},
			},
			"required": []string{"id", "date"},
		},
		Handler: createReverseJournalEntryHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_reverse_journal_entry tool: %w", err)
	}

	// Register bokio_get_accounts tool
	if err := server.RegisterTool("bokio_get_accounts", mcp.Tool{
		Name: "bokio_get_accounts",
		Description: "Get chart of accounts to see available account codes",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"type": map[string]interface{}{
					"type": "string",
					"description": "Filter by account type",
					"enum": []string{"asset", "liability", "equity", "revenue", "expense"},
				},
			},
		},
		Handler: createGetAccountsHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_get_accounts tool: %w", err)
	}

	return nil
}

// createListJournalEntriesHandler creates the handler for the list journal entries tool
func createListJournalEntriesHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		// Build query parameters
		queryParams := make(map[string]string)
		
		if page, ok := params["page"]; ok {
			queryParams["page"] = fmt.Sprintf("%v", page)
		}
		
		if perPage, ok := params["per_page"]; ok {
			queryParams["per_page"] = fmt.Sprintf("%v", perPage)
		}
		
		if fromDate, ok := params["from_date"].(string); ok && fromDate != "" {
			queryParams["from_date"] = fromDate
		}
		
		if toDate, ok := params["to_date"].(string); ok && toDate != "" {
			queryParams["to_date"] = toDate
		}
		
		if accountCode, ok := params["account_code"].(string); ok && accountCode != "" {
			queryParams["account_code"] = accountCode
		}

		// Construct URL with query parameters
		path := "/journal-entries"
		if len(queryParams) > 0 {
			path += "?"
			first := true
			for key, value := range queryParams {
				if !first {
					path += "&"
				}
				path += key + "=" + value
				first = false
			}
		}

		resp, err := client.Get(ctx, path)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to list journal entries: %v", err),
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var journalEntries bokio.ListResponse[bokio.JournalEntry]
		if err := json.Unmarshal(resp.Body(), &journalEntries); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": journalEntries.Data,
			"pagination": journalEntries.Meta,
		}, nil
	}
}

// createCreateJournalEntryHandler creates the handler for the create journal entry tool
func createCreateJournalEntryHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		// Parse and validate the request
		request, err := parseCreateJournalEntryRequest(params)
		if err != nil {
			return nil, fmt.Errorf("invalid request: %w", err)
		}

		// Validate that the entry balances
		if err := validateJournalEntryBalance(request); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Journal entry validation failed: %v", err),
			}, nil
		}

		resp, err := client.Post(ctx, "/journal-entries", request)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to create journal entry: %v", err),
			}, nil
		}

		if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var journalEntry bokio.JournalEntry
		if err := json.Unmarshal(resp.Body(), &journalEntry); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": journalEntry,
			"message": "Journal entry created successfully",
		}, nil
	}
}

// createReverseJournalEntryHandler creates the handler for the reverse journal entry tool
func createReverseJournalEntryHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		id, ok := params["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("journal entry ID is required")
		}

		date, ok := params["date"].(string)
		if !ok || date == "" {
			return nil, fmt.Errorf("reversal date is required")
		}

		reversalRequest := map[string]interface{}{
			"date": date,
		}

		if description, ok := params["description"].(string); ok && description != "" {
			reversalRequest["description"] = description
		}

		resp, err := client.Post(ctx, "/journal-entries/"+id+"/reverse", reversalRequest)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to reverse journal entry: %v", err),
			}, nil
		}

		if resp.StatusCode() == http.StatusNotFound {
			return map[string]interface{}{
				"success": false,
				"error": "Journal entry not found",
			}, nil
		}

		if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var journalEntry bokio.JournalEntry
		if err := json.Unmarshal(resp.Body(), &journalEntry); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": journalEntry,
			"message": "Journal entry reversed successfully",
		}, nil
	}
}

// createGetAccountsHandler creates the handler for the get accounts tool
func createGetAccountsHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		path := "/accounts"
		if accountType, ok := params["type"].(string); ok && accountType != "" {
			path += "?type=" + accountType
		}

		resp, err := client.Get(ctx, path)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to get accounts: %v", err),
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var accounts []bokio.Account
		if err := json.Unmarshal(resp.Body(), &accounts); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": accounts,
		}, nil
	}
}

// parseCreateJournalEntryRequest parses the parameters into a CreateJournalEntryRequest
func parseCreateJournalEntryRequest(params map[string]interface{}) (*bokio.CreateJournalEntryRequest, error) {
	date, ok := params["date"].(string)
	if !ok || date == "" {
		return nil, fmt.Errorf("date is required")
	}

	description, ok := params["description"].(string)
	if !ok || description == "" {
		return nil, fmt.Errorf("description is required")
	}

	linesRaw, ok := params["lines"].([]interface{})
	if !ok || len(linesRaw) < 2 {
		return nil, fmt.Errorf("at least 2 journal lines are required")
	}

	lines := make([]bokio.JournalEntryLine, len(linesRaw))
	for i, lineRaw := range linesRaw {
		lineMap, ok := lineRaw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid line at index %d", i)
		}

		accountCode, ok := lineMap["account_code"].(string)
		if !ok || accountCode == "" {
			return nil, fmt.Errorf("account_code is required for line %d", i)
		}

		line := bokio.JournalEntryLine{
			AccountCode: accountCode,
		}

		if lineDescription, ok := lineMap["description"].(string); ok {
			line.Description = lineDescription
		}

		// Check for debit or credit (exactly one should be provided)
		hasDebit := false
		hasCredit := false

		if debitRaw, ok := lineMap["debit"].(map[string]interface{}); ok {
			hasDebit = true
			amount, ok := debitRaw["amount"].(float64)
			if !ok {
				// Try parsing as int
				if amtInt, ok := debitRaw["amount"].(int); ok {
					amount = float64(amtInt)
				} else {
					return nil, fmt.Errorf("debit amount is required for line %d", i)
				}
			}

			currency, ok := debitRaw["currency"].(string)
			if !ok || currency == "" {
				currency = "SEK"
			}

			line.Debit = &bokio.Money{
				Amount:   amount,
				Currency: currency,
			}
		}

		if creditRaw, ok := lineMap["credit"].(map[string]interface{}); ok {
			hasCredit = true
			amount, ok := creditRaw["amount"].(float64)
			if !ok {
				// Try parsing as int
				if amtInt, ok := creditRaw["amount"].(int); ok {
					amount = float64(amtInt)
				} else {
					return nil, fmt.Errorf("credit amount is required for line %d", i)
				}
			}

			currency, ok := creditRaw["currency"].(string)
			if !ok || currency == "" {
				currency = "SEK"
			}

			line.Credit = &bokio.Money{
				Amount:   amount,
				Currency: currency,
			}
		}

		if !hasDebit && !hasCredit {
			return nil, fmt.Errorf("either debit or credit is required for line %d", i)
		}

		if hasDebit && hasCredit {
			return nil, fmt.Errorf("cannot have both debit and credit for line %d", i)
		}

		lines[i] = line
	}

	request := &bokio.CreateJournalEntryRequest{
		Description: description,
		Lines:       lines,
	}

	// Parse date (in a real implementation, convert string to time.Time)
	// For now, we'll leave Date as nil and let the API handle the string

	if reference, ok := params["reference"].(string); ok {
		request.Reference = reference
	}

	return request, nil
}

// validateJournalEntryBalance validates that debits equal credits
func validateJournalEntryBalance(request *bokio.CreateJournalEntryRequest) error {
	totalDebits := make(map[string]float64)
	totalCredits := make(map[string]float64)

	for _, line := range request.Lines {
		if line.Debit != nil {
			totalDebits[line.Debit.Currency] += line.Debit.Amount
		}
		if line.Credit != nil {
			totalCredits[line.Credit.Currency] += line.Credit.Amount
		}
	}

	// Check that debits equal credits for each currency
	for currency, debitTotal := range totalDebits {
		creditTotal, exists := totalCredits[currency]
		if !exists || debitTotal != creditTotal {
			return fmt.Errorf("journal entry does not balance for currency %s: debits=%.2f, credits=%.2f", currency, debitTotal, creditTotal)
		}
	}

	// Check that all currencies in credits are also in debits
	for currency, creditTotal := range totalCredits {
		debitTotal, exists := totalDebits[currency]
		if !exists || debitTotal != creditTotal {
			return fmt.Errorf("journal entry does not balance for currency %s: debits=%.2f, credits=%.2f", currency, debitTotal, creditTotal)
		}
	}

	return nil
}