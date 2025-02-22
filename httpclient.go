package go_http_wrapper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/cenkalti/backoff/v4"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type HTTPClient interface {
	Get(ctx context.Context, path string, opts ...RequestOption) ([]byte, error)
	Post(ctx context.Context, path string, opts ...RequestOption) ([]byte, error)
	Put(ctx context.Context, path string, opts ...RequestOption) ([]byte, error)
	Patch(ctx context.Context, path string, opts ...RequestOption) ([]byte, error)
	Delete(ctx context.Context, path string, opts ...RequestOption) ([]byte, error)
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    map[string]string
	backoff    backoff.BackOff
}

type ClientOption func(*Client)

// WithTimeout sets the client timeout
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithBackoff sets custom backoff configuration
func WithBackoff(b backoff.BackOff) ClientOption {
	return func(c *Client) {
		c.backoff = b
	}
}

// WithHeaders sets default headers
func WithHeaders(headers map[string]string) ClientOption {
	return func(c *Client) {
		c.headers = headers
	}
}

func New(baseURL string, opts ...ClientOption) *Client {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.MaxElapsedTime = 30 * time.Second

	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
		backoff: expBackoff,
	}
	client.httpClient.Transport = newrelic.NewRoundTripper(client.httpClient.Transport)

	for _, opt := range opts {
		opt(client)
	}

	return client
}

type RequestOption func(*http.Request) error

// WithQueryParams adds query parameters to the request
func WithQueryParams(params map[string][]string) RequestOption {
	return func(req *http.Request) error {
		q := req.URL.Query()
		for key, values := range params {
			for _, value := range values {
				q.Add(key, value)
			}
		}
		req.URL.RawQuery = q.Encode()
		return nil
	}
}

// WithBodyRequest adds JSON body to the request
func WithBodyRequest(body interface{}) RequestOption {
	return func(req *http.Request) error {
		if body == nil {
			return nil
		}
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		return nil
	}
}

func (c *Client) do(ctx context.Context, method, path string, opts ...RequestOption) ([]byte, error) {
	var respBody []byte
	operation := func() error {
		txn := newrelic.FromContext(ctx)

		reqURL, err := url.JoinPath(c.baseURL, path)
		if err != nil {
			return backoff.Permanent(fmt.Errorf("invalid URL: %w", err))
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, nil)
		if err != nil {
			return backoff.Permanent(fmt.Errorf("failed to create request: %w", err))
		}

		// Set default headers
		for key, value := range c.headers {
			req.Header.Set(key, value)
		}

		// Apply request options
		for _, opt := range opts {
			if err := opt(req); err != nil {
				return backoff.Permanent(err)
			}
		}

		req = newrelic.RequestWithTransactionContext(req, txn)

		// Make request
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		// Read response
		respBody, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		// Check status code
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			// Don't retry 4xx errors
			if resp.StatusCode >= 400 && resp.StatusCode < 500 {
				return backoff.Permanent(fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody)))
			}
			return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
		}

		return nil
	}

	err := backoff.RetryNotify(operation, backoff.WithContext(c.backoff, ctx),
		func(err error, duration time.Duration) {
			if txn := newrelic.FromContext(ctx); txn != nil {
				txn.NoticeError(err)
			}
		})

	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (c *Client) Get(ctx context.Context, path string, opts ...RequestOption) ([]byte, error) {
	return c.do(ctx, http.MethodGet, path, opts...)
}

func (c *Client) Post(ctx context.Context, path string, opts ...RequestOption) ([]byte, error) {
	return c.do(ctx, http.MethodPost, path, opts...)
}

func (c *Client) Put(ctx context.Context, path string, opts ...RequestOption) ([]byte, error) {
	return c.do(ctx, http.MethodPut, path, opts...)
}

func (c *Client) Patch(ctx context.Context, path string, opts ...RequestOption) ([]byte, error) {
	return c.do(ctx, http.MethodPatch, path, opts...)
}

func (c *Client) Delete(ctx context.Context, path string, opts ...RequestOption) ([]byte, error) {
	return c.do(ctx, http.MethodDelete, path, opts...)
}
