# Journal Integration Test Guide

This document explains how to use the journal integration test for the Bokio MCP server. The test demonstrates real API usage with OAuth2 authentication and comprehensive error handling.

## 📋 Prerequisites

1. **Valid Bokio OAuth2 Credentials**
   - Client ID and Client Secret from the Bokio Developer Portal
   - Registered redirect URI (default: `http://localhost:8080/callback`)

2. **Go 1.24+ Environment**
   - The test uses Go 1.24 features and testify suite framework

3. **Network Access**
   - The test connects to the real Bokio API at `https://api.bokio.se`

## 🚀 Quick Start

### 1. Set Required Environment Variables

```bash
export BOKIO_CLIENT_ID="your_client_id_here"
export BOKIO_CLIENT_SECRET="your_client_secret_here"
```

### 2. Run the Integration Test

```bash
# Run all integration tests
make test

# Or run specifically the journal integration test
go test -v -run TestJournalIntegrationSuite
```

## 🔧 Configuration Options

### Environment Variables

| Variable                       | Required | Default                          | Description                         |
| ------------------------------ | -------- | -------------------------------- | ----------------------------------- |
| `BOKIO_CLIENT_ID`              | Yes      | -                                | OAuth2 client ID from Bokio         |
| `BOKIO_CLIENT_SECRET`          | Yes      | -                                | OAuth2 client secret from Bokio     |
| `BOKIO_BASE_URL`               | No       | `https://api.bokio.se`           | Bokio API base URL                  |
| `BOKIO_REDIRECT_URL`           | No       | `http://localhost:8080/callback` | OAuth2 redirect URI                 |
| `BOKIO_READ_ONLY`              | No       | `false`                          | Enable read-only mode               |
| `SKIP_AUTH_TESTS`              | No       | `false`                          | Skip tests requiring authentication |
| `TEST_AUTH_CODE`               | No       | -                                | Authorization code from OAuth2 flow |
| `TEST_ACCOUNT_ID`              | No       | `1930`                           | Account ID for filtering tests      |
| `TEST_FROM_DATE`               | No       | `2024-01-01`                     | Start date for filtering tests      |
| `TEST_TO_DATE`                 | No       | `2024-12-31`                     | End date for filtering tests        |
| `EXPECTED_MIN_JOURNAL_ENTRIES` | No       | `0`                              | Minimum expected journal entries    |

### Example Configurations

#### Basic Usage

```bash
export BOKIO_CLIENT_ID="your_client_id"
export BOKIO_CLIENT_SECRET="your_client_secret"
go test -v -run TestJournalIntegrationSuite
```

#### Read-Only Mode Testing

```bash
export BOKIO_READ_ONLY=true
go test -v -run TestJournalIntegrationSuite
```

#### Skip Authentication (Client Setup Only)

```bash
export SKIP_AUTH_TESTS=true
go test -v -run TestJournalIntegrationSuite/Test_01_ClientConfiguration
```

## 🔐 OAuth2 Authentication Flow

The integration test demonstrates the complete OAuth2 flow:

### 1. Automatic Authorization URL Generation

The test generates an authorization URL and displays it in the output:

```
📋 MANUAL AUTHENTICATION REQUIRED:
   1. Open the following URL in your browser:
      https://api.bokio.se/authorize?client_id=...
   2. Login to Bokio and authorize the application
   3. Copy the 'code' parameter from the redirect URL
   4. Set the TEST_AUTH_CODE environment variable with that code
   5. Re-run the test
```

### 2. Manual Authorization Step

1. Copy the displayed URL
2. Open it in your browser
3. Login to Bokio and authorize the application
4. After authorization, you'll be redirected to your callback URL
5. Copy the `code` parameter from the redirect URL

### 3. Complete Token Exchange

```bash
export TEST_AUTH_CODE="authorization_code_from_step_2"
go test -v -run TestJournalIntegrationSuite/Test_02_AuthenticationFlow
```

## 📊 Test Coverage

The integration test covers eight comprehensive test scenarios:

### Test_01_ClientConfiguration

- ✅ Client initialization and setup
- ✅ Read-only mode configuration
- ✅ Initial authentication state

### Test_02_AuthenticationFlow

- ✅ Authorization URL generation
- ✅ OAuth2 token exchange
- ✅ Authentication state validation
- ✅ Tenant information retrieval

### Test_03_ListJournalEntries

- ✅ Basic journal entry listing
- ✅ Pagination support
- ✅ Date range filtering
- ✅ Account code filtering

### Test_04_GetSpecificJournalEntry

- ✅ Individual journal entry retrieval
- ✅ Entry detail validation
- ✅ Journal entry balance verification

### Test_05_GetAccountsChart

- ✅ Chart of accounts retrieval
- ✅ Account information validation
- ✅ Test account verification

### Test_06_ReadOnlyModeValidation

- ✅ Write operation blocking
- ✅ Read operation allowance
- ✅ Error message validation

### Test_07_ErrorHandling

- ✅ Non-existent entry handling
- ✅ Invalid date format handling
- ✅ Invalid account code handling

### Test_08_RateLimitingAndRetries

- ✅ Sequential request handling
- ✅ Rate limiting validation
- ✅ Performance measurement

## 🎯 Running Specific Tests

### Run Individual Test Cases

