// Integration tests for Bokio MCP Server Journal API
// This file demonstrates how to test the journal functionality against the real Bokio API
// These tests require valid OAuth2 credentials and should be run manually with proper setup

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// JournalIntegrationTestSuite provides a comprehensive test suite for journal API operations
type JournalIntegrationTestSuite struct {
	suite.Suite
	client     *bokio.Client
	testCtx    context.Context
	testConfig *TestConfig
}

// TestConfig holds configuration for integration tests
type TestConfig struct {
	ClientID       string
	ClientSecret   string
	BaseURL        string
	RedirectURL    string
	ReadOnly       bool
	SkipAuth       bool // Skip tests requiring authentication if true
	TestAccountID  string
	TestDateRange  DateRange
	ExpectedMinEntries int // Minimum number of journal entries expected
}

// DateRange defines a date range for filtering tests
type DateRange struct {
	FromDate string // YYYY-MM-DD format
	ToDate   string // YYYY-MM-DD format
}

// SetupSuite runs once before all tests in the suite
func (suite *JournalIntegrationTestSuite) SetupSuite() {
	suite.testCtx = context.Background()
	
	// Load test configuration from environment variables
	suite.testConfig = &TestConfig{
		ClientID:     getEnvOrDefault("BOKIO_CLIENT_ID", ""),
		ClientSecret: getEnvOrDefault("BOKIO_CLIENT_SECRET", ""),
		BaseURL:      getEnvOrDefault("BOKIO_BASE_URL", "https://api.bokio.se"),
		RedirectURL:  getEnvOrDefault("BOKIO_REDIRECT_URL", "http://localhost:8080/callback"),
		ReadOnly:     getEnvOrDefault("BOKIO_READ_ONLY", "false") == "true",
		SkipAuth:     getEnvOrDefault("SKIP_AUTH_TESTS", "false") == "true",
		TestAccountID: getEnvOrDefault("TEST_ACCOUNT_ID", "1930"), // Common Swedish account for bank
		ExpectedMinEntries: getIntEnvOrDefault("EXPECTED_MIN_JOURNAL_ENTRIES", 0),
		TestDateRange: DateRange{
			FromDate: getEnvOrDefault("TEST_FROM_DATE", "2024-01-01"),
			ToDate:   getEnvOrDefault("TEST_TO_DATE", "2024-12-31"),
		},
	}

	// Validate required configuration
	if suite.testConfig.ClientID == "" || suite.testConfig.ClientSecret == "" {
		if !suite.testConfig.SkipAuth {
			suite.T().Skip("Skipping integration tests: BOKIO_CLIENT_ID and BOKIO_CLIENT_SECRET must be set. Set SKIP_AUTH_TESTS=true to run without auth.")
		}
		// Use dummy credentials for client setup tests when skipping auth
		suite.testConfig.ClientID = "dummy_client_id"
		suite.testConfig.ClientSecret = "dummy_client_secret"
	}

	// Initialize Bokio client
	config := &bokio.Config{
		ClientID:     suite.testConfig.ClientID,
		ClientSecret: suite.testConfig.ClientSecret,
		BaseURL:      suite.testConfig.BaseURL,
		RedirectURI:  suite.testConfig.RedirectURL,
		Scopes:       []string{"accounting", "invoices"},
		Timeout:      30 * time.Second,
		MaxRetries:   3,
		RateLimit:    5, // Lower rate limit for tests
		UserAgent:    "Bokio-MCP-IntegrationTest/1.0",
		ReadOnly:     suite.testConfig.ReadOnly,
		Logger:       &TestLogger{t: suite.T()},
	}

	client, err := bokio.NewClient(config)
	require.NoError(suite.T(), err, "Failed to create Bokio client")
	suite.client = client
}

// TestLogger implements bokio.Logger for test output
type TestLogger struct {
	t *testing.T
}

func (l *TestLogger) Debug(msg string, fields ...interface{}) {
	l.t.Logf("[DEBUG] %s %v", msg, fields)
}

func (l *TestLogger) Info(msg string, fields ...interface{}) {
	l.t.Logf("[INFO] %s %v", msg, fields)
}

