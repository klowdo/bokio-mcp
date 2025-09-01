package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterUploadTools registers file upload-related MCP tools
func RegisterUploadTools(server *mcp.Server, client *bokio.Client) error {
	// Register bokio_list_uploads tool
	if err := server.RegisterTool("bokio_list_uploads", mcp.Tool{
		Name: "bokio_list_uploads",
		Description: "List uploaded files with optional filtering and pagination",
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
				"status": map[string]interface{}{
					"type": "string",
					"description": "Filter by upload status",
					"enum": []string{"pending", "processed", "failed"},
				},
			},
		},
		Handler: createListUploadsHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_list_uploads tool: %w", err)
	}

	// Register bokio_upload_file tool
	if err := server.RegisterTool("bokio_upload_file", mcp.Tool{
		Name: "bokio_upload_file",
		Description: "Upload a file to Bokio for document management",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"file_path": map[string]interface{}{
					"type": "string",
					"description": "Path to the file to upload",
				},
				"description": map[string]interface{}{
					"type": "string",
					"description": "Optional description for the file",
				},
				"category": map[string]interface{}{
					"type": "string",
					"description": "File category (e.g., 'invoice', 'receipt', 'contract')",
					"enum": []string{"invoice", "receipt", "contract", "bank_statement", "other"},
				},
			},
			"required": []string{"file_path"},
		},
		Handler: createUploadFileHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_upload_file tool: %w", err)
	}

	// Register bokio_get_upload tool
	if err := server.RegisterTool("bokio_get_upload", mcp.Tool{
		Name: "bokio_get_upload",
		Description: "Get information about a specific uploaded file",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
					"description": "Upload ID",
				},
			},
			"required": []string{"id"},
		},
		Handler: createGetUploadHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_get_upload tool: %w", err)
	}

	// Register bokio_download_file tool
	if err := server.RegisterTool("bokio_download_file", mcp.Tool{
		Name: "bokio_download_file",
		Description: "Download a file from Bokio",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
					"description": "Upload ID",
				},
				"output_path": map[string]interface{}{
					"type": "string",
					"description": "Path where to save the downloaded file",
				},
			},
			"required": []string{"id", "output_path"},
		},
		Handler: createDownloadFileHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_download_file tool: %w", err)
	}

	// Register bokio_delete_upload tool
	if err := server.RegisterTool("bokio_delete_upload", mcp.Tool{
		Name: "bokio_delete_upload",
		Description: "Delete an uploaded file",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{
					"type": "string",
					"description": "Upload ID",
				},
			},
			"required": []string{"id"},
		},
		Handler: createDeleteUploadHandler(client),
	}); err != nil {
		return fmt.Errorf("failed to register bokio_delete_upload tool: %w", err)
	}

	return nil
}

// createListUploadsHandler creates the handler for the list uploads tool
func createListUploadsHandler(client *bokio.Client) mcp.ToolHandler {
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
		
		if status, ok := params["status"].(string); ok && status != "" {
			queryParams["status"] = status
		}

		// Construct URL with query parameters
		path := "/uploads"
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
				"error": fmt.Sprintf("Failed to list uploads: %v", err),
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var uploadList bokio.ListResponse[bokio.Upload]
		if err := json.Unmarshal(resp.Body(), &uploadList); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": uploadList.Data,
			"pagination": uploadList.Meta,
		}, nil
	}
}

