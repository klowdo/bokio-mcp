// Package bokio contains type definitions for the Bokio API
package bokio

import (
	"time"
)

// PagedResponse represents a paginated response from the Bokio API
type PagedResponse struct {
	TotalItems  int32 `json:"totalItems"`
	TotalPages  int32 `json:"totalPages"`
	CurrentPage int32 `json:"currentPage"`
}

// Connection represents a connection to a tenant in the Bokio API
type Connection struct {
	ID       string `json:"id"`
	TenantID string `json:"tenantId"`
	Type     string `json:"type"`
}

// ConnectionsResponse represents a paginated list of connections
type ConnectionsResponse struct {
	PagedResponse
	Items []Connection `json:"items"`
}

// JournalEntryItem represents a single item in a journal entry
type JournalEntryItem struct {
	ID      int64   `json:"id,omitempty"`      // Read-only
	Debit   float64 `json:"debit"`
	Credit  float64 `json:"credit"`
	Account int32   `json:"account"`
}

// JournalEntry represents a journal entry in the accounting system
type JournalEntry struct {
	ID                        string               `json:"id,omitempty"`                       // Read-only
	Title                     string               `json:"title"`
	JournalEntryNumber        string               `json:"journalEntryNumber,omitempty"`       // Read-only
	Date                      string               `json:"date"`                               // Format: date (YYYY-MM-DD)
	Items                     []JournalEntryItem   `json:"items"`
	ReversingJournalEntryID   *string              `json:"reversingJournalEntryId,omitempty"`   // Read-only, nullable
	ReversedByJournalEntryID  *string              `json:"reversedByJournalEntryId,omitempty"`  // Read-only, nullable
}

// JournalEntriesResponse represents a paginated list of journal entries
type JournalEntriesResponse struct {
	PagedResponse
	Items []JournalEntry `json:"items"`
}

// Address represents a postal address
type Address struct {
	Line1      string  `json:"line1"`
	Line2      *string `json:"line2,omitempty"`      // nullable
	City       string  `json:"city"`
	PostalCode string  `json:"postalCode"`
	Country    string  `json:"country"`              // ISO 3166-1 alpha-2 country code
}

// CustomerType represents the type of customer
type CustomerType string

const (
	CustomerTypeCompany CustomerType = "company"
	CustomerTypePrivate CustomerType = "private"
)

// Customer represents a customer in the system
type Customer struct {
	ID              string        `json:"id,omitempty"`                 // Read-only
	Name            string        `json:"name"`
	Type            CustomerType  `json:"type"`
	VatNumber       string        `json:"vatNumber,omitempty"`
	OrgNumber       string        `json:"orgNumber,omitempty"`
	PaymentTerms    string        `json:"paymentTerms,omitempty"`
	Email           string        `json:"email,omitempty"`
	Phone           string        `json:"phone,omitempty"`
	Address         *Address      `json:"address,omitempty"`
	CreatedAt       *time.Time    `json:"createdAt,omitempty"`          // Read-only
	UpdatedAt       *time.Time    `json:"updatedAt,omitempty"`          // Read-only
}

// CustomersResponse represents a paginated list of customers
type CustomersResponse struct {
	PagedResponse
	Items []Customer `json:"items"`
}

// Upload represents a file upload
type Upload struct {
	ID              string  `json:"id,omitempty"`                    // Read-only
	Description     string  `json:"description"`
	ContentType     string  `json:"contentType"`
	JournalEntryID  *string `json:"journalEntryId,omitempty"`        // nullable
	CreatedAt       *time.Time `json:"createdAt,omitempty"`          // Read-only
}

// UploadsResponse represents a paginated list of uploads
type UploadsResponse struct {
	PagedResponse
	Items []Upload `json:"items"`
}

// Item represents an inventory or service item
type Item struct {
	ID            string     `json:"id,omitempty"`                    // Read-only
	Name          string     `json:"name"`
	Description   string     `json:"description,omitempty"`
	Price         float64    `json:"price"`
	Unit          string     `json:"unit,omitempty"`
	Account       int32      `json:"account,omitempty"`
	VatRate       float64    `json:"vatRate,omitempty"`
	Active        bool       `json:"active"`
	CreatedAt     *time.Time `json:"createdAt,omitempty"`             // Read-only
	UpdatedAt     *time.Time `json:"updatedAt,omitempty"`             // Read-only
}

// ItemsResponse represents a paginated list of items
type ItemsResponse struct {
	PagedResponse
	Items []Item `json:"items"`
}

// InvoiceStatus represents the status of an invoice
type InvoiceStatus string

const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