func (l *TestLogger) Warn(msg string, fields ...interface{}) {
	l.t.Logf("[WARN] %s %v", msg, fields)
}

func (l *TestLogger) Error(msg string, fields ...interface{}) {
	l.t.Logf("[ERROR] %s %v", msg, fields)
}

// Test_01_ClientConfiguration tests basic client setup and configuration
func (suite *JournalIntegrationTestSuite) Test_01_ClientConfiguration() {
	suite.T().Log("=== Testing Client Configuration ===")
	
	// Test client is not nil
	assert.NotNil(suite.T(), suite.client, "Client should not be nil")
	
	// Test read-only mode setting
	assert.Equal(suite.T(), suite.testConfig.ReadOnly, suite.client.IsReadOnly(), "Read-only mode should match configuration")
	
	// Test client is not authenticated initially
	assert.False(suite.T(), suite.client.IsAuthenticated(), "Client should not be authenticated initially")
	
	suite.T().Logf("✓ Client configured correctly (ReadOnly: %v)", suite.client.IsReadOnly())
}

// Test_02_AuthenticationFlow demonstrates the OAuth2 authentication flow
func (suite *JournalIntegrationTestSuite) Test_02_AuthenticationFlow() {
	if suite.testConfig.SkipAuth {
		suite.T().Skip("Skipping authentication tests (SKIP_AUTH_TESTS=true)")
	}

	suite.T().Log("=== Testing OAuth2 Authentication Flow ===")
	
	// Step 1: Get authorization URL
	state := "test-state-" + strconv.FormatInt(time.Now().Unix(), 10)
	authURL := suite.client.GetAuthorizationURL(state)
	
	assert.NotEmpty(suite.T(), authURL, "Authorization URL should not be empty")
	assert.Contains(suite.T(), authURL, suite.testConfig.BaseURL, "Authorization URL should contain base URL")
	assert.Contains(suite.T(), authURL, "client_id=", "Authorization URL should contain client_id")
	assert.Contains(suite.T(), authURL, "state="+state, "Authorization URL should contain state parameter")
	
	suite.T().Logf("✓ Generated authorization URL: %s", authURL)
	
	// Step 2: Demonstrate manual authorization process
	suite.T().Log("\n📋 MANUAL AUTHENTICATION REQUIRED:")
	suite.T().Log("   1. Open the following URL in your browser:")
	suite.T().Logf("      %s", authURL)
	suite.T().Log("   2. Login to Bokio and authorize the application")
	suite.T().Log("   3. Copy the 'code' parameter from the redirect URL")
	suite.T().Log("   4. Set the TEST_AUTH_CODE environment variable with that code")
	suite.T().Log("   5. Re-run the test")
	
	// Check if we have an authorization code to complete the flow
	authCode := os.Getenv("TEST_AUTH_CODE")
	if authCode == "" {
		suite.T().Skip("Skipping token exchange: Set TEST_AUTH_CODE environment variable with authorization code from browser")
	}
	
	// Step 3: Exchange authorization code for tokens
	suite.T().Logf("🔄 Exchanging authorization code for access token...")
	err := suite.client.ExchangeCodeForToken(suite.testCtx, authCode)
	require.NoError(suite.T(), err, "Failed to exchange code for token")
	
	// Step 4: Verify authentication
	assert.True(suite.T(), suite.client.IsAuthenticated(), "Client should be authenticated after token exchange")
	
	// Step 5: Get tenant information
	tenantID, tenantType := suite.client.GetTenantInfo()
	assert.NotEmpty(suite.T(), tenantID, "Tenant ID should not be empty")
	assert.NotEmpty(suite.T(), tenantType, "Tenant type should not be empty")
	
	suite.T().Logf("✓ Successfully authenticated (TenantID: %s, Type: %s)", tenantID, tenantType)
}

