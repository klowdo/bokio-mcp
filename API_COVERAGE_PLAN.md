# Bokio MCP Server - API Coverage Implementation Plan

## Current State Analysis

### Currently Implemented MCP Tools
- **Only 1 active tool**: `bokio_journal_entries_pure_generated` (GET journal entries only)
- **5 disabled tools**: `auth.go.disabled`, `customers.go.disabled`, `invoices.go.disabled`, `journal.go.disabled`, `uploads.go.disabled`

### Available API Endpoints in OpenAPI Schemas

#### Company API (`company-api.yaml`) - 19 endpoints
1. **Journal Entries** (4 endpoints)
   - `GET /v1/companies/{companyId}/journal-entries` ✅ **IMPLEMENTED**
   - `POST /v1/companies/{companyId}/journal-entries` ❌ **Missing**
   - `GET /v1/companies/{companyId}/journal-entries/{journalEntryId}` ❌ **Missing**
   - `POST /v1/companies/{companyId}/journal-entries/{journalEntryId}/reverse` ❌ **Missing**

2. **File Uploads** (3 endpoints)
   - `POST /v1/companies/{companyId}/uploads` ❌ **Missing**
   - `GET /v1/companies/{companyId}/uploads/{uploadId}` ❌ **Missing**
   - `GET /v1/companies/{companyId}/uploads/{uploadId}/download` ❌ **Missing**

3. **Customers** (4 endpoints)
   - `GET /v1/companies/{companyId}/customers` ❌ **Missing**
   - `POST /v1/companies/{companyId}/customers` ❌ **Missing**
   - `GET /v1/companies/{companyId}/customers/{customerId}` ❌ **Missing**
   - `PUT /v1/companies/{companyId}/customers/{customerId}` ❌ **Missing**

4. **Invoices** (6 endpoints)
   - `GET /v1/companies/{companyId}/invoices` ❌ **Missing**
   - `POST /v1/companies/{companyId}/invoices` ❌ **Missing**
   - `GET /v1/companies/{companyId}/invoices/{invoiceId}` ❌ **Missing**
   - `PUT /v1/companies/{companyId}/invoices/{invoiceId}` ❌ **Missing**
   - `GET /v1/companies/{companyId}/invoices/{invoiceId}/line-items` ❌ **Missing**
   - `POST /v1/companies/{companyId}/invoices/{invoiceId}/line-items` ❌ **Missing**

5. **Invoice Attachments** (3 endpoints)
   - `POST /v1/companies/{companyId}/invoices/{invoiceId}/attachments` ❌ **Missing**
   - `DELETE /v1/companies/{companyId}/invoices/{invoiceId}/attachments/{attachmentId}` ❌ **Missing**
   - `GET /v1/companies/{companyId}/invoices/{invoiceId}/attachments/{attachmentId}/download` ❌ **Missing**

6. **Items/Inventory** (4 endpoints)
   - `GET /v1/companies/{companyId}/items` ❌ **Missing**
   - `POST /v1/companies/{companyId}/items` ❌ **Missing**
   - `GET /v1/companies/{companyId}/items/{itemId}` ❌ **Missing**
   - `PUT /v1/companies/{companyId}/items/{itemId}` ❌ **Missing**

7. **Fiscal Years** (2 endpoints)
   - `GET /v1/companies/{companyId}/fiscal-years` ❌ **Missing**
   - `GET /v1/companies/{companyId}/fiscal-years/{fiscalYearId}` ❌ **Missing**

8. **SIE Files** (1 endpoint)
   - `GET /v1/companies/{companyId}/sie/{fiscalYearId}/download` ❌ **Missing**

#### General API (`general-api.yaml`) - 3 endpoints
1. **OAuth Authorization** (2 endpoints)
   - `POST /token` ❌ **Missing**
   - `GET /authorize` ❌ **Missing**

2. **Connections** (1 endpoint)
   - `GET /v1/connections` ❌ **Missing**

### Coverage Summary
- **Total API endpoints**: 25
- **Currently implemented**: 1 (4%)
- **Missing**: 24 (96%)

---

## Implementation Plan

### Phase 1: Rename and Clean Up Current Implementation
**Priority**: High | **Effort**: Low | **Timeline**: 1 day

#### 1.1 Rename Current Tool ✨
- Change `bokio_journal_entries_pure_generated` → `bokio_journal_entries_list`
- Remove "_pure_generated" suffix from tool names
- Update tool descriptions to be more user-friendly

#### 1.2 Enable Existing Tools
- Review and re-enable disabled tool files
- Migrate them to use generated clients
- Fix any compatibility issues

### Phase 2: Complete Journal Entries Implementation
**Priority**: High | **Effort**: Medium | **Timeline**: 2-3 days

#### 2.1 Journal Entry Operations
- `bokio_journal_entries_list` ✅ **Already implemented**
- `bokio_journal_entries_create` - Create new journal entries
- `bokio_journal_entries_get` - Get specific journal entry by ID
- `bokio_journal_entries_reverse` - Reverse/cancel a journal entry

#### 2.2 Implementation Details
- Use generated `company.Client` for all operations
- Implement proper error handling and validation
- Add support for read-only mode restrictions
- Include comprehensive parameter validation

### Phase 3: Core Business Operations
**Priority**: High | **Effort**: High | **Timeline**: 1-2 weeks

#### 3.1 Customer Management
- `bokio_customers_list` - List all customers with filtering/pagination
- `bokio_customers_create` - Create new customer
- `bokio_customers_get` - Get customer by ID
- `bokio_customers_update` - Update customer information

#### 3.2 Invoice Management
- `bokio_invoices_list` - List invoices with filtering/pagination
- `bokio_invoices_create` - Create new invoice
- `bokio_invoices_get` - Get invoice by ID
- `bokio_invoices_update` - Update invoice
- `bokio_invoices_line_items_list` - List invoice line items
- `bokio_invoices_line_items_create` - Add line items to invoice

