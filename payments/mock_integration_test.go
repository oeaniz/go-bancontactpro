//go:build integration

package payments

import (
	"context"
	"os"
	"testing"
	"time"
)

const defaultMockBaseURL = "https://docs.bancontactpro.com/_mock/apis/merchant-payment.openapi"

func TestBancontactPaymentAPIMock(t *testing.T) {
	baseURL := os.Getenv("BANCONTACT_MOCK_URL")
	if baseURL == "" {
		baseURL = defaultMockBaseURL
	}
	client, err := NewClient("test-api-key", WithBaseURL(baseURL))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	const paymentID = "5f91483d-78a7-4914-bc6f="

	t.Run("create payment", func(t *testing.T) {
		response, err := client.CreatePayment(ctx, CreatePaymentRequest{
			Amount: 1250, Currency: CurrencyEUR,
			Reference: "mock-integration", Description: "SDK mock integration test",
		})
		if err != nil {
			t.Fatal(err)
		}
		if response.PaymentID == "" || response.Status != PaymentStatusPending ||
			response.Amount != 1250 || response.Currency != CurrencyEUR {
			t.Fatalf("unexpected create response: %+v", response)
		}
		if response.Links.DeepLink == nil || response.Links.DeepLink.Href == "" {
			t.Fatalf("create response omitted deeplink: %+v", response.Links)
		}
	})

	t.Run("create static QR payment", func(t *testing.T) {
		response, err := client.CreateStaticQRPayment(ctx, CreateStaticQRPaymentRequest{
			Amount: 1250, Currency: CurrencyEUR, POSID: "POS1",
		})
		if err != nil {
			t.Fatal(err)
		}
		if response.PaymentID == "" || response.Amount != 1250 {
			t.Fatalf("unexpected static QR response: %+v", response)
		}
	})

	t.Run("get payment", func(t *testing.T) {
		response, err := client.GetPayment(ctx, paymentID)
		if err != nil {
			t.Fatal(err)
		}
		if response.PaymentID == "" || response.Status == "" || response.Links.DeepLink == nil {
			t.Fatalf("unexpected payment response: %+v", response)
		}
	})

	t.Run("search payments", func(t *testing.T) {
		from := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 11, 21, 0, 0, 0, 0, time.UTC)
		response, err := client.SearchPayments(ctx, SearchPaymentsRequest{
			Page: 0, Size: 10, From: &from, To: &to, Reference: "mock-integration",
		})
		if err != nil {
			t.Fatal(err)
		}
		if response.Details == nil {
			t.Fatalf("search response omitted details: %+v", response)
		}
		if len(response.Details) > 0 && response.Details[0].Links.DeepLink == nil {
			t.Fatalf("search response returned invalid payment links: %+v", response.Details[0].Links)
		}
	})

	t.Run("get debtor refund IBAN", func(t *testing.T) {
		response, err := client.GetRefundIBAN(ctx, paymentID)
		if err != nil {
			t.Fatal(err)
		}
		if response.IBAN == "" {
			t.Fatal("refund IBAN response is empty")
		}
	})

	t.Run("acknowledge payment", func(t *testing.T) {
		err := client.AcknowledgePayment(ctx, paymentID, AcknowledgePaymentRequest{
			Currency: CurrencyEUR, Amount: 1250, Reference: "mock-integration",
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("cancel payment", func(t *testing.T) {
		if err := client.CancelPayment(ctx, paymentID); err != nil {
			t.Fatal(err)
		}
	})
}
