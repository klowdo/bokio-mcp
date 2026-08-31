package tools

import (
	"context"

	"github.com/klowdo/bokio-mcp/bokio"
	"github.com/klowdo/bokio-mcp/bokio/generated/company"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// BankPaymentParams describes a single outgoing bank payment
type BankPaymentParams struct {
	Amount             float64 `json:"amount" jsonschema:"Payment amount"`
	PaymentDate        string  `json:"payment_date" jsonschema:"Payment date in YYYY-MM-DD format"`
	RecipientName      string  `json:"recipient_name" jsonschema:"Name of the payment recipient"`
	Kind               string  `json:"kind" jsonschema:"Recipient kind: bankgiro or transfer"`
	BankgiroNumber     *string `json:"bankgiro_number,omitempty" jsonschema:"Bankgiro number (required when kind is bankgiro)"`
	Message            *string `json:"message,omitempty" jsonschema:"Payment message (bankgiro only; set either message or ocr)"`
	OCR                *string `json:"ocr,omitempty" jsonschema:"OCR reference (bankgiro only; set either message or ocr)"`
	ClearingNumber     *string `json:"clearing_number,omitempty" jsonschema:"Clearing number (required when kind is transfer)"`
	AccountNumber      *string `json:"account_number,omitempty" jsonschema:"Account number (required when kind is transfer)"`
	RecipientReference *string `json:"recipient_reference,omitempty" jsonschema:"Reference for the recipient (transfer only)"`
	OwnNote            *string `json:"own_note,omitempty" jsonschema:"Internal note (optional)"`
}

// BankPaymentCreateParams defines parameters for creating one bank payment
type BankPaymentCreateParams struct {
	CompanyID string            `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Payment   BankPaymentParams `json:"payment" jsonschema:"The payment to create"`
}

// BankPaymentsBulkCreateParams defines parameters for creating several bank payments at once
type BankPaymentsBulkCreateParams struct {
	CompanyID string              `json:"company_id,omitempty" jsonschema:"Company UUID (defaults to BOKIO_COMPANY_ID env var)"`
	Payments  []BankPaymentParams `json:"payments" jsonschema:"The payments to create"`
}

func buildBankPayment(args BankPaymentParams) (company.BankPayment, *mcp.CallToolResult) {
	paymentDate, err := parseAPIDate(args.PaymentDate)
	if err != nil {
		return company.BankPayment{}, toolError("Invalid payment_date: %v", err)
	}

	payment := company.BankPayment{
		Amount:      args.Amount,
		PaymentDate: paymentDate,
		OwnNote:     args.OwnNote,
	}

	switch args.Kind {
	case string(company.BankgiroRecipientKindBankgiro):
		if args.BankgiroNumber == nil {
			return company.BankPayment{}, toolError("bankgiro_number is required when kind is bankgiro")
		}
		if (args.Message == nil) == (args.OCR == nil) {
			return company.BankPayment{}, toolError("Set exactly one of message or ocr for a bankgiro payment")
		}
		if err := payment.RecipientRef.FromBankgiroRecipient(company.BankgiroRecipient{
			Kind:           company.BankgiroRecipientKindBankgiro,
			BankgiroNumber: *args.BankgiroNumber,
			RecipientName:  args.RecipientName,
			Message:        args.Message,
			Ocr:            args.OCR,
		}); err != nil {
			return company.BankPayment{}, toolError("Failed to build bankgiro recipient: %v", err)
		}
	case string(company.TransferRecipientKindTransfer):
		if args.ClearingNumber == nil || args.AccountNumber == nil {
			return company.BankPayment{}, toolError("clearing_number and account_number are required when kind is transfer")
		}
		if err := payment.RecipientRef.FromTransferRecipient(company.TransferRecipient{
			Kind:               company.TransferRecipientKindTransfer,
			ClearingNumber:     *args.ClearingNumber,
			AccountNumber:      *args.AccountNumber,
			RecipientName:      args.RecipientName,
			RecipientReference: args.RecipientReference,
		}); err != nil {
			return company.BankPayment{}, toolError("Failed to build transfer recipient: %v", err)
		}
	default:
		return company.BankPayment{}, toolError("Payment kind must be bankgiro or transfer")
	}

	return payment, nil
}

// RegisterBankPaymentTools registers outgoing bank payment tools
func RegisterBankPaymentTools(server *mcp.Server, client *bokio.AuthClient) error {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_bank_payments_create",
		Description: "Create an outgoing bank payment from the Bokio Business Account",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args BankPaymentCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		payment, errResult := buildBankPayment(args.Payment)
		if errResult != nil {
			return errResult, nil, nil
		}

		resp, err := client.CompanyClient.PostBankPayment(ctx, companyUUID, payment)

		return renderAPIResponse("create bank payment", resp, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "bokio_bank_payments_create_bulk",
		Description: "Create several outgoing bank payments in one request",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, args BankPaymentsBulkCreateParams) (*mcp.CallToolResult, any, error) {
		if client.GetConfig().ReadOnly {
			return readOnlyError(), nil, nil
		}

		companyUUID, errResult := resolveCompanyUUID(args.CompanyID)
		if errResult != nil {
			return errResult, nil, nil
		}

		if len(args.Payments) == 0 {
			return toolError("At least one payment is required"), nil, nil
		}

		payments := make([]company.BankPayment, 0, len(args.Payments))
		for _, paymentArgs := range args.Payments {
			payment, errResult := buildBankPayment(paymentArgs)
			if errResult != nil {
				return errResult, nil, nil
			}
			payments = append(payments, payment)
		}

		resp, err := client.CompanyClient.PostBankPaymentsBulk(ctx, companyUUID, company.BulkCreateBankPaymentsRequest{
			Payments: payments,
		})

		return renderAPIResponse("create bank payments", resp, err)
	})

	return nil
}
