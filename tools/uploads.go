package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// UploadListParams defines parameters for listing uploads
type UploadListParams struct {
	CompanyID string `json:"company_id"`
	Page      *int32 `json:"page,omitempty"`
	PageSize  *int32 `json:"page_size,omitempty"`
}

// UploadListResult defines the result for listing uploads
type UploadListResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// UploadCreateParams defines parameters for creating an upload
type UploadCreateParams struct {
	CompanyID      string  `json:"company_id"`
	FileContent    string  `json:"file_content"` // Base64 encoded file content
	FileName       string  `json:"file_name"`    // Name of the file
	ContentType    string  `json:"content_type"` // MIME type of the file
	Description    *string `json:"description,omitempty"`
	JournalEntryID *string `json:"journal_entry_id,omitempty"`
}

// UploadCreateResult defines the result for creating an upload
type UploadCreateResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// UploadGetParams defines parameters for getting an upload
type UploadGetParams struct {
	CompanyID string `json:"company_id"`
	UploadID  string `json:"upload_id"`
}

// UploadGetResult defines the result for getting an upload
type UploadGetResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// UploadDownloadParams defines parameters for downloading an upload
type UploadDownloadParams struct {
	CompanyID string `json:"company_id"`
	UploadID  string `json:"upload_id"`
}

// UploadDownloadResult defines the result for downloading an upload
type UploadDownloadResult struct {
	Success     bool   `json:"success"`
	FileContent string `json:"file_content,omitempty"` // Base64 encoded file content
	ContentType string `json:"content_type,omitempty"`
	FileName    string `json:"file_name,omitempty"`
	Error       string `json:"error,omitempty"`
}

