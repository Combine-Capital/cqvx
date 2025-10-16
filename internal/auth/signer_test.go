package auth_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Combine-Capital/cqvx/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSigner is a test implementation of the Signer interface
type mockSigner struct {
	signFunc func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error)
}

func (m *mockSigner) Sign(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
	if m.signFunc != nil {
		return m.signFunc(ctx, req)
	}
	return &auth.SignResult{
		Headers: map[string]string{
			"X-Test-Auth": "test-signature",
		},
	}, nil
}

// TestSignRequest_Structure verifies SignRequest contains expected fields
func TestSignRequest_Structure(t *testing.T) {
	req := auth.SignRequest{
		Method:    "POST",
		Path:      "/orders",
		Body:      []byte(`{"symbol":"BTC-USD"}`),
		Timestamp: "1234567890",
		Headers:   http.Header{"Content-Type": []string{"application/json"}},
	}

	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "/orders", req.Path)
	assert.NotNil(t, req.Body)
	assert.Equal(t, "1234567890", req.Timestamp)
	assert.NotNil(t, req.Headers)
}

// TestSignResult_Structure verifies SignResult contains expected fields
func TestSignResult_Structure(t *testing.T) {
	result := auth.SignResult{
		Headers: map[string]string{
			"Authorization": "Bearer token123",
			"X-API-Key":     "key123",
		},
		QueryParams: map[string]string{
			"signature": "sig123",
			"timestamp": "1234567890",
		},
	}

	assert.Len(t, result.Headers, 2)
	assert.Equal(t, "Bearer token123", result.Headers["Authorization"])
	assert.Len(t, result.QueryParams, 2)
	assert.Equal(t, "sig123", result.QueryParams["signature"])
}

// TestMiddleware_AppliesAuthenticationHeaders tests that middleware adds headers
func TestMiddleware_AppliesAuthenticationHeaders(t *testing.T) {
	// Create a test HTTP server to capture the request
	var capturedRequest *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequest = r
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create mock signer that adds a custom header
	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			return &auth.SignResult{
				Headers: map[string]string{
					"X-Signature": "test-sig-123",
					"X-Timestamp": "1234567890",
					"X-API-Key":   "test-api-key",
				},
			}, nil
		},
	}

	// Create HTTP client with auth middleware
	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	// Make request
	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify auth headers were added
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, capturedRequest)
	assert.Equal(t, "test-sig-123", capturedRequest.Header.Get("X-Signature"))
	assert.Equal(t, "1234567890", capturedRequest.Header.Get("X-Timestamp"))
	assert.Equal(t, "test-api-key", capturedRequest.Header.Get("X-API-Key"))
}

// TestMiddleware_AppliesQueryParameters tests that middleware adds query params
func TestMiddleware_AppliesQueryParameters(t *testing.T) {
	var capturedRequest *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequest = r
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create mock signer that adds query parameters
	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			return &auth.SignResult{
				QueryParams: map[string]string{
					"api_key":   "test-key",
					"timestamp": "1234567890",
					"signature": "test-sig",
				},
			}, nil
		},
	}

	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	req, err := http.NewRequest("GET", server.URL+"/test?existing=param", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify query parameters were added
	require.NotNil(t, capturedRequest)
	query := capturedRequest.URL.Query()
	assert.Equal(t, "param", query.Get("existing"))
	assert.Equal(t, "test-key", query.Get("api_key"))
	assert.Equal(t, "1234567890", query.Get("timestamp"))
	assert.Equal(t, "test-sig", query.Get("signature"))
}

// TestMiddleware_WithRequestBody tests signing requests with body
func TestMiddleware_WithRequestBody(t *testing.T) {
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		capturedBody, err = io.ReadAll(r.Body)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	requestBody := `{"symbol":"BTC-USD","quantity":1.5}`

	// Create signer that signs based on body content
	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			// Verify the body is available for signing
			assert.Equal(t, requestBody, string(req.Body))
			assert.Equal(t, "POST", req.Method)

			return &auth.SignResult{
				Headers: map[string]string{
					"X-Body-Signature": "body-sig-" + string(req.Body),
				},
			}, nil
		},
	}

	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	req, err := http.NewRequest("POST", server.URL+"/orders", strings.NewReader(requestBody))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify body was still sent to server
	assert.Equal(t, requestBody, string(capturedBody))
}

