package payments

import (
	"encoding/json"
	"errors"
	"time"
)

// Currency is the ISO 4217 currency used by Bancontact Pro.
type Currency string

const CurrencyEUR Currency = "EUR"

// PaymentStatus is a payment lifecycle status. Unknown future values are kept
// as strings when decoding responses.
type PaymentStatus string

const (
	PaymentStatusPending                       PaymentStatus = "PENDING"
	PaymentStatusIdentified                    PaymentStatus = "IDENTIFIED"
	PaymentStatusAuthorized                    PaymentStatus = "AUTHORIZED"
	PaymentStatusAuthorizationFailed           PaymentStatus = "AUTHORIZATION_FAILED"
	PaymentStatusSucceeded                     PaymentStatus = "SUCCEEDED"
	PaymentStatusFailed                        PaymentStatus = "FAILED"
	PaymentStatusCancelled                     PaymentStatus = "CANCELLED"
	PaymentStatusExpired                       PaymentStatus = "EXPIRED"
	PaymentStatusPendingMerchantAcknowledgment PaymentStatus = "PENDING_MERCHANT_ACKNOWLEDGEMENT"
	PaymentStatusVoided                        PaymentStatus = "VOIDED"
)

type Void struct{}

// Link is a URL returned by a payment response.
type Link struct {
	Href string `json:"href"`
}

func (l *Link) UnmarshalJSON(data []byte) error {
	var href string
	if err := json.Unmarshal(data, &href); err == nil {
		l.Href = href
		return nil
	}
	type link Link
	var object link
	if err := json.Unmarshal(data, &object); err != nil {
		return errors.New("link must be a URL string or an object containing href")
	}
	l.Href = object.Href
	return nil
}

// PaymentLinks contains the URLs supplied by Bancontact for the payment flow.
type PaymentLinks struct {
	Self     *Link `json:"self,omitempty"`
	DeepLink *Link `json:"deeplink,omitempty"`
	QRCode   *Link `json:"qrcode,omitempty"`
	Cancel   *Link `json:"cancel,omitempty"`
	Refund   *Link `json:"refund,omitempty"`
	Checkout *Link `json:"checkout,omitempty"`
}

// Creditor describes the merchant receiving a payment.
type Creditor struct {
	ProfileID           string `json:"profileId,omitempty"`
	MerchantID          string `json:"merchantId,omitempty"`
	Name                string `json:"name,omitempty"`
	IBAN                string `json:"iban,omitempty"`
	IdentifyCallbackURL string `json:"identifyCallbackUrl,omitempty"`
	CallbackURL         string `json:"callbackUrl,omitempty"`
}

// CreatePaymentRequest starts a dynamic payment. Amount is in euro cents.
type CreatePaymentRequest struct {
	Reference           string                  `json:"reference,omitempty"`
	BulkID              string                  `json:"bulkId,omitempty"`
	Amount              int64                   `json:"amount"`
	Currency            Currency                `json:"currency,omitempty"`
	Description         string                  `json:"description,omitempty"`
	IdentifyCallbackURL string                  `json:"identifyCallbackUrl,omitempty"`
	CallbackURL         string                  `json:"callbackUrl,omitempty"`
	ReturnURL           string                  `json:"returnUrl,omitempty"`
	VoucherEligibility  []VoucherEligibleAmount `json:"voucherEligibility,omitempty"`
}

// CreateStaticQRPaymentRequest starts a payment for a static QR point of sale.
type CreateStaticQRPaymentRequest struct {
	Reference           string                  `json:"reference,omitempty"`
	BulkID              string                  `json:"bulkId,omitempty"`
	Amount              int64                   `json:"amount"`
	Currency            Currency                `json:"currency,omitempty"`
	Description         string                  `json:"description,omitempty"`
	POSID               string                  `json:"posId"`
	ShopID              string                  `json:"shopId,omitempty"`
	ShopName            string                  `json:"shopName,omitempty"`
	IdentifyCallbackURL string                  `json:"identifyCallbackUrl,omitempty"`
	CallbackURL         string                  `json:"callbackUrl,omitempty"`
	VoucherEligibility  []VoucherEligibleAmount `json:"voucherEligibility,omitempty"`
}

