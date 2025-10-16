package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMPCSigner(t *testing.T) {
	mockSignerFunc := func(ctx context.Context, message []byte) (string, error) {
		return "mock-signature", nil
	}

	tests := []struct {
		name    string
		config  MPCConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: MPCConfig{
				APIKey:     "valid-api-key",
				SignerFunc: mockSignerFunc,
			},
			wantErr: false,
		},
		{
			name: "empty API key",
			config: MPCConfig{
				APIKey:     "",
				SignerFunc: mockSignerFunc,
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
		{
			name: "nil signer function",
			config: MPCConfig{
				APIKey:     "valid-api-key",
				SignerFunc: nil,
			},
			wantErr: true,
			errMsg:  "signer function is required",
		},
		{
			name: "both empty",
			config: MPCConfig{
				APIKey:     "",
				SignerFunc: nil,
			},
			wantErr: true,
			errMsg:  "API key is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := NewMPCSigner(tt.config)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, signer)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, signer)
				assert.Equal(t, tt.config.APIKey, signer.config.APIKey)
			}
		})
	}
}

func TestMPCSigner_Sign(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		signerFunc     func(ctx context.Context, message []byte) (string, error)
		request        SignRequest
		wantErr        bool
		errMsg         string
		validateResult func(t *testing.T, result *SignResult)
	}{
		{
			name:   "simple GET request",
			apiKey: "test-api-key",
			signerFunc: func(ctx context.Context, message []byte) (string, error) {
				return "signature-for-get", nil
			},
			request: SignRequest{
				Method: "GET",
				Path:   "/api/v1/orders",
				Body:   nil,
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *SignResult) {
				assert.Equal(t, "test-api-key", result.Headers["X-API-KEY"])
				assert.Equal(t, "signature-for-get", result.Headers["X-SIGNATURE"])
				assert.NotEmpty(t, result.Headers["X-TIMESTAMP"])
			},
		},
		{
			name:   "POST request with body",
			apiKey: "test-api-key",
			signerFunc: func(ctx context.Context, message []byte) (string, error) {
				// Verify message format: timestamp + method + path + body
				msgStr := string(message)
				assert.Contains(t, msgStr, "POST")
				assert.Contains(t, msgStr, "/api/v1/orders")
				assert.Contains(t, msgStr, `{"symbol":"BTC-USD"}`)
				return "signature-for-post", nil
			},
			request: SignRequest{
				Method: "POST",
				Path:   "/api/v1/orders",
				Body:   []byte(`{"symbol":"BTC-USD"}`),
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *SignResult) {
				assert.Equal(t, "signature-for-post", result.Headers["X-SIGNATURE"])
			},
		},
		{
			name:   "signer function returns error",
			apiKey: "test-api-key",
			signerFunc: func(ctx context.Context, message []byte) (string, error) {
				return "", errors.New("MPC service unavailable")
			},
			request: SignRequest{
				Method: "GET",
				Path:   "/api/v1/test",
			},
			wantErr: true,
			errMsg:  "MPC signing failed",
		},
		{
			name:   "custom timestamp",
			apiKey: "test-api-key",
			signerFunc: func(ctx context.Context, message []byte) (string, error) {
				msgStr := string(message)
				// Verify the custom timestamp is used
				assert.Contains(t, msgStr, "1234567890")
				return "signature-with-custom-timestamp", nil
			},
			request: SignRequest{
				Method:    "GET",
				Path:      "/api/v1/test",
				Timestamp: "1234567890",
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *SignResult) {
				assert.Equal(t, "1234567890", result.Headers["X-TIMESTAMP"])
			},
		},
		{
			name:   "empty body",
			apiKey: "test-api-key",
			signerFunc: func(ctx context.Context, message []byte) (string, error) {
				return "signature-empty-body", nil
			},
			request: SignRequest{
				Method: "DELETE",
				Path:   "/api/v1/orders/12345",
				Body:   []byte{},
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *SignResult) {
				assert.Equal(t, "signature-empty-body", result.Headers["X-SIGNATURE"])
			},
		},
		{
			name:   "large body",
			apiKey: "test-api-key",
			signerFunc: func(ctx context.Context, message []byte) (string, error) {
				// Verify we receive the full large message
				assert.True(t, len(message) > 1000)
				return "signature-large-body", nil
			},
			request: SignRequest{
				Method: "POST",
				Path:   "/api/v1/large",
				Body:   make([]byte, 10000),
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *SignResult) {
				assert.Equal(t, "signature-large-body", result.Headers["X-SIGNATURE"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := MPCConfig{
				APIKey:     tt.apiKey,
				SignerFunc: tt.signerFunc,
			}
			signer, err := NewMPCSigner(config)
			require.NoError(t, err)

			ctx := context.Background()
			result, err := signer.Sign(ctx, tt.request)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)

			// Common validations
			assert.NotEmpty(t, result.Headers["X-API-KEY"])
			assert.NotEmpty(t, result.Headers["X-SIGNATURE"])
			assert.NotEmpty(t, result.Headers["X-TIMESTAMP"])
			assert.Empty(t, result.QueryParams, "MPC auth should not add query params")

			// Custom validations
			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestMPCSigner_Sign_TimestampGeneration(t *testing.T) {
	// Test that timestamp is generated correctly when not provided
	config := MPCConfig{
		APIKey: "test-api-key",
		SignerFunc: func(ctx context.Context, message []byte) (string, error) {
			return "test-signature", nil
		},
	}
	signer, err := NewMPCSigner(config)
	require.NoError(t, err)

	ctx := context.Background()
	req := SignRequest{
		Method: "GET",
		Path:   "/api/v1/test",
	}

	// Get time before signing
	beforeTime := time.Now().UnixMilli()

	result, err := signer.Sign(ctx, req)
	require.NoError(t, err)

	// Get time after signing
	afterTime := time.Now().UnixMilli()

	// Parse the timestamp from result
	timestamp, err := strconv.ParseInt(result.Headers["X-TIMESTAMP"], 10, 64)
	require.NoError(t, err)

	// Verify timestamp is within the expected range
	assert.GreaterOrEqual(t, timestamp, beforeTime)
	assert.LessOrEqual(t, timestamp, afterTime)
}

func TestMPCSigner_Sign_Concurrent(t *testing.T) {
	// Test that the signer is safe for concurrent use
	var callCount int
	var mu sync.Mutex

	config := MPCConfig{
		APIKey: "concurrent-test-key",
		SignerFunc: func(ctx context.Context, message []byte) (string, error) {
			mu.Lock()
			callCount++
			mu.Unlock()
			// Simulate some work
			time.Sleep(time.Millisecond)
			hash := sha256.Sum256(message)
			return hex.EncodeToString(hash[:]), nil
		},
	}
	signer, err := NewMPCSigner(config)
	require.NoError(t, err)

	const numGoroutines = 50
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
				Body:   []byte("test-data"),
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

	// Verify all results
	resultCount := 0
	for result := range results {
		assert.Equal(t, "concurrent-test-key", result.Headers["X-API-KEY"])
		assert.NotEmpty(t, result.Headers["X-SIGNATURE"])
		assert.NotEmpty(t, result.Headers["X-TIMESTAMP"])
		resultCount++
	}
	assert.Equal(t, numGoroutines, resultCount)
	assert.Equal(t, numGoroutines, callCount, "signer function should be called once per goroutine")
}

func TestMPCSigner_Sign_ContextPropagation(t *testing.T) {
	// Test that context is properly propagated to signer function
	type contextKey string
	const testKey contextKey = "test-key"
	const testValue = "test-value"

	var receivedCtx context.Context
	config := MPCConfig{
		APIKey: "test-api-key",
		SignerFunc: func(ctx context.Context, message []byte) (string, error) {
			receivedCtx = ctx
			return "test-signature", nil
		},
	}
	signer, err := NewMPCSigner(config)
	require.NoError(t, err)

	ctx := context.WithValue(context.Background(), testKey, testValue)
	req := SignRequest{
		Method: "GET",
		Path:   "/api/v1/test",
	}

	_, err = signer.Sign(ctx, req)
	require.NoError(t, err)

	// Verify context was propagated
	assert.NotNil(t, receivedCtx)
	assert.Equal(t, testValue, receivedCtx.Value(testKey))
}

func TestMPCSigner_Sign_ContextCancellation(t *testing.T) {
	// Test that context cancellation is respected by signer function
	config := MPCConfig{
		APIKey: "test-api-key",
		SignerFunc: func(ctx context.Context, message []byte) (string, error) {
			// Check if context is cancelled
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			default:
				return "test-signature", nil
			}
		},
	}
	signer, err := NewMPCSigner(config)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := SignRequest{
		Method: "GET",
		Path:   "/api/v1/test",
	}

	_, err = signer.Sign(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MPC signing failed")
}

func TestDefaultMPCSignerFunc(t *testing.T) {
	// Test the default MPC signer function
	tests := []struct {
		name    string
		message []byte
	}{
		{
			name:    "simple message",
			message: []byte("test message"),
		},
		{
			name:    "empty message",
			message: []byte{},
		},
		{
			name:    "large message",
			message: make([]byte, 10000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			signature, err := DefaultMPCSignerFunc(ctx, tt.message)

			require.NoError(t, err)
			assert.NotEmpty(t, signature)

			// Verify it's a valid hex string
			decoded, err := hex.DecodeString(signature)
			require.NoError(t, err)
			assert.Len(t, decoded, 32, "SHA256 should produce 32 bytes")

			// Verify it's deterministic
			signature2, err := DefaultMPCSignerFunc(ctx, tt.message)
			require.NoError(t, err)
			assert.Equal(t, signature, signature2, "same message should produce same signature")
		})
	}
}

func TestMPCSigner_InterfaceCompliance(t *testing.T) {
	// Verify that MPCSigner implements the Signer interface
	var _ Signer = (*MPCSigner)(nil)

	// Test that it can be used through the interface
	var signer Signer
	config := MPCConfig{
		APIKey:     "interface-test-key",
		SignerFunc: DefaultMPCSignerFunc,
	}

	mpcSigner, err := NewMPCSigner(config)
	require.NoError(t, err)

	signer = mpcSigner

	ctx := context.Background()
	req := SignRequest{
		Method: "POST",
		Path:   "/api/v1/orders",
		Body:   []byte(`{"test":"data"}`),
	}

	result, err := signer.Sign(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, "interface-test-key", result.Headers["X-API-KEY"])
	assert.NotEmpty(t, result.Headers["X-SIGNATURE"])
}

func TestMPCSigner_Middleware(t *testing.T) {
	// Test that the signer works with the Middleware function
	config := MPCConfig{
		APIKey:     "middleware-test-key",
		SignerFunc: DefaultMPCSignerFunc,
	}
	signer, err := NewMPCSigner(config)
	require.NoError(t, err)

	// Create a mock transport that captures the request
	var capturedRequest *http.Request
	mockTransport := &mockMPCRoundTripper{
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

	// Verify the headers were added
	require.NotNil(t, capturedRequest)
	assert.Equal(t, "middleware-test-key", capturedRequest.Header.Get("X-API-KEY"))
	assert.NotEmpty(t, capturedRequest.Header.Get("X-SIGNATURE"))
	assert.NotEmpty(t, capturedRequest.Header.Get("X-TIMESTAMP"))
}

func TestMPCSigner_MessageFormat(t *testing.T) {
	// Test that the message format is correct: timestamp + method + path + body
	var capturedMessage []byte
	config := MPCConfig{
		APIKey: "test-api-key",
		SignerFunc: func(ctx context.Context, message []byte) (string, error) {
			capturedMessage = message
			return "test-signature", nil
		},
	}
	signer, err := NewMPCSigner(config)
	require.NoError(t, err)

	ctx := context.Background()
	req := SignRequest{
		Method:    "POST",
		Path:      "/api/v1/orders",
		Body:      []byte(`{"symbol":"BTC-USD"}`),
		Timestamp: "1234567890",
	}

	_, err = signer.Sign(ctx, req)
	require.NoError(t, err)

	// Verify message format
	expectedMessage := "1234567890POST/api/v1/orders" + `{"symbol":"BTC-USD"}`
	assert.Equal(t, expectedMessage, string(capturedMessage))
}

func TestMPCSigner_EmptyPathAndBody(t *testing.T) {
	// Test edge case with empty path and body
	var capturedMessage []byte
	config := MPCConfig{
		APIKey: "test-api-key",
		SignerFunc: func(ctx context.Context, message []byte) (string, error) {
			capturedMessage = message
			return "test-signature", nil
		},
	}
	signer, err := NewMPCSigner(config)
	require.NoError(t, err)

	ctx := context.Background()
	req := SignRequest{
		Method:    "GET",
		Path:      "",
		Body:      nil,
		Timestamp: "1234567890",
	}

	_, err = signer.Sign(ctx, req)
	require.NoError(t, err)

	// Should be: timestamp + method (no path, no body)
	expectedMessage := "1234567890GET"
	assert.Equal(t, expectedMessage, string(capturedMessage))
}

// mockMPCRoundTripper is a mock http.RoundTripper for testing
type mockMPCRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockMPCRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}
