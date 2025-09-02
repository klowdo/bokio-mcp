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

// ListUploadsParams defines the parameters for listing uploads
type ListUploadsParams struct {
	Page    *int    `json:"page,omitempty"`
	PerPage *int    `json:"per_page,omitempty"`
	Status  *string `json:"status,omitempty"`
}

// ListUploadsResult defines the result of listing uploads
type ListUploadsResult struct {
	Success    bool                   `json:"success"`
	Data       []bokio.Upload         `json:"data,omitempty"`
	Pagination interface{}            `json:"pagination,omitempty"`
	Error      string                 `json:"error,omitempty"`
}

// UploadFileParams defines the parameters for uploading a file
type UploadFileParams struct {
	FilePath    string  `json:"file_path"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
}

// UploadFileResult defines the result of uploading a file
type UploadFileResult struct {
	Success bool          `json:"success"`
	Data    *bokio.Upload `json:"data,omitempty"`
	Message string        `json:"message,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// GetUploadParams defines the parameters for getting an upload
type GetUploadParams struct {
	ID string `json:"id"`
}

// GetUploadResult defines the result of getting an upload
type GetUploadResult struct {
	Success bool          `json:"success"`
	Data    *bokio.Upload `json:"data,omitempty"`
	Error   string        `json:"error,omitempty"`
}

// DownloadFileParams defines the parameters for downloading a file
type DownloadFileParams struct {
	ID         string `json:"id"`
	OutputPath string `json:"output_path"`
}

// DownloadFileResult defines the result of downloading a file
type DownloadFileResult struct {
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
	OutputPath string `json:"output_path,omitempty"`
	Size       int    `json:"size,omitempty"`
	Error      string `json:"error,omitempty"`
}

// DeleteUploadParams defines the parameters for deleting an upload
type DeleteUploadParams struct {
	ID string `json:"id"`
}

// DeleteUploadResult defines the result of deleting an upload
type DeleteUploadResult struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// RegisterUploadTools registers file upload-related MCP tools
func RegisterUploadTools(server *mcp.Server, client *bokio.Client) error {
	// Register bokio_list_uploads tool
	listUploadsTool := mcp.NewServerTool[ListUploadsParams, ListUploadsResult](
		"bokio_list_uploads",
		"List uploaded files with optional filtering and pagination",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[ListUploadsParams]) (*mcp.CallToolResultFor[ListUploadsResult], error) {
			if !client.IsAuthenticated() {
				return &mcp.CallToolResultFor[ListUploadsResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Not authenticated. Use bokio_authenticate first.",
						},
					},
				}, nil
			}

			// Build query parameters
			queryParams := make(map[string]string)
			
			if params.Arguments.Page != nil {
				queryParams["page"] = fmt.Sprintf("%d", *params.Arguments.Page)
			}
			
			if params.Arguments.PerPage != nil {
				queryParams["per_page"] = fmt.Sprintf("%d", *params.Arguments.PerPage)
			}
			
			if params.Arguments.Status != nil && *params.Arguments.Status != "" {
				queryParams["status"] = *params.Arguments.Status
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

			resp, err := client.GET(ctx, path)
			if err != nil {
				return &mcp.CallToolResultFor[ListUploadsResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to list uploads: %v", err),
						},
					},
				}, nil
			}

			if resp.StatusCode() != http.StatusOK {
				return &mcp.CallToolResultFor[ListUploadsResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
						},
					},
				}, nil
			}

			var uploadList bokio.UploadsResponse
			if err := json.Unmarshal(resp.Body(), &uploadList); err != nil {
				return &mcp.CallToolResultFor[ListUploadsResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to parse response: %v", err),
						},
					},
				}, nil
			}

			return &mcp.CallToolResultFor[ListUploadsResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Found %d uploads", len(uploadList.Items)),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("page",
				mcp.Description("Page number for pagination (default: 1)"),
					),
			mcp.Property("per_page",
				mcp.Description("Number of items per page (default: 25, max: 100)"),
							),
			mcp.Property("status",
				mcp.Description("Filter by upload status"),
					),
		),
	)
	
	server.AddTools(listUploadsTool)

	// Register bokio_upload_file tool
	uploadFileTool := mcp.NewServerTool[UploadFileParams, UploadFileResult](
		"bokio_upload_file",
		"Upload a file to Bokio for document management",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[UploadFileParams]) (*mcp.CallToolResultFor[UploadFileResult], error) {
			if !client.IsAuthenticated() {
				return &mcp.CallToolResultFor[UploadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Not authenticated. Use bokio_authenticate first.",
						},
					},
				}, nil
			}

			filePath := params.Arguments.FilePath
			if filePath == "" {
				return &mcp.CallToolResultFor[UploadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "file_path is required",
						},
					},
				}, fmt.Errorf("file_path is required")
			}

			// Check if file exists
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				if os.IsNotExist(err) {
					return &mcp.CallToolResultFor[UploadFileResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
									Text: fmt.Sprintf("File not found: %s", filePath),
							},
						},
					}, nil
				}
				return &mcp.CallToolResultFor[UploadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Error accessing file: %v", err),
						},
					},
				}, nil
			}

			// Read file content
			fileContent, err := os.ReadFile(filePath)
			if err != nil {
				return &mcp.CallToolResultFor[UploadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to read file: %v", err),
						},
					},
				}, nil
			}

			// Prepare upload request
			uploadRequest := map[string]interface{}{
				"filename":     fileInfo.Name(),
				"size":         fileInfo.Size(),
				"content_type": getContentType(filePath),
				"content":      fileContent, // In a real implementation, this might be base64 encoded
			}

			if params.Arguments.Description != nil && *params.Arguments.Description != "" {
				uploadRequest["description"] = *params.Arguments.Description
			}

			if params.Arguments.Category != nil && *params.Arguments.Category != "" {
				uploadRequest["category"] = *params.Arguments.Category
			}

			resp, err := client.POST(ctx, "/uploads", uploadRequest)
			if err != nil {
				return &mcp.CallToolResultFor[UploadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to upload file: %v", err),
						},
					},
				}, nil
			}

			if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
				return &mcp.CallToolResultFor[UploadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
						},
					},
				}, nil
			}

			var upload bokio.Upload
			if err := json.Unmarshal(resp.Body(), &upload); err != nil {
				return &mcp.CallToolResultFor[UploadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to parse response: %v", err),
						},
					},
				}, nil
			}

			return &mcp.CallToolResultFor[UploadFileResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("File uploaded successfully: %s (ID: %s)", upload.Description, upload.ID),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("file_path",
				mcp.Description("Path to the file to upload"),
				mcp.Required(true),
			),
			mcp.Property("description",
				mcp.Description("Optional description for the file"),
			),
			mcp.Property("category",
				mcp.Description("File category"),
				mcp.Enum("invoice", "receipt", "contract", "bank_statement", "other"),
			),
		),
	)
	
	server.AddTools(uploadFileTool)

	// Register bokio_get_upload tool
	getUploadTool := mcp.NewServerTool[GetUploadParams, GetUploadResult](
		"bokio_get_upload",
		"Get information about a specific uploaded file",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[GetUploadParams]) (*mcp.CallToolResultFor[GetUploadResult], error) {
			if !client.IsAuthenticated() {
				return &mcp.CallToolResultFor[GetUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Not authenticated. Use bokio_authenticate first.",
						},
					},
				}, nil
			}

			id := params.Arguments.ID
			if id == "" {
				return &mcp.CallToolResultFor[GetUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Upload ID is required",
						},
					},
				}, fmt.Errorf("upload ID is required")
			}

			resp, err := client.GET(ctx, "/uploads/"+id)
			if err != nil {
				return &mcp.CallToolResultFor[GetUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to get upload: %v", err),
						},
					},
				}, nil
			}

			if resp.StatusCode() == http.StatusNotFound {
				return &mcp.CallToolResultFor[GetUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Upload not found",
						},
					},
				}, nil
			}

			if resp.StatusCode() != http.StatusOK {
				return &mcp.CallToolResultFor[GetUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
						},
					},
				}, nil
			}

			var upload bokio.Upload
			if err := json.Unmarshal(resp.Body(), &upload); err != nil {
				return &mcp.CallToolResultFor[GetUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to parse response: %v", err),
						},
					},
				}, nil
			}

			return &mcp.CallToolResultFor[GetUploadResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Upload: %s (ID: %s)", upload.Description, upload.ID),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("id",
				mcp.Description("Upload ID"),
				mcp.Required(true),
			),
		),
	)
	
	server.AddTools(getUploadTool)

	// Register bokio_download_file tool
	downloadFileTool := mcp.NewServerTool[DownloadFileParams, DownloadFileResult](
		"bokio_download_file",
		"Download a file from Bokio",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[DownloadFileParams]) (*mcp.CallToolResultFor[DownloadFileResult], error) {
			if !client.IsAuthenticated() {
				return &mcp.CallToolResultFor[DownloadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Not authenticated. Use bokio_authenticate first.",
						},
					},
				}, nil
			}

			id := params.Arguments.ID
			if id == "" {
				return &mcp.CallToolResultFor[DownloadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Upload ID is required",
						},
					},
				}, fmt.Errorf("upload ID is required")
			}

			outputPath := params.Arguments.OutputPath
			if outputPath == "" {
				return &mcp.CallToolResultFor[DownloadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Output path is required",
						},
					},
				}, fmt.Errorf("output path is required")
			}

			resp, err := client.GET(ctx, "/uploads/"+id+"/download")
			if err != nil {
				return &mcp.CallToolResultFor[DownloadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to download file: %v", err),
						},
					},
				}, nil
			}

			if resp.StatusCode() == http.StatusNotFound {
				return &mcp.CallToolResultFor[DownloadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Upload not found",
						},
					},
				}, nil
			}

			if resp.StatusCode() != http.StatusOK {
				return &mcp.CallToolResultFor[DownloadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
						},
					},
				}, nil
			}

			// Write file content to output path
			if err := os.WriteFile(outputPath, resp.Body(), 0644); err != nil {
				return &mcp.CallToolResultFor[DownloadFileResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to write file: %v", err),
						},
					},
				}, nil
			}

			return &mcp.CallToolResultFor[DownloadFileResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("File downloaded successfully to: %s", outputPath),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("id",
				mcp.Description("Upload ID"),
				mcp.Required(true),
			),
			mcp.Property("output_path",
				mcp.Description("Path where to save the downloaded file"),
				mcp.Required(true),
			),
		),
	)
	
	server.AddTools(downloadFileTool)

	// Register bokio_delete_upload tool
	deleteUploadTool := mcp.NewServerTool[DeleteUploadParams, DeleteUploadResult](
		"bokio_delete_upload",
		"Delete an uploaded file",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[DeleteUploadParams]) (*mcp.CallToolResultFor[DeleteUploadResult], error) {
			if !client.IsAuthenticated() {
				return &mcp.CallToolResultFor[DeleteUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Not authenticated. Use bokio_authenticate first.",
						},
					},
				}, nil
			}

			id := params.Arguments.ID
			if id == "" {
				return &mcp.CallToolResultFor[DeleteUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Upload ID is required",
						},
					},
				}, fmt.Errorf("upload ID is required")
			}

			resp, err := client.DELETE(ctx, "/uploads/"+id)
			if err != nil {
				return &mcp.CallToolResultFor[DeleteUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to delete upload: %v", err),
						},
					},
				}, nil
			}

			if resp.StatusCode() == http.StatusNotFound {
				return &mcp.CallToolResultFor[DeleteUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Upload not found",
						},
					},
				}, nil
			}

			if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusNoContent {
				return &mcp.CallToolResultFor[DeleteUploadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API error: %d - %s", resp.StatusCode(), resp.String()),
						},
					},
				}, nil
			}

			return &mcp.CallToolResultFor[DeleteUploadResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Upload deleted successfully",
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("id",
				mcp.Description("Upload ID"),
				mcp.Required(true),
			),
		),
	)
	
	server.AddTools(deleteUploadTool)

	return nil
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