// VoucherEligibleAmount is a deprecated voucher amount field in the Payment API.
type VoucherEligibleAmount struct {
	VoucherSchemes []string `json:"voucherSchemes,omitempty"`
	Amount         int64    `json:"amount"`
}

// CreatePaymentResponse is returned when a payment has been created.
type CreatePaymentResponse struct {
	PaymentID   string        `json:"paymentId"`
	Status      PaymentStatus `json:"status"`
	CreatedAt   time.Time     `json:"createdAt"`
	ExpiresAt   time.Time     `json:"expiresAt"`
	Description string        `json:"description,omitempty"`
	Reference   string        `json:"reference,omitempty"`
	Amount      int64         `json:"amount"`
	Currency    Currency      `json:"currency"`
	Creditor    Creditor      `json:"creditor"`
	Links       PaymentLinks  `json:"_links"`
}

// Payment is a payment returned by get or search operations.
type Payment struct {
	PaymentID   string          `json:"paymentId"`
	CreatedAt   time.Time       `json:"createdAt"`
	ExpiresAt   time.Time       `json:"expireAt"`
	SucceededAt *time.Time      `json:"succeededAt,omitempty"`
	Currency    Currency        `json:"currency,omitempty"`
	Status      PaymentStatus   `json:"status"`
	Creditor    Creditor        `json:"creditor,omitempty"`
	Debtor      json.RawMessage `json:"debtor,omitempty"`
	Amount      int64           `json:"amount"`
	Description string          `json:"description,omitempty"`
	Message     string          `json:"message,omitempty"`
	Reference   string          `json:"reference,omitempty"`
	BulkID      string          `json:"bulkId,omitempty"`
	Links       PaymentLinks    `json:"_links,omitempty"`
}

// SearchPaymentsRequest filters payment search results. Page is zero-based;
// a zero Size or Page leaves that parameter to the API's default.
type SearchPaymentsRequest struct {
	Page            int             `json:"-"`
	Size            int             `json:"-"`
	From            *time.Time      `json:"from,omitempty"`
	To              *time.Time      `json:"to,omitempty"`
	PaymentStatuses []PaymentStatus `json:"paymentStatuses,omitempty"`
	Reference       string          `json:"reference,omitempty"`
}

// PaymentSearchResponse contains a paginated payment search result.
type PaymentSearchResponse struct {
	Size          int       `json:"size"`
	TotalPages    int       `json:"totalPages"`
	TotalElements int       `json:"totalElements"`
	Number        int       `json:"number"`
	Details       []Payment `json:"details"`
}

// AcknowledgePaymentRequest is sent when merchant acknowledgment is enabled.
type AcknowledgePaymentRequest struct {
	Currency  Currency `json:"currency"`
	Amount    int64    `json:"amount"`
	Reference string   `json:"reference,omitempty"`
}

// RefundIBANResponse contains the debtor's IBAN for a succeeded payment.
type RefundIBANResponse struct {
	IBAN string `json:"iban"`
}

// CreateRefundRequest initiates a full or partial refund. Amount is in cents.
type CreateRefundRequest struct {
	Amount      int64    `json:"amount"`
	Currency    Currency `json:"currency"`
	Description string   `json:"description,omitempty"`
}

// Refund describes a refund operation.
type Refund struct {
	RefundID     string       `json:"refundId"`
	PaymentID    string       `json:"paymentId"`
	Status       RefundStatus `json:"status"`
	Amount       int64        `json:"amount"`
	Currency     Currency     `json:"currency"`
	Description  string       `json:"description,omitempty"`
	CreationDate time.Time    `json:"creationDate"`
}

// RefundStatus is a refund lifecycle status.
type RefundStatus string

const (
	RefundStatusPending  RefundStatus = "PENDING"
	RefundStatusRefunded RefundStatus = "REFUNDED"
	RefundStatusFailed   RefundStatus = "FAILED"
)