// InvoiceItem represents an item on an invoice
type InvoiceItem struct {
	ID          string  `json:"id,omitempty"`               // Read-only
	ItemID      *string `json:"itemId,omitempty"`           // Reference to Item
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Quantity    float64 `json:"quantity"`
	Price       float64 `json:"price"`
	Unit        string  `json:"unit,omitempty"`
	Account     int32   `json:"account,omitempty"`
	VatRate     float64 `json:"vatRate,omitempty"`
	Total       float64 `json:"total,omitempty"`            // Read-only, calculated
}

// Invoice represents an invoice
type Invoice struct {
	ID             string          `json:"id,omitempty"`                    // Read-only
	InvoiceNumber  string          `json:"invoiceNumber,omitempty"`         // Read-only
	CustomerID     string          `json:"customerId"`                      // Reference to Customer
	Customer       *Customer       `json:"customer,omitempty"`              // Read-only, populated when requested
	Status         InvoiceStatus   `json:"status,omitempty"`                // Read-only
	Date           string          `json:"date"`                            // Format: date (YYYY-MM-DD)
	DueDate        string          `json:"dueDate"`                         // Format: date (YYYY-MM-DD)
	Items          []InvoiceItem   `json:"items"`
	Notes          string          `json:"notes,omitempty"`
	PaymentTerms   string          `json:"paymentTerms,omitempty"`
	Currency       string          `json:"currency,omitempty"`
	Subtotal       float64         `json:"subtotal,omitempty"`              // Read-only, calculated
	VatAmount      float64         `json:"vatAmount,omitempty"`             // Read-only, calculated
	Total          float64         `json:"total,omitempty"`                 // Read-only, calculated
	CreatedAt      *time.Time      `json:"createdAt,omitempty"`             // Read-only
	UpdatedAt      *time.Time      `json:"updatedAt,omitempty"`             // Read-only
}

// InvoicesResponse represents a paginated list of invoices
type InvoicesResponse struct {
	PagedResponse
	Items []Invoice `json:"items"`
}

// FiscalYear represents a fiscal year
type FiscalYear struct {
	ID        string     `json:"id,omitempty"`                    // Read-only
	StartDate string     `json:"startDate"`                       // Format: date (YYYY-MM-DD)
	EndDate   string     `json:"endDate"`                         // Format: date (YYYY-MM-DD)
	Status    string     `json:"status,omitempty"`                // Read-only
	CreatedAt *time.Time `json:"createdAt,omitempty"`             // Read-only
}

// FiscalYearsResponse represents a paginated list of fiscal years
type FiscalYearsResponse struct {
	PagedResponse
	Items []FiscalYear `json:"items"`
}

// APIErrorDetail represents detailed error information
type APIErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// APIErrorResponse represents the standard API error response structure
type APIErrorResponse struct {
	Code         string           `json:"code"`
	Message      string           `json:"message"`
	BokioErrorID string           `json:"bokioErrorId,omitempty"`
	Errors       []APIErrorDetail `json:"errors,omitempty"`
}

// QueryParams represents common query parameters for API requests
type QueryParams struct {
	Page     int32  `json:"page,omitempty"`
	PageSize int32  `json:"pageSize,omitempty"`
	Query    string `json:"query,omitempty"`
}

// DefaultQueryParams returns default query parameters
func DefaultQueryParams() QueryParams {
	return QueryParams{
		Page:     1,
		PageSize: 25,
	}
}

// TenantType represents the type of tenant
type TenantType string

const (
	TenantTypeGeneral TenantType = "general"
	TenantTypeCompany TenantType = "company"
)

// CompanyInfo represents basic company information
type CompanyInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	OrgNumber string `json:"orgNumber,omitempty"`
	VatNumber string `json:"vatNumber,omitempty"`
}

// CreateJournalEntryRequest represents the request to create a journal entry
type CreateJournalEntryRequest struct {
	Title string               `json:"title"`
	Date  string               `json:"date"`
	Items []JournalEntryItem   `json:"items"`
}

// CreateCustomerRequest represents the request to create a customer
type CreateCustomerRequest struct {
	Name         string       `json:"name"`
	Type         CustomerType `json:"type"`
	VatNumber    string       `json:"vatNumber,omitempty"`
	OrgNumber    string       `json:"orgNumber,omitempty"`
	PaymentTerms string       `json:"paymentTerms,omitempty"`
	Email        string       `json:"email,omitempty"`
	Phone        string       `json:"phone,omitempty"`
	Address      *Address     `json:"address,omitempty"`
}

// UpdateCustomerRequest represents the request to update a customer
type UpdateCustomerRequest struct {
	Name         *string      `json:"name,omitempty"`
	VatNumber    *string      `json:"vatNumber,omitempty"`
	OrgNumber    *string      `json:"orgNumber,omitempty"`
	PaymentTerms *string      `json:"paymentTerms,omitempty"`
	Email        *string      `json:"email,omitempty"`
	Phone        *string      `json:"phone,omitempty"`
	Address      *Address     `json:"address,omitempty"`
}

