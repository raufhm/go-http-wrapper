package go_http_wrapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/stretchr/testify/assert"
)

func TestClient_Get(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "1", r.URL.Query().Get("page"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
	defer ts.Close()

	// Create client
	client := New(ts.URL)

	// Test GET request
	resp, err := client.Get(context.Background(), "/test",
		WithQueryParams(map[string][]string{
			"page": {"1"},
		}),
	)

	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"message":"ok"}`), resp)
}

func TestClient_Post(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/test", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer ts.Close()

	// Create client
	client := New(ts.URL)

	// Test POST request
	resp, err := client.Post(context.Background(), "/test",
		WithBodyRequest(map[string]interface{}{
			"name": "test",
		}),
	)

	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"id":1}`), resp)
}

func TestClient_RetryWithBackoff(t *testing.T) {
	attempts := 0
	maxRetries := 2 // Changed to 2 to match the test scenario

	// Create test server that fails twice then succeeds
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts <= maxRetries {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"ok"}`))
	}))
	defer ts.Close()

	// Create client with custom backoff
	client := New(ts.URL,
		WithBackoff(newTestBackoff(maxRetries, 100*time.Millisecond)),
	)

	// Test retry behavior
	resp, err := client.Get(context.Background(), "/test")

	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"message":"ok"}`), resp)
	assert.Equal(t, maxRetries+1, attempts) // +1 for the successful attempt
}

// Updated helper function to properly handle maxRetries
func newTestBackoff(maxRetries int, interval time.Duration) backoff.BackOff {
	b := backoff.NewConstantBackOff(interval)
	return backoff.WithMaxRetries(b, uint64(maxRetries))
}