// Test_03_ListJournalEntries tests the journal entry listing functionality
func (suite *JournalIntegrationTestSuite) Test_03_ListJournalEntries() {
	if !suite.client.IsAuthenticated() {
		suite.T().Skip("Skipping journal tests: Client not authenticated. Run authentication test first.")
	}

	suite.T().Log("=== Testing Journal Entry Listing ===")
	
	// Test 1: Basic listing without filters
	suite.T().Log("📝 Test 1: Basic journal entry listing")
	resp, err := suite.client.GET(suite.testCtx, "/journal-entries")
	require.NoError(suite.T(), err, "Failed to list journal entries")
	assert.Equal(suite.T(), 200, resp.StatusCode(), "Expected 200 OK status")
	
	var journalEntries bokio.JournalEntriesResponse
	err = json.Unmarshal(resp.Body(), &journalEntries)
	require.NoError(suite.T(), err, "Failed to parse journal entries response")
	
	suite.T().Logf("✓ Found %d journal entries (Page %d of %d, Total: %d)", 
		len(journalEntries.Items), 
		journalEntries.CurrentPage,
		journalEntries.TotalPages,
		journalEntries.TotalItems)
	
	if suite.testConfig.ExpectedMinEntries > 0 {
		assert.GreaterOrEqual(suite.T(), int(journalEntries.TotalItems), suite.testConfig.ExpectedMinEntries,
			"Total journal entries should meet minimum expected count")
	}
	
	// Test 2: Pagination
	if journalEntries.TotalPages > 1 {
		suite.T().Log("📝 Test 2: Testing pagination")
		resp, err := suite.client.GET(suite.testCtx, "/journal-entries?page=2&per_page=10")
		require.NoError(suite.T(), err, "Failed to get second page of journal entries")
		assert.Equal(suite.T(), 200, resp.StatusCode(), "Expected 200 OK status for page 2")
		
		var page2Entries bokio.JournalEntriesResponse
		err = json.Unmarshal(resp.Body(), &page2Entries)
		require.NoError(suite.T(), err, "Failed to parse page 2 response")
		
		assert.Equal(suite.T(), int32(2), page2Entries.CurrentPage, "Should be on page 2")
		suite.T().Logf("✓ Page 2 retrieved successfully with %d entries", len(page2Entries.Items))
	}
	
	// Test 3: Date range filtering
	suite.T().Log("📝 Test 3: Testing date range filtering")
	dateFilterURL := fmt.Sprintf("/journal-entries?from_date=%s&to_date=%s",
		suite.testConfig.TestDateRange.FromDate,
		suite.testConfig.TestDateRange.ToDate)
	
	resp, err = suite.client.GET(suite.testCtx, dateFilterURL)
	require.NoError(suite.T(), err, "Failed to list journal entries with date filter")
	assert.Equal(suite.T(), 200, resp.StatusCode(), "Expected 200 OK status for date filtered request")
	
	var filteredEntries bokio.JournalEntriesResponse
	err = json.Unmarshal(resp.Body(), &filteredEntries)
	require.NoError(suite.T(), err, "Failed to parse filtered journal entries response")
	
	suite.T().Logf("✓ Date range filter (%s to %s) returned %d entries", 
		suite.testConfig.TestDateRange.FromDate,
		suite.testConfig.TestDateRange.ToDate,
		len(filteredEntries.Items))
	
	// Test 4: Account code filtering (if we have entries)
	if len(journalEntries.Items) > 0 && suite.testConfig.TestAccountID != "" {
		suite.T().Log("📝 Test 4: Testing account code filtering")
		accountFilterURL := fmt.Sprintf("/journal-entries?account_code=%s", suite.testConfig.TestAccountID)
		
		resp, err := suite.client.GET(suite.testCtx, accountFilterURL)
		require.NoError(suite.T(), err, "Failed to list journal entries with account filter")
		assert.Equal(suite.T(), 200, resp.StatusCode(), "Expected 200 OK status for account filtered request")
		
		var accountFilteredEntries bokio.JournalEntriesResponse
		err = json.Unmarshal(resp.Body(), &accountFilteredEntries)
		require.NoError(suite.T(), err, "Failed to parse account filtered journal entries response")
		
		suite.T().Logf("✓ Account filter (account %s) returned %d entries", 
			suite.testConfig.TestAccountID,
			len(accountFilteredEntries.Items))
	}
}