// CreateInvoiceRequest represents the request to create an invoice
type CreateInvoiceRequest struct {
	CustomerID   string        `json:"customerId"`
	Date         string        `json:"date"`
	DueDate      string        `json:"dueDate"`
	Items        []InvoiceItem `json:"items"`
	Notes        string        `json:"notes,omitempty"`
	PaymentTerms string        `json:"paymentTerms,omitempty"`
	Currency     string        `json:"currency,omitempty"`
}

// UpdateInvoiceRequest represents the request to update an invoice
type UpdateInvoiceRequest struct {
	CustomerID   *string       `json:"customerId,omitempty"`
	Date         *string       `json:"date,omitempty"`
	DueDate      *string       `json:"dueDate,omitempty"`
	Items        []InvoiceItem `json:"items,omitempty"`
	Notes        *string       `json:"notes,omitempty"`
	PaymentTerms *string       `json:"paymentTerms,omitempty"`
	Currency     *string       `json:"currency,omitempty"`
}

// CreateItemRequest represents the request to create an item
type CreateItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description,omitempty"`
	Price       float64 `json:"price"`
	Unit        string  `json:"unit,omitempty"`
	Account     int32   `json:"account,omitempty"`
	VatRate     float64 `json:"vatRate,omitempty"`
	Active      bool    `json:"active"`
}

// UpdateItemRequest represents the request to update an item
type UpdateItemRequest struct {
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Unit        *string  `json:"unit,omitempty"`
	Account     *int32   `json:"account,omitempty"`
	VatRate     *float64 `json:"vatRate,omitempty"`
	Active      *bool    `json:"active,omitempty"`
}

// UploadFileRequest represents the request to upload a file
type UploadFileRequest struct {
	Description    string  `json:"description"`
	JournalEntryID *string `json:"journalEntryId,omitempty"`
	FileContent    []byte  `json:"-"` // File content, not JSON serialized
	FileName       string  `json:"-"` // File name, not JSON serialized
	ContentType    string  `json:"-"` // Content type, not JSON serialized
}

// FilterOperator represents filter operators for API queries
type FilterOperator string

const (
	FilterOperatorEquals              FilterOperator = "="
	FilterOperatorNotEquals           FilterOperator = "!="
	FilterOperatorGreaterThan         FilterOperator = ">"
	FilterOperatorGreaterThanOrEqual  FilterOperator = ">="
	FilterOperatorLessThan            FilterOperator = "<"
	FilterOperatorLessThanOrEqual     FilterOperator = "<="
	FilterOperatorContains            FilterOperator = "~"
	FilterOperatorNotContains         FilterOperator = "!~"
	FilterOperatorStartsWith          FilterOperator = "^"
	FilterOperatorEndsWith            FilterOperator = "$"
	FilterOperatorIn                  FilterOperator = "@"
	FilterOperatorNotIn               FilterOperator = "!@"
)

// QueryBuilder helps build query strings for API requests
type QueryBuilder struct {
	filters []string
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		filters: make([]string, 0),
	}
}

// AddFilter adds a filter to the query
func (qb *QueryBuilder) AddFilter(field string, operator FilterOperator, value string) *QueryBuilder {
	qb.filters = append(qb.filters, field+string(operator)+value)
	return qb
}

// Build returns the query string
func (qb *QueryBuilder) Build() string {
	if len(qb.filters) == 0 {
		return ""
	}
	result := ""
	for i, filter := range qb.filters {
		if i > 0 {
			result += " AND "
		}
		result += filter
	}
	return result
}

// Reset clears all filters
func (qb *QueryBuilder) Reset() *QueryBuilder {
	qb.filters = qb.filters[:0]
	return qb
}

// SIEFile represents a SIE (Standard Import Export) file for Swedish accounting
type SIEFile struct {
	ID         string     `json:"id,omitempty"`                    // Read-only
	FiscalYear string     `json:"fiscalYear"`
	FileType   string     `json:"fileType,omitempty"`              // Read-only
	CreatedAt  *time.Time `json:"createdAt,omitempty"`             // Read-only
}

// SIEFilesResponse represents a paginated list of SIE files
type SIEFilesResponse struct {
	PagedResponse
	Items []SIEFile `json:"items"`
}

// Account represents a chart of accounts entry
type Account struct {
	Number      int32  `json:"number"`
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Description string `json:"description,omitempty"`
	Active      bool   `json:"active"`
}

// AccountsResponse represents a list of accounts
type AccountsResponse struct {
	Items []Account `json:"items"`
}