#### 3.3 Invoice Attachments
- `bokio_invoice_attachments_upload` - Upload attachment to invoice
- `bokio_invoice_attachments_delete` - Delete invoice attachment
- `bokio_invoice_attachments_download` - Download invoice attachment

### Phase 4: File Management and Inventory
**Priority**: Medium | **Effort**: Medium | **Timeline**: 1 week

#### 4.1 File Upload System
- `bokio_uploads_create` - Upload files to Bokio
- `bokio_uploads_get` - Get upload information
- `bokio_uploads_download` - Download uploaded files

#### 4.2 Inventory Management
- `bokio_items_list` - List inventory items
- `bokio_items_create` - Create new item
- `bokio_items_get` - Get item by ID
- `bokio_items_update` - Update item information

### Phase 5: Financial Reporting and Authentication
**Priority**: Medium | **Effort**: Medium | **Timeline**: 1 week

#### 5.1 Fiscal Years and SIE Export
- `bokio_fiscal_years_list` - List company fiscal years
- `bokio_fiscal_years_get` - Get specific fiscal year
- `bokio_sie_download` - Download SIE (Swedish accounting standard) file

#### 5.2 OAuth and Connection Management
- `bokio_oauth_authorize` - Start OAuth authorization flow
- `bokio_oauth_token` - Exchange codes for tokens
- `bokio_connections_list` - List user connections

### Phase 6: Advanced Features and Optimization
**Priority**: Low | **Effort**: Medium | **Timeline**: 1 week

#### 6.1 Bulk Operations
- `bokio_journal_entries_bulk_create` - Create multiple entries
- `bokio_customers_bulk_import` - Import customer data
- `bokio_invoices_bulk_process` - Process multiple invoices

#### 6.2 Search and Filtering
- Enhanced search capabilities across all resources
- Advanced filtering options
- Date range queries
- Full-text search where supported

#### 6.3 Webhooks and Real-time Updates
- `bokio_webhooks_list` - List configured webhooks
- `bokio_webhooks_create` - Create webhook subscriptions
- Real-time notifications for data changes

---

## Technical Implementation Guidelines

### 1. Tool Naming Convention ✨
**OLD**: `bokio_journal_entries_pure_generated`
**NEW**: `bokio_journal_entries_list`

```
bokio_{resource}_{action}
```
**Examples:**
- `bokio_journal_entries_list`
- `bokio_customers_create`
- `bokio_invoices_get`
- `bokio_uploads_download`

### 2. File Organization
```
tools/
├── journal_entries.go     # All journal entry operations
├── customers.go           # Customer management operations
├── invoices.go           # Invoice operations
├── uploads.go            # File upload operations
├── items.go              # Inventory items
├── fiscal_years.go       # Fiscal year operations
├── auth.go               # OAuth and authentication
└── connections.go        # Connection management
```

### 3. Generated Client Usage
- Use `bokio/generated/company.Client` for company API
- Use `bokio/generated/general.Client` for general API
- Implement proper error handling and response parsing
- Respect read-only mode for write operations

### 4. Parameter Validation
- UUID validation for company/resource IDs
- Date format validation (ISO 8601)
- Required parameter checking
- Enum value validation

### 5. Error Handling Standards
```go
// Standard error response format
type ToolResult struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
    Code    string      `json:"code,omitempty"`
}
```

### 6. Read-Only Mode Support
```go
if client.GetConfig().ReadOnly {
    return &mcp.CallToolResultFor[Result]{
        Content: []mcp.Content{
            &mcp.TextContent{
                Text: "Write operations not allowed in read-only mode",
            },
        },
    }, nil
}
```

---

## Testing Strategy

### 1. Unit Tests
- Test each MCP tool individually
- Mock generated clients for isolated testing
- Validate parameter parsing and response formatting

### 2. Integration Tests
- Test against real Bokio API (with test company)
- Validate generated client integration
- Test authentication flows

### 3. End-to-End Tests
- Test complete workflows (create customer → create invoice → add items)
- Test error scenarios and edge cases
- Validate read-only mode restrictions

---

## Success Metrics

### 1. API Coverage
- **Target**: 100% of available API endpoints (25 tools total)
- **Current**: 4% (1 tool)
- **Phase milestones**: 25%, 50%, 75%, 100%

### 2. Tool Quality
- Error rate < 1% for successful authentication scenarios
- Average response time < 2 seconds
- User satisfaction score > 4.5/5

### 3. Usage Analytics
- Track most/least used tools
- Identify common error patterns
- Monitor authentication success rates

---

## Priority Action Items 🚀

### Immediate (This Week)
1. **Rename tool**: `bokio_journal_entries_pure_generated` → `bokio_journal_entries_list`
2. **Complete journal entries**: Add create, get, and reverse operations
3. **Enable customers.go**: Migrate to generated client and re-enable

### Short Term (Next 2 Weeks)
4. **Invoice management**: Full CRUD operations for invoices
5. **Customer management**: Full CRUD operations for customers
6. **File uploads**: Basic upload and download functionality

### Medium Term (Next Month)
7. **Inventory management**: Items CRUD operations
8. **Financial reporting**: Fiscal years and SIE exports
9. **Authentication flows**: OAuth and connection management

### Long Term (Next Quarter)
10. **Advanced features**: Bulk operations, webhooks, advanced search
11. **Performance optimization**: Caching, rate limiting, monitoring
12. **Documentation**: Complete API reference and usage examples

---

*This implementation plan provides a comprehensive roadmap for achieving full Bokio API coverage through MCP tools, starting with the critical rename task and building towards 100% API coverage.*
