package payments

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// CreatePayment creates a dynamic payment.
func (c *Client) CreatePayment(ctx context.Context, request CreatePaymentRequest) (*CreatePaymentResponse, error) {
	return c.post[CreatePaymentResponse](ctx, "/v3/payments", request, CallParams{})
}

// CreateStaticQRPayment creates a payment for a static QR point of sale.
func (c *Client) CreateStaticQRPayment(ctx context.Context, request CreateStaticQRPaymentRequest) (*CreatePaymentResponse, error) {
	return c.post[CreatePaymentResponse](ctx, "/v3/payments/pos", request, CallParams{})
}

// GetPayment retrieves a payment by its Bancontact payment ID.
func (c *Client) GetPayment(ctx context.Context, paymentID string) (*Payment, error) {
	return c.get[Payment](ctx, "/v3/payments/"+url.PathEscape(paymentID), CallParams{})
}

// SearchPayments searches payments using the supplied filters and pagination.
func (c *Client) SearchPayments(ctx context.Context, request SearchPaymentsRequest) (*PaymentSearchResponse, error) {
	query := make(url.Values)
	if request.Page > 0 {
		query.Set("page", strconv.Itoa(request.Page))
	}
	if request.Size > 0 {
		query.Set("size", strconv.Itoa(request.Size))
	}
	return c.post[PaymentSearchResponse](ctx, "/v3/payments/search", request, CallParams{query: query})
}

// CancelPayment cancels a payment that is still pending or identified.
func (c *Client) CancelPayment(ctx context.Context, paymentID string) error {
	_, err := c.do[Void](ctx, http.MethodDelete, "/v3/payments/"+url.PathEscape(paymentID), nil, CallParams{})
	return err
}

// AcknowledgePayment acknowledges receipt of a payment status callback.
func (c *Client) AcknowledgePayment(ctx context.Context, paymentID string, request AcknowledgePaymentRequest) error {
	_, err := c.post[Void](ctx, "/v3/payments/"+url.PathEscape(paymentID)+"/acknowledge", request, CallParams{})
	return err
}

// GetRefundIBAN returns the debtor's IBAN for a succeeded payment.
func (c *Client) GetRefundIBAN(ctx context.Context, paymentID string) (*RefundIBANResponse, error) {
	return c.get[RefundIBANResponse](ctx, "/v3/payments/"+url.PathEscape(paymentID)+"/debtor/refundIban", CallParams{})
}
