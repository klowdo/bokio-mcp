# Bokio MCP Examples

This directory contains example programs that demonstrate how to use the Bokio MCP server with pure generated API clients.

## 📋 Prerequisites

Set the required environment variables:

```bash
# Required: Your Bokio Integration Token
export BOKIO_INTEGRATION_TOKEN="your_integration_token_here"

# Required: The Company ID you want to query
export BOKIO_COMPANY_ID="your_company_uuid_here"

# Optional: API base URL (defaults to https://api.bokio.se)
export BOKIO_BASE_URL="https://api.bokio.se"
```

## 🚀 Examples

### List Journal Entries

**File**: `list_journal_entries.go`

Fetches and displays the latest 5 journal entries from your Bokio company using only the generated API clients.

**Run**:
```bash
# Build and run the example
nix develop -c go run examples/list_journal_entries.go

# Or build first, then run
nix develop -c go build -o bin/list-journal-entries examples/list_journal_entries.go
./bin/list-journal-entries
```

**What it does**:
- ✅ Uses ONLY generated `company.Client.GetJournalentry()` method
- ✅ Authenticates with Bearer token from environment
- ✅ Fetches latest 5 journal entries
- ✅ Pretty prints the full API response
- ✅ Shows summary of journal entry IDs, titles, and dates

**Example Output**:
```
🚀 Fetching latest 5 journal entries from Bokio API
📊 Company ID: 123e4567-e89b-12d3-a456-426614174000
🔗 API Base URL: https://api.bokio.se
🔐 Using Integration Token authentication

✅ SUCCESS! Retrieved journal entries
📋 Response Status: 200

📊 Latest 5 Journal Entries:
═══════════════════════════════════════
{
  "items": [
    {
      "id": "abc123",
      "title": "Sales Invoice #001",
      "date": "2025-01-15",
      "items": [...]
    }
  ],
  "totalItems": 25,
  "totalPages": 5,
  "currentPage": 1
}
═══════════════════════════════════════

📈 Summary: Found 5 journal entries
  1. ID: abc123 | Title: Sales Invoice #001 | Date: 2025-01-15
  2. ID: def456 | Title: Purchase #002 | Date: 2025-01-14
  ...

🎉 Example completed successfully!
```

## 🔧 Technical Details

- **Pure Generated Clients**: Uses only `bokio/generated/company` package
- **Type Safety**: All API calls use generated types from OpenAPI schema
- **Authentication**: Simple Bearer token authentication
- **Error Handling**: Comprehensive error messages and status checking
- **Response Parsing**: Handles the actual API response format

## 🏗️ Architecture

The examples demonstrate the clean architecture:

1. **Environment Config**: Load from `BOKIO_*` environment variables
2. **Auth Client**: Create authenticated wrapper around generated clients
3. **Generated API Call**: Use type-safe generated methods like `GetJournalentry()`
4. **Response Handling**: Parse and display actual API responses

This is exactly what the MCP server does internally - no manual API implementations, only generated clients from the real OpenAPI schema.
