package tools

import (
	"bytes"
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
	CompanyID string  `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	Page      *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32  `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
	Query     *string `json:"query,omitempty" jsonschema:"Optional query to filter items (optional)"`
}

// ItemCreateParams defines parameters for creating an item
type ItemCreateParams struct {
	CompanyID   string   `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	ItemType    string   `json:"item_type" jsonschema:"Type of item: 'salesItem' or 'descriptionOnlyItem'"`
	Description string   `json:"description" jsonschema:"Item description"`
	UnitPrice   *float64 `json:"unit_price,omitempty" jsonschema:"Unit price (required for salesItem)"`
	TaxRate     *float64 `json:"tax_rate,omitempty" jsonschema:"Tax rate as decimal (e.g., 0.25 for 25%, required for salesItem)"`
	ProductType *string  `json:"product_type,omitempty" jsonschema:"Product type: 'goods' or 'services' (for salesItem, defaults to 'goods')"`
	UnitType    *string  `json:"unit_type,omitempty" jsonschema:"Unit type: 'piece', 'hour', 'meter', etc. (for salesItem, defaults to 'piece')"`
}

// ItemGetParams defines parameters for getting a specific item
type ItemGetParams struct {
	CompanyID string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	ItemID    string `json:"item_id" jsonschema:"Item UUID"`
}

// ItemUpdateParams defines parameters for updating an item
type ItemUpdateParams struct {
	CompanyID   string   `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	ItemID      string   `json:"item_id" jsonschema:"Item UUID"`
	ItemType    string   `json:"item_type" jsonschema:"Type of item: 'salesItem' or 'descriptionOnlyItem'"`
	Description string   `json:"description" jsonschema:"Item description"`
	UnitPrice   *float64 `json:"unit_price,omitempty" jsonschema:"Unit price (required for salesItem)"`
	TaxRate     *float64 `json:"tax_rate,omitempty" jsonschema:"Tax rate as decimal (e.g., 0.25 for 25%, required for salesItem)"`
	ProductType *string  `json:"product_type,omitempty" jsonschema:"Product type: 'goods' or 'services' (for salesItem, defaults to 'goods')"`
	UnitType    *string  `json:"unit_type,omitempty" jsonschema:"Unit type: 'piece', 'hour', 'meter', etc. (for salesItem, defaults to 'piece')"`
}

