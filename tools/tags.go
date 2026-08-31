package tools

import (
	"context"

	"github.com/google/uuid"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// TagGroupsListParams defines parameters for listing tag groups
type TagGroupsListParams struct {
	CompanyID string  `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Page      *int32  `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32  `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
	Query     *string `json:"query,omitempty" jsonschema:"Filter query (optional)"`
}

// TagGroupRefParams identifies a single tag group
type TagGroupRefParams struct {
	CompanyID  string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	TagGroupID string `json:"tag_group_id" jsonschema:"Tag group UUID"`
}

// TagGroupCreateParams defines parameters for creating a tag group
type TagGroupCreateParams struct {
	CompanyID       string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Name            string `json:"name" jsonschema:"Tag group name"`
	Alias           string `json:"alias" jsonschema:"Short alias for the tag group"`
	DimensionNumber *int32 `json:"dimension_number,omitempty" jsonschema:"SIE dimension number (optional)"`
}

// TagGroupUpdateParams defines parameters for updating a tag group
type TagGroupUpdateParams struct {
	CompanyID  string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	TagGroupID string `json:"tag_group_id" jsonschema:"Tag group UUID to update"`
	Name       string `json:"name" jsonschema:"Tag group name"`
	Alias      string `json:"alias" jsonschema:"Short alias for the tag group"`
}

// TagCreateParams defines parameters for creating a tag in a tag group
type TagCreateParams struct {
	CompanyID  string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	TagGroupID string `json:"tag_group_id" jsonschema:"Tag group UUID to add the tag to"`
	Name       string `json:"name" jsonschema:"Tag name"`
}

// TagRefParams identifies a single tag within a tag group
type TagRefParams struct {
	CompanyID  string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	TagGroupID string `json:"tag_group_id" jsonschema:"Tag group UUID"`
	TagID      string `json:"tag_id" jsonschema:"Tag UUID"`
}

// TagUpdateParams defines parameters for renaming a tag
type TagUpdateParams struct {
	CompanyID  string `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	TagGroupID string `json:"tag_group_id" jsonschema:"Tag group UUID"`
	TagID      string `json:"tag_id" jsonschema:"Tag UUID to update"`
	Name       string `json:"name" jsonschema:"New tag name"`
}

// JournalEntryTagParams applies one tag with an allocation weight
type JournalEntryTagParams struct {
	TagID  string  `json:"tag_id" jsonschema:"Tag UUID"`
	Weight float64 `json:"weight" jsonschema:"Allocation weight above 0 and up to 1"`
}

// JournalEntryItemTagsParams applies tags to a single journal entry line item
type JournalEntryItemTagsParams struct {
	ItemID int64                   `json:"item_id" jsonschema:"Journal entry line item id"`
	Tags   []JournalEntryTagParams `json:"tags" jsonschema:"Tags to apply to this line item"`
}

// JournalEntryTagsSetParams defines parameters for setting tags on a journal entry
type JournalEntryTagsSetParams struct {
	CompanyID      string                       `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	JournalEntryID string                       `json:"journal_entry_id" jsonschema:"Journal entry UUID"`
	Tags           []JournalEntryTagParams      `json:"tags,omitempty" jsonschema:"Verification-level tags applied to the whole entry"`
	ItemTags       []JournalEntryItemTagsParams `json:"item_tags,omitempty" jsonschema:"Transaction-level tags per line item; cannot be combined with verification-level tags"`
}

func buildJournalEntryTags(tags []JournalEntryTagParams) ([]company.JournalEntryTag, *mcp.CallToolResult) {
	built := make([]company.JournalEntryTag, 0, len(tags))
	for _, tag := range tags {
		tagUUID, err := uuid.Parse(tag.TagID)
		if err != nil {
			return nil, toolError("Invalid tag ID format: %v", err)
		}
		built = append(built, company.JournalEntryTag{TagId: tagUUID, Weight: tag.Weight})
	}
	return built, nil
}

// RegisterTagTools registers tag group, tag and journal entry tagging tools
func RegisterTagTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_tag_groups_list",
		Description: "List tag groups (cost centres, projects and other dimensions) for a company",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args TagGroupsListParams) (*mcp.CallToolResult, any, error) {
		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetTagGroups(ctx, companyUUID, &company.GetTagGroupsParams{
			Page:     args.Page,
			PageSize: args.PageSize,
			Query:    args.Query,
		})

		return renderAPIResponse("list tag groups", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_tag_groups_get",
		Description: "Get a single tag group with its tags",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args TagGroupRefParams) (*mcp.CallToolResult, any, error) {
		companyUUID, tagGroupUUID, errResult := resolveEntityRef(args.CompanyID, args.TagGroupID, "tag group")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.GetTagGroupById(ctx, companyUUID, tagGroupUUID)

		return renderAPIResponse("get tag group", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_tag_groups_create",
		Description: "Create a new tag group",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args TagGroupCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		if args.Name == "" || args.Alias == "" {
			return toolError("Tag group name and alias are required"), nil, nil
		}

		resp, err := client.CompanyClient.PostTagGroup(ctx, companyUUID, company.TagGroupWrite{
			Name:            args.Name,
			Alias:           args.Alias,
			DimensionNumber: args.DimensionNumber,
		})

		return renderAPIResponse("create tag group", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_tag_groups_update",
		Description: "Update the name or alias of a tag group",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args TagGroupUpdateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, tagGroupUUID, errResult := resolveEntityRef(args.CompanyID, args.TagGroupID, "tag group")
		if errResult != nil {
			return errResult, nil, nil
		}

		if args.Name == "" || args.Alias == "" {
			return toolError("Tag group name and alias are required"), nil, nil
		}

		resp, err := client.CompanyClient.PutTagGroup(ctx, companyUUID, tagGroupUUID, company.TagGroupUpdate{
			Name:  args.Name,
			Alias: args.Alias,
		})

		return renderAPIResponse("update tag group", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_tag_groups_delete",
		Description: "Delete a tag group and its tags",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args TagGroupRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, tagGroupUUID, errResult := resolveEntityRef(args.CompanyID, args.TagGroupID, "tag group")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteTagGroup(ctx, companyUUID, tagGroupUUID)

		return renderAPIResponse("delete tag group", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_tags_create",
		Description: "Create a tag inside a tag group",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args TagCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, tagGroupUUID, errResult := resolveEntityRef(args.CompanyID, args.TagGroupID, "tag group")
		if errResult != nil {
			return errResult, nil, nil
		}

		if args.Name == "" {
			return toolError("Tag name is required"), nil, nil
		}

		resp, err := client.CompanyClient.PostTagGroupTag(ctx, companyUUID, tagGroupUUID, company.TagValueWrite{
			Name: args.Name,
		})

		return renderAPIResponse("create tag", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_tags_update",
		Description: "Rename a tag inside a tag group",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args TagUpdateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, tagGroupUUID, tagUUID, errResult := resolveNestedRef(args.CompanyID, args.TagGroupID, args.TagID, "tag group", "tag")
		if errResult != nil {
			return errResult, nil, nil
		}

		if args.Name == "" {
			return toolError("Tag name is required"), nil, nil
		}

		resp, err := client.CompanyClient.PutTagGroupTag(ctx, companyUUID, tagGroupUUID, tagUUID, company.TagValueWrite{
			Name: args.Name,
		})

		return renderAPIResponse("update tag", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_tags_delete",
		Description: "Delete a tag from a tag group",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args TagRefParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, tagGroupUUID, tagUUID, errResult := resolveNestedRef(args.CompanyID, args.TagGroupID, args.TagID, "tag group", "tag")
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.DeleteTagGroupTag(ctx, companyUUID, tagGroupUUID, tagUUID)

		return renderAPIResponse("delete tag", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_journal_entries_tags_set",
		Description: "Set tags on a journal entry, either on the whole verification or per line item, but not both",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args JournalEntryTagsSetParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, journalEntryUUID, errResult := resolveEntityRef(args.CompanyID, args.JournalEntryID, "journal entry")
		if errResult != nil {
			return errResult, nil, nil
		}

		if len(args.Tags) > 0 && len(args.ItemTags) > 0 {
			return toolError("Set either tags or item_tags on a journal entry, not both"), nil, nil
		}

		body := company.JournalEntryTags{}

		if len(args.Tags) > 0 {
			tags, errResult := buildJournalEntryTags(args.Tags)
			if errResult != nil {
				return errResult, nil, nil
			}
			body.Tags = &tags
		}

		if len(args.ItemTags) > 0 {
			items := make([]struct {
				Id   int64                     `json:"id"`
				Tags []company.JournalEntryTag `json:"tags"`
			}, 0, len(args.ItemTags))
			for _, itemTags := range args.ItemTags {
				tags, errResult := buildJournalEntryTags(itemTags.Tags)
				if errResult != nil {
					return errResult, nil, nil
				}
				items = append(items, struct {
					Id   int64                     `json:"id"`
					Tags []company.JournalEntryTag `json:"tags"`
				}{Id: itemTags.ItemID, Tags: tags})
			}
			body.Items = &items
		}

		resp, err := client.CompanyClient.PutJournalentryTags(ctx, companyUUID, journalEntryUUID, body)

		return renderAPIResponse("set journal entry tags", resp, err)
	})

	return nil
}
