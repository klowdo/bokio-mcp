package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ItemListParams defines parameters for listing items
type ItemListParams struct {
	CompanyID string  `json:"company_id"`
	Page      *int32  `json:"page,omitempty"`
	PageSize  *int32  `json:"page_size,omitempty"`
	Query     *string `json:"query,omitempty"`
}

// ItemCreateParams defines parameters for creating an item
type ItemCreateParams struct {
	CompanyID   string   `json:"company_id"`
	ItemType    string   `json:"item_type"` // "salesItem" or "descriptionOnlyItem"
	Description string   `json:"description"`
	UnitPrice   *float64 `json:"unit_price,omitempty"`   // required for salesItem
	TaxRate     *float64 `json:"tax_rate,omitempty"`     // required for salesItem
	ProductType *string  `json:"product_type,omitempty"` // "goods" or "services" for salesItem
	UnitType    *string  `json:"unit_type,omitempty"`    // for salesItem
}

// ItemGetParams defines parameters for getting a specific item
type ItemGetParams struct {
	CompanyID string `json:"company_id"`
	ItemID    string `json:"item_id"`
}

// ItemUpdateParams defines parameters for updating an item
type ItemUpdateParams struct {
	CompanyID   string   `json:"company_id"`
	ItemID      string   `json:"item_id"`
	ItemType    string   `json:"item_type"` // "salesItem" or "descriptionOnlyItem"
	Description string   `json:"description"`
	UnitPrice   *float64 `json:"unit_price,omitempty"`   // required for salesItem
	TaxRate     *float64 `json:"tax_rate,omitempty"`     // required for salesItem
	ProductType *string  `json:"product_type,omitempty"` // "goods" or "services" for salesItem
	UnitType    *string  `json:"unit_type,omitempty"`    // for salesItem
}

