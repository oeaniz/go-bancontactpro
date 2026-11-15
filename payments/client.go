package payments

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	productionBaseURL    = "https://merchant.api.bancontact.net"
	preProductionBaseURL = "https://merchant.api.preprod.bancontact.net"
)

// Environment selects a Bancontact Pro API environment.
type Environment string

const (
	EnvironmentProduction    Environment = "production"
	EnvironmentPreProduction Environment = "preproduction"
)

type Option func(*options)
type options struct {
	env     *Environment
	baseURL string
}

func WithEnvironment(env Environment) Option {
	return func(opts *options) {
		opts.env = new(env)
	}
}

// WithBaseURL overrides the API base URL. It can be used to point the client at
// a compatible mock server.
func WithBaseURL(baseURL string) Option {
	return func(opts *options) {
		opts.baseURL = strings.TrimRight(baseURL, "/")
	}
}

// Client calls the Bancontact Pro merchant APIs.
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a client configured for a Bancontact Pro environment.
// It defaults to production unless an environment or custom base URL is set.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	cfg := &options{}
	for _, opt := range opts {
		opt(cfg)
	}

	baseURL := productionBaseURL
	if cfg.env != nil {
		switch *cfg.env {
		case EnvironmentProduction:
			baseURL = productionBaseURL
		case EnvironmentPreProduction:
			baseURL = preProductionBaseURL
		default:
			return nil, errors.New("unsupported environment")
		}
	}
	if cfg.baseURL != "" {
		parsedURL, err := url.Parse(cfg.baseURL)
		if err != nil {
			return nil, fmt.Errorf("cannot parse base URL: %w", err)
		}
		if parsedURL.Scheme != "https" || parsedURL.Host == "" ||
			parsedURL.User != nil || parsedURL.RawQuery != "" || parsedURL.Fragment != "" {
			return nil, errors.New("base URL must use HTTPS and cannot contain credentials, query, or fragment")
		}
		baseURL = strings.TrimRight(parsedURL.String(), "/")
	}

	client := &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{},
		baseURL:    baseURL,
	}

	return client, nil
}

type CallParams struct {
	query          url.Values
	idempotencyKey string
}

func (c *Client) requesting(path string, query url.Values, body any) (string, io.Reader, error) {
	requestURL, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", nil, fmt.Errorf("cannot parse url: %w", err)
	}
	requestURL.RawQuery = query.Encode()

	var requestBody io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return "", nil, fmt.Errorf("cannot unmarshal body: %w", err)
		}
		requestBody = bytes.NewReader(bodyBytes)
	}
	return requestURL.String(), requestBody, nil
}

func (c *Client) get[T any](ctx context.Context, path string, params CallParams) (*T, error) {
	return c.do[T](ctx, http.MethodGet, path, nil, params)
}

func (c *Client) post[T any](ctx context.Context, path string, body any, params CallParams) (*T, error) {
	return c.do[T](ctx, http.MethodPost, path, body, params)
}

func (c *Client) do[T any](ctx context.Context, method string, path string, body any, params CallParams) (*T, error) {
	requestURL, requestBody, err := c.requesting(path, params.query, body)
	if err != nil {
		return nil, fmt.Errorf("cannot construct request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL, requestBody)
	if err != nil {
		return nil, fmt.Errorf("could not create request for bancontact: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", c.apiKey)
	if params.idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", params.idempotencyKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request for bancontact: %w", err)
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not read Bancontact Pro response: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, parseAPIError(resp.StatusCode, responseBody)
	}
	var result T
	if len(responseBody) > 0 {
		if err = json.Unmarshal(responseBody, &result); err != nil {
			return nil, fmt.Errorf("cannot decode Bancontact Pro response: %w", err)
		}
	}
	return &result, nil
}

// APIError represents an unsuccessful response from a Bancontact Pro API.
type APIError struct {
	StatusCode int
	Code       string
	Message    string
	TraceID    string
	SpanID     string
	Body       string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("Bancontact Pro API returned HTTP %d (%s): %s", e.StatusCode, e.Code, e.Message)
	}
	return fmt.Sprintf("Bancontact Pro API returned HTTP %d: %s", e.StatusCode, e.Message)
}

func parseAPIError(statusCode int, body []byte) error {
	apiErr := &APIError{StatusCode: statusCode, Body: string(body)}
	var response struct {
		Code    string `json:"code"`
		Message string `json:"message"`
		TraceID string `json:"traceId"`
		SpanID  string `json:"spanId"`
	}
	if len(body) != 0 && json.Unmarshal(body, &response) == nil {
		apiErr.Code = response.Code
		apiErr.Message = response.Message
		apiErr.TraceID = response.TraceID
		apiErr.SpanID = response.SpanID
	}
	if apiErr.Message == "" {
		apiErr.Message = strings.TrimSpace(string(body))
	}
	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(statusCode)
	}
	return apiErr
}
