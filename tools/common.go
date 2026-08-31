package tools

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func resolveNestedRef(companyID, parentID, childID, parentLabel, childLabel string) (uuid.UUID, uuid.UUID, uuid.UUID, *mcp.CallToolResult) {
	companyUUID, parentUUID, errResult := resolveEntityRef(companyID, parentID, parentLabel)
	if errResult != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, errResult
	}

	if childID == "" {
		return uuid.Nil, uuid.Nil, uuid.Nil, toolError("%s ID is required", childLabel)
	}

	childUUID, err := uuid.Parse(childID)
	if err != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, toolError("Invalid %s ID format: %v", childLabel, err)
	}

	return companyUUID, parentUUID, childUUID, nil
}

func renderBinaryResponse(label string, resp *http.Response, err error) (*mcp.CallToolResult, any, error) {
	if err != nil {
		return toolError("Failed to %s: %v", label, err), nil, nil
	}
	defer resp.Body.Close()

	content, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return toolError("Failed to read %s content: %v", label, readErr), nil, nil
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return toolError("API returned status %d: %s", resp.StatusCode, content), nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{
				Text: fmt.Sprintf("✅ Successfully completed: %s\n\nContent-Type: %s\nFile Size: %d bytes\nStatus: %d\n\nBase64 Content: %s",
					label, resp.Header.Get("Content-Type"), len(content), resp.StatusCode,
					base64.StdEncoding.EncodeToString(content)),
			},
		},
	}, nil, nil
}