// Test_04_GetSpecificJournalEntry tests retrieving specific journal entry details
func (suite *JournalIntegrationTestSuite) Test_04_GetSpecificJournalEntry() {
	if !suite.client.IsAuthenticated() {
		suite.T().Skip("Skipping journal entry detail tests: Client not authenticated")
	}

	suite.T().Log("=== Testing Specific Journal Entry Retrieval ===")
	
	// First, get a list of journal entries to find one to test with
	resp, err := suite.client.GET(suite.testCtx, "/journal-entries?per_page=5")
	require.NoError(suite.T(), err, "Failed to list journal entries for detail test")
	
	var journalEntries bokio.JournalEntriesResponse
	err = json.Unmarshal(resp.Body(), &journalEntries)
	require.NoError(suite.T(), err, "Failed to parse journal entries response")
	
	if len(journalEntries.Items) == 0 {
		suite.T().Skip("No journal entries available for testing specific entry retrieval")
	}
	
	// Test getting details of the first journal entry
	testEntry := journalEntries.Items[0]
	suite.T().Logf("📝 Testing details for journal entry: %s (ID: %s)", testEntry.Title, testEntry.ID)
	
	detailURL := fmt.Sprintf("/journal-entries/%s", testEntry.ID)
	resp, err = suite.client.GET(suite.testCtx, detailURL)
	require.NoError(suite.T(), err, "Failed to get journal entry details")
	assert.Equal(suite.T(), 200, resp.StatusCode(), "Expected 200 OK status for journal entry details")
	
	var detailedEntry bokio.JournalEntry
	err = json.Unmarshal(resp.Body(), &detailedEntry)
	require.NoError(suite.T(), err, "Failed to parse journal entry details response")
	
	// Validate the detailed entry
	assert.Equal(suite.T(), testEntry.ID, detailedEntry.ID, "Entry ID should match")
	assert.Equal(suite.T(), testEntry.Title, detailedEntry.Title, "Entry title should match")
	assert.NotEmpty(suite.T(), detailedEntry.Date, "Entry date should not be empty")
	assert.Greater(suite.T(), len(detailedEntry.Items), 0, "Entry should have at least one item")
	
	suite.T().Logf("✓ Retrieved journal entry details:")
	suite.T().Logf("   ID: %s", detailedEntry.ID)
	suite.T().Logf("   Title: %s", detailedEntry.Title)
	suite.T().Logf("   Date: %s", detailedEntry.Date)
	suite.T().Logf("   Number: %s", detailedEntry.JournalEntryNumber)
	suite.T().Logf("   Items: %d", len(detailedEntry.Items))
	
	// Validate journal entry items
	totalDebit := 0.0
	totalCredit := 0.0
	for i, item := range detailedEntry.Items {
		totalDebit += item.Debit
		totalCredit += item.Credit
		suite.T().Logf("   Item %d: Account %d, Debit %.2f, Credit %.2f", 
			i+1, item.Account, item.Debit, item.Credit)
	}
	
	// Journal entries should balance (debits equal credits)
	assert.InDelta(suite.T(), totalDebit, totalCredit, 0.01, "Journal entry should balance (debits = credits)")
	suite.T().Logf("✓ Journal entry balances correctly (Debits: %.2f, Credits: %.2f)", totalDebit, totalCredit)
}