// createUploadFileHandler creates the handler for the upload file tool
func createUploadFileHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		filePath, ok := params["file_path"].(string)
		if !ok || filePath == "" {
			return nil, fmt.Errorf("file_path is required")
		}

		// Check if file exists
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			if os.IsNotExist(err) {
				return map[string]interface{}{
					"success": false,
					"error": fmt.Sprintf("File not found: %s", filePath),
				}, nil
			}
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Error accessing file: %v", err),
			}, nil
		}

		// Read file content
		fileContent, err := os.ReadFile(filePath)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to read file: %v", err),
			}, nil
		}

		// Prepare upload request
		uploadRequest := map[string]interface{}{
			"filename":     fileInfo.Name(),
			"size":         fileInfo.Size(),
			"content_type": getContentType(filePath),
			"content":      fileContent, // In a real implementation, this might be base64 encoded
		}

		if description, ok := params["description"].(string); ok && description != "" {
			uploadRequest["description"] = description
		}

		if category, ok := params["category"].(string); ok && category != "" {
			uploadRequest["category"] = category
		}

		resp, err := client.Post(ctx, "/uploads", uploadRequest)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to upload file: %v", err),
			}, nil
		}

		if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var upload bokio.Upload
		if err := json.Unmarshal(resp.Body(), &upload); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": upload,
			"message": "File uploaded successfully",
		}, nil
	}
}

// createGetUploadHandler creates the handler for the get upload tool
func createGetUploadHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		id, ok := params["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("upload ID is required")
		}

		resp, err := client.Get(ctx, "/uploads/"+id)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to get upload: %v", err),
			}, nil
		}

		if resp.StatusCode() == http.StatusNotFound {
			return map[string]interface{}{
				"success": false,
				"error": "Upload not found",
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		var upload bokio.Upload
		if err := json.Unmarshal(resp.Body(), &upload); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to parse response: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"data": upload,
		}, nil
	}
}

// createDownloadFileHandler creates the handler for the download file tool
func createDownloadFileHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		id, ok := params["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("upload ID is required")
		}

		outputPath, ok := params["output_path"].(string)
		if !ok || outputPath == "" {
			return nil, fmt.Errorf("output_path is required")
		}

		resp, err := client.Get(ctx, "/uploads/"+id+"/download")
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to download file: %v", err),
			}, nil
		}

		if resp.StatusCode() == http.StatusNotFound {
			return map[string]interface{}{
				"success": false,
				"error": "Upload not found",
			}, nil
		}

		if resp.StatusCode() != http.StatusOK {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		// Write file content to output path
		if err := os.WriteFile(outputPath, resp.Body(), 0644); err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to write file: %v", err),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("File downloaded successfully to: %s", outputPath),
			"output_path": outputPath,
			"size": len(resp.Body()),
		}, nil
	}
}

// createDeleteUploadHandler creates the handler for the delete upload tool
func createDeleteUploadHandler(client *bokio.Client) mcp.ToolHandler {
	return func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
		if !client.IsAuthenticated() {
			return map[string]interface{}{
				"success": false,
				"error": "Not authenticated. Use bokio_authenticate first.",
			}, nil
		}

		id, ok := params["id"].(string)
		if !ok || id == "" {
			return nil, fmt.Errorf("upload ID is required")
		}

		resp, err := client.Delete(ctx, "/uploads/"+id)
		if err != nil {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("Failed to delete upload: %v", err),
			}, nil
		}

		if resp.StatusCode() == http.StatusNotFound {
			return map[string]interface{}{
				"success": false,
				"error": "Upload not found",
			}, nil
		}

		if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
			return map[string]interface{}{
				"success": false,
				"error": fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
			}, nil
		}

		return map[string]interface{}{
			"success": true,
			"message": "Upload deleted successfully",
		}, nil
	}
}

// getContentType returns the MIME type based on file extension
func getContentType(filePath string) string {
	// Basic content type detection based on file extension
	// In a real implementation, you might use a more sophisticated library
	switch {
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".pdf":
		return "application/pdf"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".jpg" || filePath[len(filePath)-5:] == ".jpeg":
		return "image/jpeg"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".png":
		return "image/png"
	case len(filePath) >= 4 && filePath[len(filePath)-4:] == ".txt":
		return "text/plain"
	case len(filePath) >= 5 && filePath[len(filePath)-5:] == ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case len(filePath) >= 5 && filePath[len(filePath)-5:] == ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		return "application/octet-stream"
	}
}