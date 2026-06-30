package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
)

func main() {
	// Load configuration from environment variables
	config := bokio.LoadConfigFromEnv()

	// Validate required environment variables
	if config.IntegrationToken == "" {
		log.Fatal("❌ BOKIO_INTEGRATION_TOKEN environment variable is required")
	}

	companyIDStr := os.Getenv("BOKIO_COMPANY_ID")
	if companyIDStr == "" {
		log.Fatal("❌ BOKIO_COMPANY_ID environment variable is required")
	}

	// Parse company UUID
	companyUUID, err := uuid.Parse(companyIDStr)
	if err != nil {
		log.Fatalf("❌ Invalid BOKIO_COMPANY_ID format: %v", err)
	}

	fmt.Printf("🚀 Fetching latest 5 journal entries from Bokio API\n")
	fmt.Printf("📊 Company ID: %s\n", companyIDStr)
	fmt.Printf("🔗 API Base URL: %s\n", config.BaseURL)
	fmt.Printf("🔐 Using Integration Token authentication\n\n")

	// Create authenticated client using ONLY generated clients
	client, err := bokio.NewAuthClient(config)
	if err != nil {
		log.Fatalf("❌ Failed to create auth client: %v", err)
	}

	// Set up parameters to get latest 5 journal entries
	pageSize := int32(5)
	params := &company.GetJournalentryParams{
		PageSize: &pageSize,
	}

	// Call the generated client method
	ctx := context.Background()
	resp, err := client.CompanyClient.GetJournalentry(ctx, companyUUID, params)
	if err != nil {
		log.Fatalf("❌ Failed to fetch journal entries: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check response status
	if resp.StatusCode != 200 {
		log.Fatalf("❌ API returned status %d", resp.StatusCode)
	}

	// Parse response
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		log.Fatalf("❌ Failed to decode response: %v", err)
	}

	fmt.Printf("✅ SUCCESS! Retrieved journal entries\n")
	fmt.Printf("📋 Response Status: %d\n\n", resp.StatusCode)

	// Pretty print the response
	prettyJSON, err := json.MarshalIndent(responseData, "", "  ")
	if err != nil {
		log.Fatalf("❌ Failed to format response: %v", err)
	}

	fmt.Printf("📊 Latest 5 Journal Entries:\n")
	fmt.Printf("═══════════════════════════════════════\n")
	fmt.Println(string(prettyJSON))
	fmt.Printf("═══════════════════════════════════════\n")

	// Extract and display summary if items are present
	if items, ok := responseData["items"].([]interface{}); ok && len(items) > 0 {
		fmt.Printf("\n📈 Summary: Found %d journal entries\n", len(items))

		for i, item := range items {
			if entry, ok := item.(map[string]interface{}); ok {
				id := "N/A"
				title := "N/A"
				date := "N/A"

				if idVal, exists := entry["id"]; exists {
					id = fmt.Sprintf("%v", idVal)
				}
				if titleVal, exists := entry["title"]; exists {
					title = fmt.Sprintf("%v", titleVal)
				}
				if dateVal, exists := entry["date"]; exists {
					date = fmt.Sprintf("%v", dateVal)
				}

				fmt.Printf("  %d. ID: %s | Title: %s | Date: %s\n", i+1, id, title, date)
			}
		}
	} else {
		fmt.Printf("\n📝 No journal entries found or unexpected response format\n")
	}

	fmt.Printf("\n🎉 Example completed successfully!\n")
}