// Test_05_GetAccountsChart tests retrieving the chart of accounts
func (suite *JournalIntegrationTestSuite) Test_05_GetAccountsChart() {
	if !suite.client.IsAuthenticated() {
		suite.T().Skip("Skipping accounts chart test: Client not authenticated")
	}

	suite.T().Log("=== Testing Chart of Accounts Retrieval ===")
	
	// Test getting all accounts
	resp, err := suite.client.GET(suite.testCtx, "/accounts")
	require.NoError(suite.T(), err, "Failed to get accounts")
	assert.Equal(suite.T(), 200, resp.StatusCode(), "Expected 200 OK status for accounts request")
	
	var accounts []bokio.Account
	err = json.Unmarshal(resp.Body(), &accounts)
	require.NoError(suite.T(), err, "Failed to parse accounts response")
	
	assert.Greater(suite.T(), len(accounts), 0, "Should have at least some accounts")
	
	suite.T().Logf("✓ Retrieved %d accounts from chart of accounts", len(accounts))
	
	// Display first few accounts for reference
	displayCount := 5
	if len(accounts) < displayCount {
		displayCount = len(accounts)
	}
	
	suite.T().Log("📊 Sample accounts:")
	for i := 0; i < displayCount; i++ {
		account := accounts[i]
		suite.T().Logf("   %d: %s (%s) - Active: %v", 
			account.Number, account.Name, account.Type, account.Active)
	}
	
	// Verify we have the test account if specified
	if suite.testConfig.TestAccountID != "" {
		testAccountNum, err := strconv.Atoi(suite.testConfig.TestAccountID)
		if err == nil {
			found := false
			for _, account := range accounts {
				if int(account.Number) == testAccountNum {
					found = true
					suite.T().Logf("✓ Found test account %d: %s", account.Number, account.Name)
					break
				}
			}
			if !found {
				suite.T().Logf("⚠ Test account %s not found in chart of accounts", suite.testConfig.TestAccountID)
			}
		}
	}
}

// Test_06_ReadOnlyModeValidation tests that write operations are blocked in read-only mode
func (suite *JournalIntegrationTestSuite) Test_06_ReadOnlyModeValidation() {
	if !suite.testConfig.ReadOnly {
		suite.T().Skip("Skipping read-only tests: Client not in read-only mode")
	}

	if !suite.client.IsAuthenticated() {
		suite.T().Skip("Skipping read-only validation: Client not authenticated")
	}

	suite.T().Log("=== Testing Read-Only Mode Validation ===")
	
	// Test that POST requests are blocked
	testJournalEntry := bokio.CreateJournalEntryRequest{
		Title: "Test Entry - Should Fail",
		Date:  "2024-01-01",
		Items: []bokio.JournalEntryItem{
			{Account: 1930, Debit: 1000.0, Credit: 0.0},
			{Account: 3000, Debit: 0.0, Credit: 1000.0},
		},
	}
	
	resp, err := suite.client.POST(suite.testCtx, "/journal-entries", testJournalEntry)
	
	// In read-only mode, we should get an error before making the HTTP request
	assert.Error(suite.T(), err, "POST request should fail in read-only mode")
	assert.Contains(suite.T(), err.Error(), "read-only mode", "Error should mention read-only mode")
	
	suite.T().Log("✓ POST request correctly blocked in read-only mode")
	
	// Test that GET requests still work
	resp, err = suite.client.GET(suite.testCtx, "/journal-entries?per_page=1")
	require.NoError(suite.T(), err, "GET request should work in read-only mode")
	assert.Equal(suite.T(), 200, resp.StatusCode(), "GET request should succeed in read-only mode")
	
	suite.T().Log("✓ GET request works correctly in read-only mode")
}

// Test_07_ErrorHandling tests various error scenarios
func (suite *JournalIntegrationTestSuite) Test_07_ErrorHandling() {
	if !suite.client.IsAuthenticated() {
		suite.T().Skip("Skipping error handling tests: Client not authenticated")
	}

	suite.T().Log("=== Testing Error Handling ===")
	
	// Test 1: Non-existent journal entry
	suite.T().Log("📝 Test 1: Non-existent journal entry")
	resp, err := suite.client.GET(suite.testCtx, "/journal-entries/non-existent-id")
	
	// This should either return an error or a 404 status
	if err != nil {
		suite.T().Logf("✓ Non-existent entry properly returned error: %v", err)
	} else {
		assert.Equal(suite.T(), 404, resp.StatusCode(), "Non-existent entry should return 404")
		suite.T().Logf("✓ Non-existent entry returned 404 status")
	}
	
	// Test 2: Invalid date format
	suite.T().Log("📝 Test 2: Invalid date format in filter")
	resp, err = suite.client.GET(suite.testCtx, "/journal-entries?from_date=invalid-date")
	
	if err != nil {
		suite.T().Logf("✓ Invalid date format properly returned error: %v", err)
	} else {
		// Some APIs might accept invalid dates and return empty results
		// or return a 400 status code
		if resp.StatusCode() >= 400 {
			suite.T().Logf("✓ Invalid date format returned error status: %d", resp.StatusCode())
		} else {
			suite.T().Logf("ℹ Invalid date format was accepted by API (status %d)", resp.StatusCode())
		}
	}
	
	// Test 3: Invalid account code
	suite.T().Log("📝 Test 3: Invalid account code filter")
	resp, err = suite.client.GET(suite.testCtx, "/journal-entries?account_code=99999")
	require.NoError(suite.T(), err, "Request with invalid account should not fail")
	
	// This should return empty results or all results (depending on API behavior)
	var entries bokio.JournalEntriesResponse
	err = json.Unmarshal(resp.Body(), &entries)
	require.NoError(suite.T(), err, "Should be able to parse response even with invalid account")
	
	suite.T().Logf("✓ Invalid account code handled gracefully (%d entries returned)", len(entries.Items))
}

