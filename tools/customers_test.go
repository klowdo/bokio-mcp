package tools

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCustomersListParams(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CustomersListParams
		wantErr bool
	}{
		{
			name:  "valid minimal params",
			input: `{"company_id": "test-company-123"}`,
			want: CustomersListParams{
				CompanyID: "test-company-123",
			},
		},
		{
			name:  "valid full params",
			input: `{"company_id": "test-company-123", "page": 2, "page_size": 50, "search": "test"}`,
			want: CustomersListParams{
				CompanyID: "test-company-123",
				Page:      int32Ptr(2),
				PageSize:  int32Ptr(50),
				Search:    stringPtr("test"),
			},
		},
		{
			name:    "missing company_id",
			input:   `{"page": 1}`,
			wantErr: false, // JSON unmarshaling won't fail, but CompanyID will be empty
			want: CustomersListParams{
				Page: int32Ptr(1),
			},
		},
		{
			name:    "invalid JSON",
			input:   `{"company_id": "test", "page":}`,
			wantErr: true,
		},
		{
			name:  "negative page",
			input: `{"company_id": "test", "page": -1}`,
			want: CustomersListParams{
				CompanyID: "test",
				Page:      int32Ptr(-1),
			},
		},
		{
			name:  "large page_size",
			input: `{"company_id": "test", "page_size": 1000}`,
			want: CustomersListParams{
				CompanyID: "test",
				PageSize:  int32Ptr(1000),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CustomersListParams
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCustomersListParamsValidation(t *testing.T) {
	tests := []struct {
		name   string
		params CustomersListParams
		valid  bool
		reason string
	}{
		{
			name: "valid minimal",
			params: CustomersListParams{
				CompanyID: "test-company-123",
			},
			valid: true,
		},
		{
			name: "empty company_id",
			params: CustomersListParams{
				CompanyID: "",
			},
			valid:  false,
			reason: "company_id is required",
		},
		{
			name: "valid with pagination",
			params: CustomersListParams{
				CompanyID: "test-company-123",
				Page:      int32Ptr(1),
				PageSize:  int32Ptr(25),
			},
			valid: true,
		},
		{
			name: "zero page should be invalid in business logic",
			params: CustomersListParams{
				CompanyID: "test-company-123",
				Page:      int32Ptr(0),
			},
			valid:  false,
			reason: "page should be >= 1",
		},
		{
			name: "negative page size",
			params: CustomersListParams{
				CompanyID: "test-company-123",
				PageSize:  int32Ptr(-1),
			},
			valid:  false,
			reason: "page_size should be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCustomersListParams(&tt.params)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.reason)
			}
		})
	}
}

func TestCustomerCreateParams(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CustomerCreateParams
		wantErr bool
	}{
		{
			name:  "valid minimal params",
			input: `{"company_id": "test-company", "name": "Test Customer", "type": "private"}`,
			want: CustomerCreateParams{
				CompanyID: "test-company",
				Name:      "Test Customer",
				Type:      "private",
			},
		},
		{
			name: "valid full params",
			input: `{
				"company_id": "test-company",
				"name": "Test Company Ltd",
				"email": "test@example.com",
				"phone": "+46701234567",
				"organization_number": "556123-4567",
				"vat_number": "SE556123456701",
				"type": "company",
				"payment_terms": 30
			}`,
			want: CustomerCreateParams{
				CompanyID:          "test-company",
				Name:               "Test Company Ltd",
				Email:              stringPtr("test@example.com"),
				Phone:              stringPtr("+46701234567"),
				OrganizationNumber: stringPtr("556123-4567"),
				VatNumber:          stringPtr("SE556123456701"),
				Type:               "company",
				PaymentTerms:       intPtr(30),
			},
		},
		{
			name:    "invalid JSON",
			input:   `{"name": "Test", "type":}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CustomerCreateParams
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCustomerCreateParamsValidation(t *testing.T) {
	tests := []struct {
		name   string
		params CustomerCreateParams
		valid  bool
		reason string
	}{
		{
			name: "valid private customer",
			params: CustomerCreateParams{
				CompanyID: "test-company",
				Name:      "John Doe",
				Type:      "private",
			},
			valid: true,
		},
		{
			name: "valid company customer",
			params: CustomerCreateParams{
				CompanyID: "test-company",
				Name:      "Test Company Ltd",
				Type:      "company",
				Email:     stringPtr("test@company.com"),
			},
			valid: true,
		},
		{
			name: "missing company_id",
			params: CustomerCreateParams{
				Name: "Test Customer",
				Type: "private",
			},
			valid:  false,
			reason: "company_id is required",
		},
		{
			name: "missing name",
			params: CustomerCreateParams{
				CompanyID: "test-company",
				Type:      "private",
			},
			valid:  false,
			reason: "name is required",
		},
		{
			name: "missing type",
			params: CustomerCreateParams{
				CompanyID: "test-company",
				Name:      "Test Customer",
			},
			valid:  false,
			reason: "type is required",
		},
		{
			name: "invalid type",
			params: CustomerCreateParams{
				CompanyID: "test-company",
				Name:      "Test Customer",
				Type:      "invalid",
			},
			valid:  false,
			reason: "type must be 'company' or 'private'",
		},
		{
			name: "negative payment terms",
			params: CustomerCreateParams{
				CompanyID:    "test-company",
				Name:         "Test Customer",
				Type:         "company",
				PaymentTerms: intPtr(-1),
			},
			valid:  false,
			reason: "payment_terms must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCustomerCreateParams(&tt.params)
			if tt.valid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.reason)
			}
		})
	}
}

func TestCustomerUpdateParams(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    CustomerUpdateParams
		wantErr bool
	}{
		{
			name:  "valid minimal params",
			input: `{"company_id": "test-company", "customer_id": "cust-123"}`,
			want: CustomerUpdateParams{
				CompanyID:  "test-company",
				CustomerID: "cust-123",
			},
		},
		{
			name: "valid with updates",
			input: `{
				"company_id": "test-company",
				"customer_id": "cust-123",
				"name": "Updated Name",
				"email": "new@example.com",
				"type": "company"
			}`,
			want: CustomerUpdateParams{
				CompanyID:  "test-company",
				CustomerID: "cust-123",
				Name:       stringPtr("Updated Name"),
				Email:      stringPtr("new@example.com"),
				Type:       stringPtr("company"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got CustomerUpdateParams
			err := json.Unmarshal([]byte(tt.input), &got)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCustomerResultStructures(t *testing.T) {
	// Test that result structures marshal/unmarshal correctly
	tests := []struct {
		name   string
		result interface{}
	}{
		{
			name: "CustomersListResult success",
			result: CustomersListResult{
				Success: true,
				Data:    map[string]string{"test": "data"},
			},
		},
		{
			name: "CustomersListResult error",
			result: CustomersListResult{
				Success: false,
				Error:   "Test error message",
			},
		},
		{
			name: "CustomerCreateResult success",
			result: CustomerCreateResult{
				Success: true,
				Data:    map[string]interface{}{"id": "cust-123", "name": "Test Customer"},
			},
		},
		{
			name: "CustomerGetResult",
			result: CustomerGetResult{
				Success: true,
				Data:    map[string]string{"customer_id": "cust-123"},
			},
		},
		{
			name: "CustomerUpdateResult",
			result: CustomerUpdateResult{
				Success: true,
				Data:    "Updated successfully",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Marshal to JSON
			data, err := json.Marshal(tt.result)
			require.NoError(t, err)

			// Unmarshal back to verify structure
			var unmarshaled map[string]interface{}
			err = json.Unmarshal(data, &unmarshaled)
			require.NoError(t, err)

			// Verify basic structure
			assert.Contains(t, unmarshaled, "success")
		})
	}
}

// Helper functions for creating pointers to primitive types
func stringPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func intPtr(i int) *int {
	return &i
}

// Validation functions that would be used in the actual implementation
// These are examples of validation logic that could be added to the tools

func validateCustomersListParams(params *CustomersListParams) error {
	if params.CompanyID == "" {
		return fmt.Errorf("company_id is required")
	}

	if params.Page != nil && *params.Page <= 0 {
		return fmt.Errorf("page should be >= 1")
	}

	if params.PageSize != nil && *params.PageSize <= 0 {
		return fmt.Errorf("page_size should be positive")
	}

	return nil
}

func validateCustomerCreateParams(params *CustomerCreateParams) error {
	if params.CompanyID == "" {
		return fmt.Errorf("company_id is required")
	}

	if params.Name == "" {
		return fmt.Errorf("name is required")
	}

	if params.Type == "" {
		return fmt.Errorf("type is required")
	}

	if params.Type != "company" && params.Type != "private" {
		return fmt.Errorf("type must be 'company' or 'private'")
	}

	if params.PaymentTerms != nil && *params.PaymentTerms < 0 {
		return fmt.Errorf("payment_terms must be non-negative")
	}

	return nil
}