// ItemResult defines the result structure for item operations
type ItemResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// RegisterItemTools registers item management tools using ONLY generated API clients
func RegisterItemTools(server *mcp.Server, client *bokio.AuthClient) error {
	// Tool to list items
	listItemsTool := mcp.NewServerTool[ItemListParams, ItemResult](
		"bokio_items_list",
		"List inventory items for a company with optional pagination and filtering",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ItemListParams]) (*mcp.CallToolResultFor[ItemResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
						},
					},
				}, nil
			}

			// Parse company UUID
			companyUUID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Create parameters for the generated client
			genParams := &company.GetItemsParams{
				Page:     params.Arguments.Page,
				PageSize: params.Arguments.PageSize,
				Query:    params.Arguments.Query,
			}

			// Call the generated client method
			resp, err := client.CompanyClient.GetItems(ctx, companyUUID, genParams)
			if err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to list items: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
						},
					},
				}, nil
			}

			// Parse response body as generic interface
			var responseData interface{}
			if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[ItemResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully retrieved items\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("page",
				mcp.Description("Page number (optional)"),
			),
			mcp.Property("page_size",
				mcp.Description("Items per page (optional)"),
			),
			mcp.Property("query",
				mcp.Description("Optional query to filter items (optional)"),
			),
		),
	)

	// Tool to create a new item
	createItemTool := mcp.NewServerTool[ItemCreateParams, ItemResult](
		"bokio_items_create",
		"Create a new inventory item (salesItem or descriptionOnlyItem)",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ItemCreateParams]) (*mcp.CallToolResultFor[ItemResult], error) {
			// Check for read-only mode
			if client.GetConfig().ReadOnly {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Create operation not allowed in read-only mode",
						},
					},
				}, nil
			}

			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
						},
					},
				}, nil
			}

			// Parse company UUID
			companyUUID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Validate required parameters
			if params.Arguments.ItemType == "" {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "item_type is required (salesItem or descriptionOnlyItem)",
						},
					},
				}, nil
			}

			if params.Arguments.Description == "" {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "description is required",
						},
					},
				}, nil
			}

			// Create the request body based on item type
			var requestBody company.PostItemJSONRequestBody

			if params.Arguments.ItemType == "salesItem" {
				// Validate required fields for salesItem
				if params.Arguments.UnitPrice == nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: "unit_price is required for salesItem",
							},
						},
					}, nil
				}
				if params.Arguments.TaxRate == nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: "tax_rate is required for salesItem",
							},
						},
					}, nil
				}

				// Default values
				productType := "goods"
				if params.Arguments.ProductType != nil {
					productType = *params.Arguments.ProductType
				}
				unitType := "piece"
				if params.Arguments.UnitType != nil {
					unitType = *params.Arguments.UnitType
				}

				salesItem := company.SalesItem{
					Description: params.Arguments.Description,
					ItemType:    company.SalesItemItemTypeSalesItem,
					ProductType: company.SalesItemProductType(productType),
					TaxRate:     *params.Arguments.TaxRate,
					UnitPrice:   *params.Arguments.UnitPrice,
					UnitType:    company.SalesItemUnitType(unitType),
				}

				// Marshal to JSON to create the union type
				salesItemJSON, err := json.Marshal(salesItem)
				if err != nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Failed to marshal salesItem: %v", err),
							},
						},
					}, nil
				}

				// Unmarshal into the request body type to handle union type properly
				if err := json.Unmarshal(salesItemJSON, &requestBody); err != nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Failed to create request body: %v", err),
							},
						},
					}, nil
				}

			} else if params.Arguments.ItemType == "descriptionOnlyItem" {
				descItem := company.DescriptionOnlyItem{
					Description: params.Arguments.Description,
					ItemType:    company.DescriptionOnlyItemItemTypeDescriptionOnlyItem,
				}

				// Marshal to JSON to create the union type
				descItemJSON, err := json.Marshal(descItem)
				if err != nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Failed to marshal descriptionOnlyItem: %v", err),
							},
						},
					}, nil
				}

				// Unmarshal into the request body type to handle union type properly
				if err := json.Unmarshal(descItemJSON, &requestBody); err != nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Failed to create request body: %v", err),
							},
						},
					}, nil
				}

			} else {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "item_type must be either 'salesItem' or 'descriptionOnlyItem'",
						},
					},
				}, nil
			}

			// Call the generated client method
			resp, err := client.CompanyClient.PostItem(ctx, companyUUID, requestBody)
			if err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to create item: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
						},
					},
				}, nil
			}

			// Parse response body as generic interface
			var responseData interface{}
			if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[ItemResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully created item\n\nCompany: %s\nItem Type: %s\nDescription: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.ItemType, params.Arguments.Description, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("item_type",
				mcp.Description("Type of item: 'salesItem' or 'descriptionOnlyItem'"),
				mcp.Required(true),
			),
			mcp.Property("description",
				mcp.Description("Item description"),
				mcp.Required(true),
			),
			mcp.Property("unit_price",
				mcp.Description("Unit price (required for salesItem)"),
			),
			mcp.Property("tax_rate",
				mcp.Description("Tax rate as decimal (e.g., 0.25 for 25%, required for salesItem)"),
			),
			mcp.Property("product_type",
				mcp.Description("Product type: 'goods' or 'services' (for salesItem, defaults to 'goods')"),
			),
			mcp.Property("unit_type",
				mcp.Description("Unit type: 'piece', 'hour', 'meter', etc. (for salesItem, defaults to 'piece')"),
			),
		),
	)

	// Tool to get a specific item by ID
	getItemTool := mcp.NewServerTool[ItemGetParams, ItemResult](
		"bokio_items_get",
		"Get a specific inventory item by ID",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ItemGetParams]) (*mcp.CallToolResultFor[ItemResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
						},
					},
				}, nil
			}

			if params.Arguments.ItemID == "" {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "item_id is required",
						},
					},
				}, nil
			}

			// Parse company UUID
			companyUUID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Parse item UUID
			itemUUID, err := uuid.Parse(params.Arguments.ItemID)
			if err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid item ID format: %v", err),
						},
					},
				}, nil
			}

			// Call the generated client method
			resp, err := client.CompanyClient.GetItemsItemId(ctx, companyUUID, itemUUID)
			if err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to get item: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
						},
					},
				}, nil
			}

			// Parse response body as generic interface
			var responseData interface{}
			if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[ItemResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully retrieved item\n\nCompany: %s\nItem ID: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.ItemID, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("item_id",
				mcp.Description("Item UUID"),
				mcp.Required(true),
			),
		),
	)

	// Tool to update an item
	updateItemTool := mcp.NewServerTool[ItemUpdateParams, ItemResult](
		"bokio_items_update",
		"Update an existing inventory item",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ItemUpdateParams]) (*mcp.CallToolResultFor[ItemResult], error) {
			// Check for read-only mode
			if client.GetConfig().ReadOnly {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Update operation not allowed in read-only mode",
						},
					},
				}, nil
			}

			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
						},
					},
				}, nil
			}

			if params.Arguments.ItemID == "" {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "item_id is required",
						},
					},
				}, nil
			}

			// Parse company UUID
			companyUUID, err := uuid.Parse(companyIDStr)
			if err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Parse item UUID
			itemUUID, err := uuid.Parse(params.Arguments.ItemID)
			if err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid item ID format: %v", err),
						},
					},
				}, nil
			}

			// Validate required parameters
			if params.Arguments.ItemType == "" {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "item_type is required (salesItem or descriptionOnlyItem)",
						},
					},
				}, nil
			}

			if params.Arguments.Description == "" {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "description is required",
						},
					},
				}, nil
			}

			// Create the request body based on item type
			var requestBody company.PutItemJSONRequestBody

			if params.Arguments.ItemType == "salesItem" {
				// Validate required fields for salesItem
				if params.Arguments.UnitPrice == nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: "unit_price is required for salesItem",
							},
						},
					}, nil
				}
				if params.Arguments.TaxRate == nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: "tax_rate is required for salesItem",
							},
						},
					}, nil
				}

				// Default values
				productType := "goods"
				if params.Arguments.ProductType != nil {
					productType = *params.Arguments.ProductType
				}
				unitType := "piece"
				if params.Arguments.UnitType != nil {
					unitType = *params.Arguments.UnitType
				}

				salesItem := company.SalesItem{
					Description: params.Arguments.Description,
					Id:          &itemUUID,
					ItemType:    company.SalesItemItemTypeSalesItem,
					ProductType: company.SalesItemProductType(productType),
					TaxRate:     *params.Arguments.TaxRate,
					UnitPrice:   *params.Arguments.UnitPrice,
					UnitType:    company.SalesItemUnitType(unitType),
				}

				// Marshal to JSON to create the union type
				salesItemJSON, err := json.Marshal(salesItem)
				if err != nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Failed to marshal salesItem: %v", err),
							},
						},
					}, nil
				}

				// Unmarshal into the request body type to handle union type properly
				if err := json.Unmarshal(salesItemJSON, &requestBody); err != nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Failed to create request body: %v", err),
							},
						},
					}, nil
				}

			} else if params.Arguments.ItemType == "descriptionOnlyItem" {
				descItem := company.DescriptionOnlyItem{
					Description: params.Arguments.Description,
					Id:          &itemUUID,
					ItemType:    company.DescriptionOnlyItemItemTypeDescriptionOnlyItem,
				}

				// Marshal to JSON to create the union type
				descItemJSON, err := json.Marshal(descItem)
				if err != nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Failed to marshal descriptionOnlyItem: %v", err),
							},
						},
					}, nil
				}

				// Unmarshal into the request body type to handle union type properly
				if err := json.Unmarshal(descItemJSON, &requestBody); err != nil {
					return &mcp.CallToolResultFor[ItemResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Failed to create request body: %v", err),
							},
						},
					}, nil
				}

			} else {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "item_type must be either 'salesItem' or 'descriptionOnlyItem'",
						},
					},
				}, nil
			}

			// Call the generated client method
			resp, err := client.CompanyClient.PutItem(ctx, companyUUID, itemUUID, requestBody)
			if err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to update item: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
						},
					},
				}, nil
			}

			// Parse response body as generic interface
			var responseData interface{}
			if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
				return &mcp.CallToolResultFor[ItemResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[ItemResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully updated item\n\nCompany: %s\nItem ID: %s\nItem Type: %s\nDescription: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.ItemID, params.Arguments.ItemType, params.Arguments.Description, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("item_id",
				mcp.Description("Item UUID"),
				mcp.Required(true),
			),
			mcp.Property("item_type",
				mcp.Description("Type of item: 'salesItem' or 'descriptionOnlyItem'"),
				mcp.Required(true),
			),
			mcp.Property("description",
				mcp.Description("Item description"),
				mcp.Required(true),
			),
			mcp.Property("unit_price",
				mcp.Description("Unit price (required for salesItem)"),
			),
			mcp.Property("tax_rate",
				mcp.Description("Tax rate as decimal (e.g., 0.25 for 25%, required for salesItem)"),
			),
			mcp.Property("product_type",
				mcp.Description("Product type: 'goods' or 'services' (for salesItem, defaults to 'goods')"),
			),
			mcp.Property("unit_type",
				mcp.Description("Unit type: 'piece', 'hour', 'meter', etc. (for salesItem, defaults to 'piece')"),
			),
		),
	)

	server.AddTools(listItemsTool, createItemTool, getItemTool, updateItemTool)
	return nil
}
