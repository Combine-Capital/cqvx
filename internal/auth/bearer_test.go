package auth

import (
	"context"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBearerSigner(t *testing.T) {
	tests := []struct {
		name    string
		config  BearerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: BearerConfig{
				Token: "valid-bearer-token-12345",
			},
			wantErr: false,
		},
		{
			name: "empty token",
			config: BearerConfig{
				Token: "",
			},
			wantErr: true,
			errMsg:  "token is required",
		},
		{
			name: "long token",
			config: BearerConfig{
				Token: "very-long-bearer-token-with-many-characters-that-might-represent-a-jwt-or-other-complex-token-format",
			},
			wantErr: false,
		},
		{
			name: "token with special characters",
			config: BearerConfig{
				Token: "token-with-special-chars_!@#$%^&*()",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := NewBearerSigner(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, signer)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, signer)
				assert.Equal(t, tt.config.Token, signer.config.Token)
			}
		})
	}
}

func TestBearerSigner_Sign(t *testing.T) {
	tests := []struct {
		name       string
		config     BearerConfig
		request    SignRequest
		wantHeader string
	}{
		{
			name: "simple GET request",
			config: BearerConfig{
				Token: "test-token-123",
			},
			request: SignRequest{
				Method: "GET",
				Path:   "/api/v1/orders",
				Body:   nil,
			},
			wantHeader: "Bearer test-token-123",
		},
		{
			name: "POST request with body",
			config: BearerConfig{
				Token: "test-token-456",
			},
			request: SignRequest{
				Method: "POST",
				Path:   "/api/v1/orders",
				Body:   []byte(`{"symbol":"BTC-USD","side":"buy"}`),
			},
			wantHeader: "Bearer test-token-456",
		},
		{
			name: "long token",
			config: BearerConfig{
				Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
			},
			request: SignRequest{
				Method: "GET",
				Path:   "/api/v1/balances",
				Body:   nil,
			},
			wantHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ",
		},
		{
			name: "request with timestamp ignored",
			config: BearerConfig{
				Token: "test-token-789",
			},
			request: SignRequest{
				Method:    "POST",
				Path:      "/api/v1/orders",
				Body:      []byte(`{"test":"data"}`),
				Timestamp: "1234567890",
			},
			wantHeader: "Bearer test-token-789",
		},
		{
			name: "request with existing headers",
			config: BearerConfig{
				Token: "test-token-abc",
			},
			request: SignRequest{
				Method: "GET",
				Path:   "/api/v1/accounts",
				Headers: http.Header{
					"Content-Type": []string{"application/json"},
					"User-Agent":   []string{"test-client"},
				},
			},
			wantHeader: "Bearer test-token-abc",
		},
		{
			name: "empty body",
			config: BearerConfig{
				Token: "test-token-def",
			},
			request: SignRequest{
				Method: "DELETE",
				Path:   "/api/v1/orders/12345",
				Body:   []byte{},
			},
			wantHeader: "Bearer test-token-def",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := NewBearerSigner(tt.config)
			require.NoError(t, err)

			ctx := context.Background()
			result, err := signer.Sign(ctx, tt.request)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantHeader, result.Headers["Authorization"])
			assert.Empty(t, result.QueryParams, "Bearer auth should not add query params")
		})
	}
}

func TestBearerSigner_Sign_Concurrent(t *testing.T) {
	// Test that the signer is safe for concurrent use
	config := BearerConfig{
		Token: "concurrent-test-token",
	}
	signer, err := NewBearerSigner(config)
	require.NoError(t, err)

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	errors := make(chan error, numGoroutines)
	results := make(chan *SignResult, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			ctx := context.Background()
			req := SignRequest{
				Method: "GET",
				Path:   "/api/v1/test",
				Body:   []byte(`{"test":"data"}`),
			}

			result, err := signer.Sign(ctx, req)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}(i)
	}

	wg.Wait()
	close(errors)
	close(results)

	// Check for errors
	for err := range errors {
		t.Errorf("concurrent signing error: %v", err)
	}

	// Verify all results have the correct header
	count := 0
	for result := range results {
		assert.Equal(t, "Bearer concurrent-test-token", result.Headers["Authorization"])
		count++
	}
	assert.Equal(t, numGoroutines, count, "all goroutines should have completed")
}