```bash
# Test only client configuration
go test -v -run TestJournalIntegrationSuite/Test_01_ClientConfiguration

# Test only authentication flow
go test -v -run TestJournalIntegrationSuite/Test_02_AuthenticationFlow

# Test only journal listing functionality
go test -v -run TestJournalIntegrationSuite/Test_03_ListJournalEntries
```

### Run Tests with Custom Date Range

```bash
export TEST_FROM_DATE="2024-06-01"
export TEST_TO_DATE="2024-06-30"
go test -v -run TestJournalIntegrationSuite/Test_03_ListJournalEntries
```

### Run Tests with Specific Account

```bash
export TEST_ACCOUNT_ID="1510"  # Swedish chart of accounts: "Customer receivables"
go test -v -run TestJournalIntegrationSuite/Test_03_ListJournalEntries
```

## 🔍 Test Output Examples

### Successful Authentication

```
=== RUN   TestJournalIntegrationSuite/Test_02_AuthenticationFlow
=== Testing OAuth2 Authentication Flow ===
✓ Generated authorization URL: https://api.bokio.se/authorize?client_id=...
🔄 Exchanging authorization code for access token...
✓ Successfully authenticated (TenantID: company-123, Type: company)
```

### Journal Entry Listing

```
=== RUN   TestJournalIntegrationSuite/Test_03_ListJournalEntries
=== Testing Journal Entry Listing ===
📝 Test 1: Basic journal entry listing
✓ Found 247 journal entries (Page 1 of 10, Total: 247)
📝 Test 2: Testing pagination
✓ Page 2 retrieved successfully with 25 entries
📝 Test 3: Testing date range filtering
✓ Date range filter (2024-01-01 to 2024-12-31) returned 247 entries
```

### Read-Only Mode Validation

```
=== RUN   TestJournalIntegrationSuite/Test_06_ReadOnlyModeValidation
=== Testing Read-Only Mode Validation ===
✓ POST request correctly blocked in read-only mode
✓ GET request works correctly in read-only mode
```

## 🛠️ Troubleshooting

### Common Issues and Solutions

#### 1. Authentication Failures

```
Error: Failed to exchange code for token: invalid_grant
```

**Solution:** The authorization code has expired or been used. Generate a new one.

#### 2. Missing Credentials

```
Skipping integration tests: BOKIO_CLIENT_ID and BOKIO_CLIENT_SECRET must be set
```

**Solution:** Set the required environment variables or use `SKIP_AUTH_TESTS=true` for client-only tests.

#### 3. Network Connectivity

```
Error: Failed to list journal entries: dial tcp: no such host
```

**Solution:** Check internet connection and verify `BOKIO_BASE_URL` is correct.

#### 4. Rate Limiting

```
Error: 429 Too Many Requests
```

**Solution:** The test includes built-in rate limiting. If you see this, wait a moment and retry.

### Debug Mode

Enable verbose logging by running tests with the `-v` flag:

```bash
go test -v -run TestJournalIntegrationSuite 2>&1 | tee test_output.log
```

## 📚 Integration with MCP Tools

This test demonstrates the same API calls that the MCP tools make:

- `bokio_list_journal_entries` → `GET /journal-entries`
- `bokio_get_journal_entry` → `GET /journal-entries/{id}`
- `bokio_get_accounts` → `GET /accounts`

The test validates that these endpoints work correctly with:

- OAuth2 authentication
- Proper error handling
- Read-only mode restrictions
- Rate limiting
- Data validation

## 🚦 CI/CD Integration

### GitHub Actions Example

```yaml
name: Integration Tests
on:
  workflow_dispatch:
    inputs:
      client_id:
        description: "Bokio Client ID"
        required: true
      client_secret:
        description: "Bokio Client Secret"
        required: true

jobs:
  integration-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: "1.24"
      - name: Run Integration Tests
        env:
          BOKIO_CLIENT_ID: ${{ github.event.inputs.client_id }}
          BOKIO_CLIENT_SECRET: ${{ github.event.inputs.client_secret }}
          BOKIO_READ_ONLY: true # Safe for CI/CD
          SKIP_AUTH_TESTS: true # Skip manual auth in CI
        run: |
          go test -v -run TestJournalIntegrationSuite
```

## 📈 Performance Benchmarks

The test includes basic performance measurement:

```bash
# Run with timing
go test -v -run TestJournalIntegrationSuite/Test_08_RateLimitingAndRetries
```

Expected output:

```
📝 Making 5 sequential requests to test rate limiting...
   Request 1: 245ms
   Request 2: 198ms
   Request 3: 201ms
   Request 4: 203ms
   Request 5: 199ms
✓ Completed 5 requests in 1.046s (avg: 209ms per request)
✓ Rate limiting appears to be working (took 1.046s, expected >800ms)
```

This validates that:

- API responses are reasonably fast (< 500ms typical)
- Rate limiting is working as configured
- Client retry logic functions correctly

---

## 📞 Support

If you encounter issues with the integration test:

1. Check the [project CLAUDE.md](./CLAUDE.md) for development guidelines
2. Review the [main README.md](./README.md) for general setup instructions
3. Ensure all environment variables are correctly set
4. Verify your Bokio OAuth2 application is properly configured

The integration test is designed to be comprehensive and educational, demonstrating best practices for testing real API integrations with proper authentication, error handling, and edge case coverage.