// RegisterItemTools registers item management tools using ONLY generated API clients
func RegisterItemTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_items_list",
		Description: "List inventory items for a company with optional pagination and filtering",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ItemListParams) (*mcp.CallToolResult, any, error) {
		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, nil, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, nil, nil
		}

		genParams := &company.GetItemsParams{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Query,
		}

		resp, err := client.CompanyClient.GetItems(ctx, companyUUID, genParams)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to list items: %v", err),
					},
				},
			}, nil, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, nil, nil
		}

		var responseData json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved items\n\nCompany: %s\nStatus: %d\nResponse: %s", companyIDStr, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_items_create",
		Description: "Create a new inventory item (salesItem or descriptionOnlyItem)",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ItemCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Create operation not allowed in read-only mode",
					},
				},
			}, nil, nil
		}

		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, nil, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, nil, nil
		}

		if args.ItemType == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "item_type is required (salesItem or descriptionOnlyItem)",
					},
				},
			}, nil, nil
		}

		if args.Description == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "description is required",
					},
				},
			}, nil, nil
		}

		var itemBody []byte

		switch args.ItemType {
		case "salesItem":
			if args.UnitPrice == nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "unit_price is required for salesItem",
						},
					},
				}, nil, nil
			}
			if args.TaxRate == nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "tax_rate is required for salesItem",
						},
					},
				}, nil, nil
			}

			productType := "goods"
			if args.ProductType != nil {
				productType = *args.ProductType
			}
			unitType := "piece"
			if args.UnitType != nil {
				unitType = *args.UnitType
			}

			salesItem := company.SalesItem{
				Description: args.Description,
				ItemType:    company.SalesItemItemTypeSalesItem,
				ProductType: company.SalesItemProductType(productType),
				TaxRate:     *args.TaxRate,
				UnitPrice:   *args.UnitPrice,
				UnitType:    company.SalesItemUnitType(unitType),
			}

			salesItemJSON, err := json.Marshal(salesItem)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to marshal salesItem: %v", err),
						},
					},
				}, nil, nil
			}
			itemBody = salesItemJSON

		case "descriptionOnlyItem":
			descItem := company.DescriptionOnlyItem{
				Description: args.Description,
				ItemType:    company.DescriptionOnlyItemItemTypeDescriptionOnlyItem,
			}

			descItemJSON, err := json.Marshal(descItem)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to marshal descriptionOnlyItem: %v", err),
						},
					},
				}, nil, nil
			}
			itemBody = descItemJSON

		default:
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "item_type must be either 'salesItem' or 'descriptionOnlyItem'",
					},
				},
			}, nil, nil
		}

		resp, err := client.CompanyClient.PostItemWithBody(ctx, companyUUID, "application/json", bytes.NewReader(itemBody))
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to create item: %v", err),
					},
				},
			}, nil, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, nil, nil
		}

		var responseData json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully created item\n\nCompany: %s\nItem Type: %s\nDescription: %s\nStatus: %d\nResponse: %s", companyIDStr, args.ItemType, args.Description, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_items_get",
		Description: "Get a specific inventory item by ID",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ItemGetParams) (*mcp.CallToolResult, any, error) {
		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, nil, nil
		}

		if args.ItemID == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "item_id is required",
					},
				},
			}, nil, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, nil, nil
		}

		itemUUID, err := uuid.Parse(args.ItemID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid item ID format: %v", err),
					},
				},
			}, nil, nil
		}

		resp, err := client.CompanyClient.GetItemsItemId(ctx, companyUUID, itemUUID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to get item: %v", err),
					},
				},
			}, nil, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, nil, nil
		}

		var responseData json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved item\n\nCompany: %s\nItem ID: %s\nStatus: %d\nResponse: %s", companyIDStr, args.ItemID, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_items_update",
		Description: "Update an existing inventory item",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args ItemUpdateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Update operation not allowed in read-only mode",
					},
				},
			}, nil, nil
		}

		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, nil, nil
		}

		if args.ItemID == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "item_id is required",
					},
				},
			}, nil, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, nil, nil
		}

		itemUUID, err := uuid.Parse(args.ItemID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid item ID format: %v", err),
					},
				},
			}, nil, nil
		}

		if args.ItemType == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "item_type is required (salesItem or descriptionOnlyItem)",
					},
				},
			}, nil, nil
		}

		if args.Description == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "description is required",
					},
				},
			}, nil, nil
		}

		var itemBody []byte

		switch args.ItemType {
		case "salesItem":
			if args.UnitPrice == nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "unit_price is required for salesItem",
						},
					},
				}, nil, nil
			}
			if args.TaxRate == nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "tax_rate is required for salesItem",
						},
					},
				}, nil, nil
			}

			productType := "goods"
			if args.ProductType != nil {
				productType = *args.ProductType
			}
			unitType := "piece"
			if args.UnitType != nil {
				unitType = *args.UnitType
			}

			salesItem := company.SalesItem{
				Description: args.Description,
				Id:          &itemUUID,
				ItemType:    company.SalesItemItemTypeSalesItem,
				ProductType: company.SalesItemProductType(productType),
				TaxRate:     *args.TaxRate,
				UnitPrice:   *args.UnitPrice,
				UnitType:    company.SalesItemUnitType(unitType),
			}

			salesItemJSON, err := json.Marshal(salesItem)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to marshal salesItem: %v", err),
						},
					},
				}, nil, nil
			}
			itemBody = salesItemJSON

		case "descriptionOnlyItem":
			descItem := company.DescriptionOnlyItem{
				Description: args.Description,
				Id:          &itemUUID,
				ItemType:    company.DescriptionOnlyItemItemTypeDescriptionOnlyItem,
			}

			descItemJSON, err := json.Marshal(descItem)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to marshal descriptionOnlyItem: %v", err),
						},
					},
				}, nil, nil
			}
			itemBody = descItemJSON

		default:
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "item_type must be either 'salesItem' or 'descriptionOnlyItem'",
					},
				},
			}, nil, nil
		}

		resp, err := client.CompanyClient.PutItemWithBody(ctx, companyUUID, itemUUID, "application/json", bytes.NewReader(itemBody))
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to update item: %v", err),
					},
				},
			}, nil, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, nil, nil
		}

		var responseData json.RawMessage
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, nil, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully updated item\n\nCompany: %s\nItem ID: %s\nItem Type: %s\nDescription: %s\nStatus: %d\nResponse: %s", companyIDStr, args.ItemID, args.ItemType, args.Description, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	return nil
}