// RegisterUploadTools registers upload tools using ONLY generated API clients
func RegisterUploadTools(server *mcp.Server, client *bokio.AuthClient) error {
	// Tool to list uploads using generated client
	listUploadsTool := mcp.NewServerTool[UploadListParams, UploadListResult](
		"bokio_uploads_list",
		"List uploads for a company with optional pagination",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[UploadListParams]) (*mcp.CallToolResultFor[UploadListResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[UploadListResult]{
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
				return &mcp.CallToolResultFor[UploadListResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Create parameters for the generated client
			genParams := &company.GetUploadsParams{
				Page:     params.Arguments.Page,
				PageSize: params.Arguments.PageSize,
			}

			// Call the generated client method
			resp, err := client.CompanyClient.GetUploads(ctx, companyUUID, genParams)
			if err != nil {
				return &mcp.CallToolResultFor[UploadListResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to list uploads: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[UploadListResult]{
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
				return &mcp.CallToolResultFor[UploadListResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[UploadListResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully retrieved uploads list\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
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
		),
	)

	// Tool to create upload using generated client
	createUploadTool := mcp.NewServerTool[UploadCreateParams, UploadCreateResult](
		"bokio_uploads_create",
		"Upload a file to Bokio",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[UploadCreateParams]) (*mcp.CallToolResultFor[UploadCreateResult], error) {
			// Check read-only mode
			if client.GetConfig().ReadOnly {
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "Upload creation not allowed in read-only mode",
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
				return &mcp.CallToolResultFor[UploadCreateResult]{
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
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Validate required fields
			if params.Arguments.FileContent == "" {
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "file_content is required (base64 encoded file)",
						},
					},
				}, nil
			}

			if params.Arguments.FileName == "" {
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "file_name is required",
						},
					},
				}, nil
			}

			if params.Arguments.ContentType == "" {
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "content_type is required",
						},
					},
				}, nil
			}

			// Decode base64 file content
			fileData, err := base64.StdEncoding.DecodeString(params.Arguments.FileContent)
			if err != nil {
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid base64 file content: %v", err),
						},
					},
				}, nil
			}

			// Parse journal entry ID if provided
			var journalEntryUUID *openapi_types.UUID
			if params.Arguments.JournalEntryID != nil && *params.Arguments.JournalEntryID != "" {
				journalUUID, err := uuid.Parse(*params.Arguments.JournalEntryID)
				if err != nil {
					return &mcp.CallToolResultFor[UploadCreateResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Invalid journal entry ID format: %v", err),
							},
						},
					}, nil
				}
				journalEntryUUID = &journalUUID
			}

			// Create multipart form data
			var buf bytes.Buffer
			writer := multipart.NewWriter(&buf)

			// Add file field
			fileWriter, err := writer.CreateFormFile("file", params.Arguments.FileName)
			if err != nil {
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to create form file: %v", err),
						},
					},
				}, nil
			}
			_, err = fileWriter.Write(fileData)
			if err != nil {
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to write file data: %v", err),
						},
					},
				}, nil
			}

			// Add description field if provided
			if params.Arguments.Description != nil {
				err = writer.WriteField("description", *params.Arguments.Description)
				if err != nil {
					return &mcp.CallToolResultFor[UploadCreateResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Failed to write description field: %v", err),
							},
						},
					}, nil
				}
			}

			// Add journal entry ID field if provided
			if journalEntryUUID != nil {
				err = writer.WriteField("journalEntryId", journalEntryUUID.String())
				if err != nil {
					return &mcp.CallToolResultFor[UploadCreateResult]{
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: fmt.Sprintf("Failed to write journal entry ID field: %v", err),
							},
						},
					}, nil
				}
			}

			err = writer.Close()
			if err != nil {
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to close multipart writer: %v", err),
						},
					},
				}, nil
			}

			// Create parameters for the generated client
			genParams := &company.AddUploadParams{}

			// Call the generated client method
			resp, err := client.CompanyClient.AddUploadWithBody(ctx, companyUUID, genParams, writer.FormDataContentType(), &buf)
			if err != nil {
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to upload file: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
				return &mcp.CallToolResultFor[UploadCreateResult]{
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
				return &mcp.CallToolResultFor[UploadCreateResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[UploadCreateResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully uploaded file\n\nCompany: %s\nFile: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.FileName, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("file_content",
				mcp.Description("Base64 encoded file content"),
				mcp.Required(true),
			),
			mcp.Property("file_name",
				mcp.Description("Name of the file to upload"),
				mcp.Required(true),
			),
			mcp.Property("content_type",
				mcp.Description("MIME type of the file (e.g., image/jpeg, application/pdf)"),
				mcp.Required(true),
			),
			mcp.Property("description",
				mcp.Description("Description of the upload (optional)"),
			),
			mcp.Property("journal_entry_id",
				mcp.Description("Journal entry UUID to attach the upload to (optional)"),
			),
		),
	)

	// Tool to get upload using generated client
	getUploadTool := mcp.NewServerTool[UploadGetParams, UploadGetResult](
		"bokio_uploads_get",
		"Get upload information by ID",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[UploadGetParams]) (*mcp.CallToolResultFor[UploadGetResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[UploadGetResult]{
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
				return &mcp.CallToolResultFor[UploadGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Validate upload ID
			if params.Arguments.UploadID == "" {
				return &mcp.CallToolResultFor[UploadGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "upload_id is required",
						},
					},
				}, nil
			}

			// Parse upload UUID
			uploadUUID, err := uuid.Parse(params.Arguments.UploadID)
			if err != nil {
				return &mcp.CallToolResultFor[UploadGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid upload ID format: %v", err),
						},
					},
				}, nil
			}

			// Call the generated client method
			resp, err := client.CompanyClient.GetUpload(ctx, companyUUID, uploadUUID)
			if err != nil {
				return &mcp.CallToolResultFor[UploadGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to get upload: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[UploadGetResult]{
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
				return &mcp.CallToolResultFor[UploadGetResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to decode response: %v", err),
						},
					},
				}, nil
			}

			// Return success with the actual API response
			return &mcp.CallToolResultFor[UploadGetResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully retrieved upload information\n\nCompany: %s\nUpload ID: %s\nStatus: %d\nResponse: %v", companyIDStr, params.Arguments.UploadID, resp.StatusCode, responseData),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("upload_id",
				mcp.Description("Upload UUID"),
				mcp.Required(true),
			),
		),
	)

	// Tool to download upload using generated client
	downloadUploadTool := mcp.NewServerTool[UploadDownloadParams, UploadDownloadResult](
		"bokio_uploads_download",
		"Download an uploaded file",
		func(ctx context.Context, session *mcp.ServerSession, params *mcp.CallToolParamsFor[UploadDownloadParams]) (*mcp.CallToolResultFor[UploadDownloadResult], error) {
			// Get company ID from params or environment
			companyIDStr := params.Arguments.CompanyID
			if companyIDStr == "" {
				companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
			}

			if companyIDStr == "" {
				return &mcp.CallToolResultFor[UploadDownloadResult]{
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
				return &mcp.CallToolResultFor[UploadDownloadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid company ID format: %v", err),
						},
					},
				}, nil
			}

			// Validate upload ID
			if params.Arguments.UploadID == "" {
				return &mcp.CallToolResultFor[UploadDownloadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: "upload_id is required",
						},
					},
				}, nil
			}

			// Parse upload UUID
			uploadUUID, err := uuid.Parse(params.Arguments.UploadID)
			if err != nil {
				return &mcp.CallToolResultFor[UploadDownloadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid upload ID format: %v", err),
						},
					},
				}, nil
			}

			// Call the generated client method
			resp, err := client.CompanyClient.DownloadUpload(ctx, companyUUID, uploadUUID)
			if err != nil {
				return &mcp.CallToolResultFor[UploadDownloadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to download upload: %v", err),
						},
					},
				}, nil
			}
			defer resp.Body.Close()

			// Handle different response codes
			if resp.StatusCode != http.StatusOK {
				return &mcp.CallToolResultFor[UploadDownloadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
						},
					},
				}, nil
			}

			// Read the file content
			fileContent, err := io.ReadAll(resp.Body)
			if err != nil {
				return &mcp.CallToolResultFor[UploadDownloadResult]{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to read file content: %v", err),
						},
					},
				}, nil
			}

			// Get content type and filename from response headers
			contentType := resp.Header.Get("Content-Type")
			fileName := resp.Header.Get("Content-Disposition")
			if fileName == "" {
				fileName = fmt.Sprintf("upload_%s", params.Arguments.UploadID)
			}

			// Encode file content as base64
			base64Content := base64.StdEncoding.EncodeToString(fileContent)

			// Return success with the file data
			return &mcp.CallToolResultFor[UploadDownloadResult]{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("✅ Successfully downloaded file\n\nCompany: %s\nUpload ID: %s\nContent-Type: %s\nFile Name: %s\nFile Size: %d bytes\nStatus: %d\n\nBase64 Content: %s", companyIDStr, params.Arguments.UploadID, contentType, fileName, len(fileContent), resp.StatusCode, base64Content),
					},
				},
			}, nil
		},
		mcp.Input(
			mcp.Property("company_id",
				mcp.Description("Company UUID (or use BOKIO_COMPANY_ID env var)"),
			),
			mcp.Property("upload_id",
				mcp.Description("Upload UUID"),
				mcp.Required(true),
			),
		),
	)

	// Add all tools to the server
	server.AddTools(listUploadsTool, createUploadTool, getUploadTool, downloadUploadTool)
	return nil
}