// TestMiddleware_SignerError tests error handling when signer fails
func TestMiddleware_SignerError(t *testing.T) {
	// Create signer that returns an error
	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			return nil, assert.AnError
		},
	}

	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	req, err := http.NewRequest("GET", "http://example.com/test", nil)
	require.NoError(t, err)

	_, err = client.Do(req)
	assert.Error(t, err)
	// The error is wrapped in url.Error, check it contains our error
	assert.Contains(t, err.Error(), assert.AnError.Error())
}

// TestMiddleware_PreservesExistingHeaders tests that existing headers are not lost
func TestMiddleware_PreservesExistingHeaders(t *testing.T) {
	var capturedRequest *http.Request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedRequest = r
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			return &auth.SignResult{
				Headers: map[string]string{
					"X-Auth": "auth-value",
				},
			}, nil
		},
	}

	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-client")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify both existing and new headers are present
	require.NotNil(t, capturedRequest)
	assert.Equal(t, "application/json", capturedRequest.Header.Get("Content-Type"))
	assert.Equal(t, "test-client", capturedRequest.Header.Get("User-Agent"))
	assert.Equal(t, "auth-value", capturedRequest.Header.Get("X-Auth"))
}

// TestMiddleware_HandlesEmptyBody tests requests without body
func TestMiddleware_HandlesEmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			// Verify empty body
			assert.Nil(t, req.Body)
			return &auth.SignResult{
				Headers: map[string]string{"X-Auth": "test"},
			}, nil
		},
	}

	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	req, err := http.NewRequest("GET", server.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestMiddleware_HandlesLargeBody tests requests with large body
func TestMiddleware_HandlesLargeBody(t *testing.T) {
	var capturedBody []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		capturedBody, err = io.ReadAll(r.Body)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a large body (1MB)
	largeBody := strings.Repeat("x", 1024*1024)

	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			assert.Equal(t, len(largeBody), len(req.Body))
			return &auth.SignResult{
				Headers: map[string]string{"X-Auth": "test"},
			}, nil
		},
	}

	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	req, err := http.NewRequest("POST", server.URL+"/test", strings.NewReader(largeBody))
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify entire body was transmitted
	assert.Equal(t, len(largeBody), len(capturedBody))
	assert.Equal(t, largeBody, string(capturedBody))
}

// TestMiddleware_ContextPropagation tests that context is passed to signer
func TestMiddleware_ContextPropagation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	type ctxKey string
	const testKey ctxKey = "test-key"
	const testValue = "test-value"

	var capturedContext context.Context
	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			capturedContext = ctx
			return &auth.SignResult{
				Headers: map[string]string{"X-Auth": "test"},
			}, nil
		},
	}

	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	ctx := context.WithValue(context.Background(), testKey, testValue)
	req, err := http.NewRequestWithContext(ctx, "GET", server.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Verify context was passed through
	require.NotNil(t, capturedContext)
	assert.Equal(t, testValue, capturedContext.Value(testKey))
}

// TestMiddleware_ConcurrentRequests tests thread safety
func TestMiddleware_ConcurrentRequests(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	callCount := 0
	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			callCount++
			return &auth.SignResult{
				Headers: map[string]string{"X-Auth": "test"},
			}, nil
		},
	}

	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	// Make concurrent requests
	const numRequests = 10
	done := make(chan bool, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req, err := http.NewRequest("GET", server.URL+"/test", nil)
			if err != nil {
				t.Error(err)
				done <- false
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Error(err)
				done <- false
				return
			}
			resp.Body.Close()

			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < numRequests; i++ {
		success := <-done
		assert.True(t, success)
	}
}

// TestMiddleware_NilTransport tests that nil transport uses default
func TestMiddleware_NilTransport(t *testing.T) {
	signer := &mockSigner{}
	transport := auth.Middleware(signer, nil)
	assert.NotNil(t, transport)
}

// TestSignRequest_PathExtraction tests that path is correctly extracted
func TestSignRequest_PathExtraction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	var capturedPath string
	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			capturedPath = req.Path
			return &auth.SignResult{}, nil
		},
	}

	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	req, err := http.NewRequest("GET", server.URL+"/api/v1/orders/123?limit=10", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "/api/v1/orders/123", capturedPath)
}