func TestBearerSigner_Sign_ContextCancellation(t *testing.T) {
	// Test that signing respects context cancellation
	// Note: Current implementation doesn't use context, but this tests future compatibility
	config := BearerConfig{
		Token: "context-test-token",
	}
	signer, err := NewBearerSigner(config)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := SignRequest{
		Method: "GET",
		Path:   "/api/v1/test",
	}

	// Should still succeed since current implementation doesn't check context
	result, err := signer.Sign(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "Bearer context-test-token", result.Headers["Authorization"])
}

func TestBearerSigner_InterfaceCompliance(t *testing.T) {
	// Verify that BearerSigner implements the Signer interface
	var _ Signer = (*BearerSigner)(nil)

	// Test that it can be used through the interface
	var signer Signer
	config := BearerConfig{
		Token: "interface-test-token",
	}

	bearerSigner, err := NewBearerSigner(config)
	require.NoError(t, err)

	signer = bearerSigner

	ctx := context.Background()
	req := SignRequest{
		Method: "POST",
		Path:   "/api/v1/orders",
		Body:   []byte(`{"test":"data"}`),
	}

	result, err := signer.Sign(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "Bearer interface-test-token", result.Headers["Authorization"])
}

func TestBearerSigner_Middleware(t *testing.T) {
	// Test that the signer works with the Middleware function
	config := BearerConfig{
		Token: "middleware-test-token",
	}
	signer, err := NewBearerSigner(config)
	require.NoError(t, err)

	// Create a mock transport that captures the request
	var capturedRequest *http.Request
	mockTransport := &mockRoundTripper{
		roundTripFunc: func(req *http.Request) (*http.Response, error) {
			capturedRequest = req
			return &http.Response{
				StatusCode: 200,
				Body:       http.NoBody,
			}, nil
		},
	}

	// Create middleware
	transport := Middleware(signer, mockTransport)

	// Create and send a request
	req, err := http.NewRequest("GET", "http://example.com/api/v1/test", nil)
	require.NoError(t, err)

	_, err = transport.RoundTrip(req)
	require.NoError(t, err)

	// Verify the Authorization header was added
	require.NotNil(t, capturedRequest)
	assert.Equal(t, "Bearer middleware-test-token", capturedRequest.Header.Get("Authorization"))
}

func TestBearerSigner_EmptyBody(t *testing.T) {
	// Test various empty body scenarios
	config := BearerConfig{
		Token: "empty-body-token",
	}
	signer, err := NewBearerSigner(config)
	require.NoError(t, err)

	testCases := []struct {
		name string
		body []byte
	}{
		{"nil body", nil},
		{"empty slice", []byte{}},
		{"empty string bytes", []byte("")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			req := SignRequest{
				Method: "GET",
				Path:   "/api/v1/test",
				Body:   tc.body,
			}

			result, err := signer.Sign(ctx, req)
			require.NoError(t, err)
			assert.Equal(t, "Bearer empty-body-token", result.Headers["Authorization"])
		})
	}
}

func TestBearerSigner_LargeBody(t *testing.T) {
	// Test with a large request body
	config := BearerConfig{
		Token: "large-body-token",
	}
	signer, err := NewBearerSigner(config)
	require.NoError(t, err)

	// Create a large body (1MB)
	largeBody := make([]byte, 1024*1024)
	for i := range largeBody {
		largeBody[i] = byte('a' + (i % 26))
	}

	ctx := context.Background()
	req := SignRequest{
		Method: "POST",
		Path:   "/api/v1/large",
		Body:   largeBody,
	}

	result, err := signer.Sign(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "Bearer large-body-token", result.Headers["Authorization"])
}

// mockRoundTripper is a mock http.RoundTripper for testing
type mockRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}
