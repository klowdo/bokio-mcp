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

// UploadCreateParams defines parameters for creating an upload
type UploadCreateParams struct {
	CompanyID      string  `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	FileContent    string  `json:"file_content" jsonschema:"Base64 encoded file content"`
	FileName       string  `json:"file_name" jsonschema:"Name of the file to upload"`
	ContentType    string  `json:"content_type" jsonschema:"MIME type of the file (e.g., image/jpeg, application/pdf)"`
	Description    *string `json:"description,omitempty" jsonschema:"Description of the upload (optional)"`
	JournalEntryID *string `json:"journal_entry_id,omitempty" jsonschema:"Journal entry UUID to attach the upload to (optional)"`
}

// UploadGetParams defines parameters for getting an upload
type UploadGetParams struct {
	CompanyID string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	UploadID  string `json:"upload_id" jsonschema:"Upload UUID"`
}

// UploadDownloadParams defines parameters for downloading an upload
type UploadDownloadParams struct {
	CompanyID string `json:"company_id" jsonschema:"Company UUID (or use BOKIO_COMPANY_ID env var)"`
	UploadID  string `json:"upload_id" jsonschema:"Upload UUID"`
}

// RegisterUploadTools registers upload tools using ONLY generated API clients
func RegisterUploadTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_uploads_list",
		Description: "List uploads for a company with optional pagination",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UploadListParams) (*mcp.CallToolResult, any, error) {
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

		genParams := &company.GetUploadsParams{
			Page:     args.Page,
			PageSize: args.PageSize,
		}

		resp, err := client.CompanyClient.GetUploads(ctx, companyUUID, genParams)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to list uploads: %v", err),
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
					Text: fmt.Sprintf("✅ Successfully retrieved uploads list\n\nCompany: %s\nStatus: %d\nResponse: %s", companyIDStr, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_uploads_create",
		Description: "Upload a file to Bokio",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UploadCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "Upload creation not allowed in read-only mode",
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

		if args.FileContent == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "file_content is required (base64 encoded file)",
					},
				},
			}, nil, nil
		}

		if args.FileName == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "file_name is required",
					},
				},
			}, nil, nil
		}

		if args.ContentType == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "content_type is required",
					},
				},
			}, nil, nil
		}

		fileData, err := base64.StdEncoding.DecodeString(args.FileContent)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid base64 file content: %v", err),
					},
				},
			}, nil, nil
		}

		var journalEntryUUID *openapi_types.UUID
		if args.JournalEntryID != nil && *args.JournalEntryID != "" {
			journalUUID, err := uuid.Parse(*args.JournalEntryID)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Invalid journal entry ID format: %v", err),
						},
					},
				}, nil, nil
			}
			journalEntryUUID = &journalUUID
		}

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		fileWriter, err := writer.CreateFormFile("file", args.FileName)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to create form file: %v", err),
					},
				},
			}, nil, nil
		}
		_, err = fileWriter.Write(fileData)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to write file data: %v", err),
					},
				},
			}, nil, nil
		}

		if args.Description != nil {
			err = writer.WriteField("description", *args.Description)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to write description field: %v", err),
						},
					},
				}, nil, nil
			}
		}

		if journalEntryUUID != nil {
			err = writer.WriteField("journalEntryId", journalEntryUUID.String())
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Failed to write journal entry ID field: %v", err),
						},
					},
				}, nil, nil
			}
		}

		err = writer.Close()
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to close multipart writer: %v", err),
					},
				},
			}, nil, nil
		}

		genParams := &company.AddUploadParams{}

		resp, err := client.CompanyClient.AddUploadWithBody(ctx, companyUUID, genParams, writer.FormDataContentType(), &buf)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to upload file: %v", err),
					},
				},
			}, nil, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
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
					Text: fmt.Sprintf("✅ Successfully uploaded file\n\nCompany: %s\nFile: %s\nStatus: %d\nResponse: %s", companyIDStr, args.FileName, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_uploads_get",
		Description: "Get upload information by ID",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UploadGetParams) (*mcp.CallToolResult, any, error) {
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

		if args.UploadID == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "upload_id is required",
					},
				},
			}, nil, nil
		}

		uploadUUID, err := uuid.Parse(args.UploadID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid upload ID format: %v", err),
					},
				},
			}, nil, nil
		}

		resp, err := client.CompanyClient.GetUpload(ctx, companyUUID, uploadUUID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to get upload: %v", err),
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
					Text: fmt.Sprintf("✅ Successfully retrieved upload information\n\nCompany: %s\nUpload ID: %s\nStatus: %d\nResponse: %s", companyIDStr, args.UploadID, resp.StatusCode, responseData),
				},
			},
		}, nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_uploads_download",
		Description: "Download an uploaded file",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args UploadDownloadParams) (*mcp.CallToolResult, any, error) {
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

		if args.UploadID == "" {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: "upload_id is required",
					},
				},
			}, nil, nil
		}

		uploadUUID, err := uuid.Parse(args.UploadID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Invalid upload ID format: %v", err),
					},
				},
			}, nil, nil
		}

		resp, err := client.CompanyClient.DownloadUpload(ctx, companyUUID, uploadUUID)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to download upload: %v", err),
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

		fileContent, err := io.ReadAll(resp.Body)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: fmt.Sprintf("Failed to read file content: %v", err),
					},
				},
			}, nil, nil
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
		}, nil, nil
	})

	return nil
}