// Test_08_RateLimitingAndRetries tests the client's rate limiting and retry logic
func (suite *JournalIntegrationTestSuite) Test_08_RateLimitingAndRetries() {
	if !suite.client.IsAuthenticated() {
		suite.T().Skip("Skipping rate limiting tests: Client not authenticated")
	}

	suite.T().Log("=== Testing Rate Limiting and Performance ===")
	
	startTime := time.Now()
	requestCount := 5
	
	suite.T().Logf("📝 Making %d sequential requests to test rate limiting...", requestCount)
	
	for i := 0; i < requestCount; i++ {
		requestStart := time.Now()
		resp, err := suite.client.GET(suite.testCtx, "/journal-entries?per_page=1")
		requestDuration := time.Since(requestStart)
		
		require.NoError(suite.T(), err, "Request %d should succeed", i+1)
		assert.Equal(suite.T(), 200, resp.StatusCode(), "Request %d should return 200", i+1)
		
		suite.T().Logf("   Request %d: %v", i+1, requestDuration)
	}
	
	totalDuration := time.Since(startTime)
	avgDuration := totalDuration / time.Duration(requestCount)
	
	suite.T().Logf("✓ Completed %d requests in %v (avg: %v per request)", 
		requestCount, totalDuration, avgDuration)
	
	// Rate limiting should ensure we don't exceed configured limits
	// With a 5 req/sec rate limit, 5 requests should take at least 800ms
	minExpectedDuration := 800 * time.Millisecond
	if totalDuration >= minExpectedDuration {
		suite.T().Logf("✓ Rate limiting appears to be working (took %v, expected >%v)", 
			totalDuration, minExpectedDuration)
	} else {
		suite.T().Logf("ℹ Requests completed quickly (%v) - rate limiting may not be active", totalDuration)
	}
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// TestJournalIntegrationSuite runs the integration test suite
func TestJournalIntegrationSuite(t *testing.T) {
	// Add a banner to make it clear these are integration tests
	log.Println("🚀 Starting Bokio MCP Server Journal Integration Tests")
	log.Println("================================================")
	log.Println("These tests require:")
	log.Println("  - Valid BOKIO_CLIENT_ID and BOKIO_CLIENT_SECRET")
	log.Println("  - Internet connection to Bokio API")
	log.Println("  - Manual authorization step for OAuth2 flow")
	log.Println("================================================")
	
	suite.Run(t, new(JournalIntegrationTestSuite))
}

// Example of how to run specific tests:
//
// Run all integration tests:
//   go test -v -run TestJournalIntegrationSuite
//
// Run only authentication tests:
//   go test -v -run TestJournalIntegrationSuite/Test_02_AuthenticationFlow
//
// Run with environment variables:
//   BOKIO_CLIENT_ID=your_id BOKIO_CLIENT_SECRET=your_secret go test -v -run TestJournalIntegrationSuite
//
// Run in read-only mode:
//   BOKIO_READ_ONLY=true go test -v -run TestJournalIntegrationSuite
//
// Skip authentication (for testing client setup only):
//   SKIP_AUTH_TESTS=true go test -v -run TestJournalIntegrationSuite