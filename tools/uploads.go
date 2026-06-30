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
	CompanyID string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	Page      *int32 `json:"page,omitempty" jsonschema:"Page number (optional)"`
	PageSize  *int32 `json:"page_size,omitempty" jsonschema:"Items per page (optional)"`
}

// UploadListResult defines the result for listing uploads
type UploadListResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// UploadCreateParams defines parameters for creating an upload
type UploadCreateParams struct {
	CompanyID      string  `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	FileContent    string  `json:"file_content" jsonschema:"Base64 encoded file content"`
	FileName       string  `json:"file_name" jsonschema:"Name of the file to upload"`
	ContentType    string  `json:"content_type" jsonschema:"MIME type of the file (e.g., image/jpeg, application/pdf)"`
	Description    *string `json:"description,omitempty" jsonschema:"Description of the upload (optional)"`
	JournalEntryID *string `json:"journal_entry_id,omitempty" jsonschema:"Journal entry UUID to attach the upload to (optional)"`
}

// UploadCreateResult defines the result for creating an upload
type UploadCreateResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// UploadGetParams defines parameters for getting an upload
type UploadGetParams struct {
	CompanyID string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	UploadID  string `json:"upload_id" jsonschema:"Upload UUID"`
}

// UploadGetResult defines the result for getting an upload
type UploadGetResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// UploadDownloadParams defines parameters for downloading an upload
type UploadDownloadParams struct {
	CompanyID string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	UploadID  string `json:"upload_id" jsonschema:"Upload UUID"`
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
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_uploads_list",
		Description: "List uploads for a company with optional pagination",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UploadListParams) (*mcp.CallToolResult, UploadListResult, error) {
		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, UploadListResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, UploadListResult{}, nil
		}

		genParams := &company.GetUploadsParams{
			Page:     args.Page,
			PageSize: args.PageSize,
		}

		resp, err := client.CompanyClient.GetUploads(ctx, companyUUID, genParams)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to list uploads: %v", err),
					},
				},
			}, UploadListResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, UploadListResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, UploadListResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved uploads list\n\nCompany: %s\nStatus: %d\nResponse: %v", companyIDStr, resp.StatusCode, responseData),
				},
			},
		}, UploadListResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_uploads_create",
		Description: "Upload a file to Bokio",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UploadCreateParams) (*mcp.CallToolResult, UploadCreateResult, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Upload creation not allowed in read-only mode",
					},
				},
			}, UploadCreateResult{}, nil
		}

		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, UploadCreateResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, UploadCreateResult{}, nil
		}

		if args.FileContent == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "file_content is required (base64 encoded file)",
					},
				},
			}, UploadCreateResult{}, nil
		}

		if args.FileName == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "file_name is required",
					},
				},
			}, UploadCreateResult{}, nil
		}

		if args.ContentType == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "content_type is required",
					},
				},
			}, UploadCreateResult{}, nil
		}

		fileData, err := base64.StdEncoding.DecodeString(args.FileContent)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid base64 file content: %v", err),
					},
				},
			}, UploadCreateResult{}, nil
		}

		var journalEntryUUID *openapi_types.UUID
		if args.JournalEntryID != nil && *args.JournalEntryID != "" {
			journalUUID, err := uuid.Parse(*args.JournalEntryID)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid journal entry ID format: %v", err),
						},
					},
				}, UploadCreateResult{}, nil
			}
			journalEntryUUID = &journalUUID
		}

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		fileWriter, err := writer.CreateFormFile("file", args.FileName)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to create form file: %v", err),
					},
				},
			}, UploadCreateResult{}, nil
		}
		_, err = fileWriter.Write(fileData)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to write file data: %v", err),
					},
				},
			}, UploadCreateResult{}, nil
		}

		if args.Description != nil {
			err = writer.WriteField("description", *args.Description)
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to write description field: %v", err),
						},
					},
				}, UploadCreateResult{}, nil
			}
		}

		if journalEntryUUID != nil {
			err = writer.WriteField("journalEntryId", journalEntryUUID.String())
			if err != nil {
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to write journal entry ID field: %v", err),
						},
					},
				}, UploadCreateResult{}, nil
			}
		}

		err = writer.Close()
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to close multipart writer: %v", err),
					},
				},
			}, UploadCreateResult{}, nil
		}

		genParams := &company.AddUploadParams{}

		resp, err := client.CompanyClient.AddUploadWithBody(ctx, companyUUID, genParams, writer.FormDataContentType(), &buf)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to upload file: %v", err),
					},
				},
			}, UploadCreateResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, UploadCreateResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, UploadCreateResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully uploaded file\n\nCompany: %s\nFile: %s\nStatus: %d\nResponse: %v", companyIDStr, args.FileName, resp.StatusCode, responseData),
				},
			},
		}, UploadCreateResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_uploads_get",
		Description: "Get upload information by ID",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UploadGetParams) (*mcp.CallToolResult, UploadGetResult, error) {
		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, UploadGetResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, UploadGetResult{}, nil
		}

		if args.UploadID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "upload_id is required",
					},
				},
			}, UploadGetResult{}, nil
		}

		uploadUUID, err := uuid.Parse(args.UploadID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid upload ID format: %v", err),
					},
				},
			}, UploadGetResult{}, nil
		}

		resp, err := client.CompanyClient.GetUpload(ctx, companyUUID, uploadUUID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to get upload: %v", err),
					},
				},
			}, UploadGetResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, UploadGetResult{}, nil
		}

		var responseData interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to decode response: %v", err),
					},
				},
			}, UploadGetResult{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully retrieved upload information\n\nCompany: %s\nUpload ID: %s\nStatus: %d\nResponse: %v", companyIDStr, args.UploadID, resp.StatusCode, responseData),
				},
			},
		}, UploadGetResult{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_uploads_download",
		Description: "Download an uploaded file",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UploadDownloadParams) (*mcp.CallToolResult, UploadDownloadResult, error) {
		companyIDStr := args.CompanyID
		if companyIDStr == "" {
			companyIDStr = os.Getenv("BOKIO_COMPANY_ID")
		}

		if companyIDStr == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Company ID is required (provide in company_id parameter or BOKIO_COMPANY_ID env var)",
					},
				},
			}, UploadDownloadResult{}, nil
		}

		companyUUID, err := uuid.Parse(companyIDStr)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid company ID format: %v", err),
					},
				},
			}, UploadDownloadResult{}, nil
		}

		if args.UploadID == "" {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "upload_id is required",
					},
				},
			}, UploadDownloadResult{}, nil
		}

		uploadUUID, err := uuid.Parse(args.UploadID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid upload ID format: %v", err),
					},
				},
			}, UploadDownloadResult{}, nil
		}

		resp, err := client.CompanyClient.DownloadUpload(ctx, companyUUID, uploadUUID)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to download upload: %v", err),
					},
				},
			}, UploadDownloadResult{}, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("API returned status %d", resp.StatusCode),
					},
				},
			}, UploadDownloadResult{}, nil
		}

		fileContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to read file content: %v", err),
					},
				},
			}, UploadDownloadResult{}, nil
		}

		contentType := resp.Header.Get("Content-Type")
		fileName := resp.Header.Get("Content-Disposition")
		if fileName == "" {
			fileName = fmt.Sprintf("upload_%s", args.UploadID)
		}

		base64Content := base64.StdEncoding.EncodeToString(fileContent)

		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{
					Text: fmt.Sprintf("✅ Successfully downloaded file\n\nCompany: %s\nUpload ID: %s\nContent-Type: %s\nFile Name: %s\nFile Size: %d bytes\nStatus: %d\n\nBase64 Content: %s", companyIDStr, args.UploadID, contentType, fileName, len(fileContent), resp.StatusCode, base64Content),
				},
			},
		}, UploadDownloadResult{}, nil
	})

	return nil
}
