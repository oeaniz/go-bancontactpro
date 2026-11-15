package payments

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientDecodesResponseAndSendsAPIKey(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v3/payments" {
			t.Errorf("request = %s %s, want POST /v3/payments", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "test-api-key" {
			t.Errorf("Authorization = %q, want test-api-key", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		var request CreatePaymentRequest
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		} else if err := json.Unmarshal(body, &request); err != nil {
			t.Errorf("decode request body: %v", err)
		} else if request.Amount != 1250 || request.Currency != CurrencyEUR {
			t.Errorf("request body = %+v, want amount 1250 EUR", request)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{
			"paymentId":"payment-123",
			"status":"PENDING",
			"createdAt":"2026-11-21T10:00:00Z",
			"expiresAt":"2026-11-21T10:15:00Z",
			"amount":1250,
			"currency":"EUR",
			"creditor":{"profileId":"profile-1"},
			"_links":{"deeplink":"https://pay.bancontact.net/pay/2/payment-123"}
		}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	client.httpClient = server.Client()
	response, err := client.CreatePayment(context.Background(), CreatePaymentRequest{
		Amount: 1250, Currency: CurrencyEUR,
	})
	if err != nil {
		t.Fatalf("CreatePayment: %v", err)
	}
	if response.PaymentID != "payment-123" || response.Amount != 1250 ||
		response.Links.DeepLink == nil ||
		response.Links.DeepLink.Href != "https://pay.bancontact.net/pay/2/payment-123" {
		t.Fatalf("unexpected decoded response: %+v", response)
	}
}

func TestLinkAcceptsStringAndObjectForms(t *testing.T) {
	for _, test := range []struct {
		name string
		json string
		want string
	}{
		{name: "string", json: `"https://example.com/pay"`, want: "https://example.com/pay"},
		{name: "object", json: `{"href":"https://example.com/pay"}`, want: "https://example.com/pay"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var link Link
			if err := json.Unmarshal([]byte(test.json), &link); err != nil {
				t.Fatal(err)
			}
			if link.Href != test.want {
				t.Fatalf("Href = %q, want %q", link.Href, test.want)
			}
		})
	}
}

func TestClientReturnsStructuredAPIError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"code":"UNAUTHORIZED","message":"bad key","traceId":"trace-1","spanId":"span-1"}`))
	}))
	defer server.Close()

	client, err := NewClient("test-api-key", WithBaseURL(server.URL))
	if err != nil {
		t.Fatal(err)
	}
	client.httpClient = server.Client()
	_, err = client.GetPayment(context.Background(), "payment-123")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error type = %T, want *APIError: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized || apiErr.Code != "UNAUTHORIZED" ||
		apiErr.TraceID != "trace-1" || apiErr.SpanID != "span-1" {
		t.Fatalf("unexpected API error: %+v", apiErr)
	}
}
