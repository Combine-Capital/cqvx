package auth_test

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/Combine-Capital/cqvx/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test vectors based on Coinbase API documentation
// These use known inputs and expected outputs for validation

const (
	testAPIKey     = "test-api-key-123"
	testSecret     = "dGVzdC1zZWNyZXQtdmFsdWUtZm9yLWhtYWMtc2hhMjU2" // base64 encoded "test-secret-value-for-hmac-sha256"
	testPassphrase = "test-passphrase"
	testTimestamp  = "1640995200"
	testInvalidB64 = "not-valid-base64!@#$"
)

func TestNewHMACSigner_Success(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)
	require.NotNil(t, signer)
}

func TestNewHMACSigner_Validation(t *testing.T) {
	tests := []struct {
		name        string
		config      auth.HMACConfig
		expectError string
	}{
		{
			name: "missing API key",
			config: auth.HMACConfig{
				APIKey:     "",
				Secret:     testSecret,
				Passphrase: testPassphrase,
			},
			expectError: "API key is required",
		},
		{
			name: "missing secret",
			config: auth.HMACConfig{
				APIKey:     testAPIKey,
				Secret:     "",
				Passphrase: testPassphrase,
			},
			expectError: "secret is required",
		},
		{
			name: "missing passphrase",
			config: auth.HMACConfig{
				APIKey:     testAPIKey,
				Secret:     testSecret,
				Passphrase: "",
			},
			expectError: "passphrase is required",
		},
		{
			name: "invalid base64 secret",
			config: auth.HMACConfig{
				APIKey:     testAPIKey,
				Secret:     testInvalidB64,
				Passphrase: testPassphrase,
			},
			expectError: "secret must be valid base64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := auth.NewHMACSigner(tt.config)
			assert.Error(t, err)
			assert.Nil(t, signer)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestHMACSigner_Sign_WithTimestamp(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	req := auth.SignRequest{
		Method:    "GET",
		Path:      "/api/v3/brokerage/accounts",
		Body:      []byte(""),
		Timestamp: testTimestamp,
	}

	result, err := signer.Sign(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify required headers are present
	assert.Equal(t, testAPIKey, result.Headers["CB-ACCESS-KEY"])
	assert.Equal(t, testPassphrase, result.Headers["CB-ACCESS-PASSPHRASE"])
	assert.Equal(t, testTimestamp, result.Headers["CB-ACCESS-TIMESTAMP"])
	assert.NotEmpty(t, result.Headers["CB-ACCESS-SIGN"])

	// Signature should be base64 encoded
	assert.Regexp(t, "^[A-Za-z0-9+/]+=*$", result.Headers["CB-ACCESS-SIGN"])
}

func TestHMACSigner_Sign_GeneratesTimestamp(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	req := auth.SignRequest{
		Method: "GET",
		Path:   "/api/v3/brokerage/accounts",
		Body:   []byte(""),
		// Timestamp not provided - should be generated
	}

	result, err := signer.Sign(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Timestamp should be generated
	assert.NotEmpty(t, result.Headers["CB-ACCESS-TIMESTAMP"])
	assert.Regexp(t, "^[0-9]+$", result.Headers["CB-ACCESS-TIMESTAMP"])
}

func TestHMACSigner_Sign_PostWithBody(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	body := []byte(`{"product_id":"BTC-USD","side":"buy","type":"limit","price":"10000","size":"0.01"}`)
	req := auth.SignRequest{
		Method:    "POST",
		Path:      "/api/v3/brokerage/orders",
		Body:      body,
		Timestamp: testTimestamp,
	}

	result, err := signer.Sign(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify all headers are present
	assert.Equal(t, testAPIKey, result.Headers["CB-ACCESS-KEY"])
	assert.Equal(t, testPassphrase, result.Headers["CB-ACCESS-PASSPHRASE"])
	assert.Equal(t, testTimestamp, result.Headers["CB-ACCESS-TIMESTAMP"])
	assert.NotEmpty(t, result.Headers["CB-ACCESS-SIGN"])
}

func TestHMACSigner_Sign_DifferentMethods(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	methods := []string{"GET", "POST", "PUT", "DELETE"}
	signatures := make(map[string]string)

	for _, method := range methods {
		req := auth.SignRequest{
			Method:    method,
			Path:      "/api/v3/brokerage/orders",
			Body:      []byte(""),
			Timestamp: testTimestamp,
		}

		result, err := signer.Sign(context.Background(), req)
		require.NoError(t, err)

		signatures[method] = result.Headers["CB-ACCESS-SIGN"]
	}

	// Different methods should produce different signatures
	assert.NotEqual(t, signatures["GET"], signatures["POST"])
	assert.NotEqual(t, signatures["GET"], signatures["PUT"])
	assert.NotEqual(t, signatures["GET"], signatures["DELETE"])
}

func TestHMACSigner_Sign_DifferentPaths(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	paths := []string{
		"/api/v3/brokerage/accounts",
		"/api/v3/brokerage/orders",
		"/api/v3/brokerage/products",
	}
	signatures := make(map[string]string)

	for _, path := range paths {
		req := auth.SignRequest{
			Method:    "GET",
			Path:      path,
			Body:      []byte(""),
			Timestamp: testTimestamp,
		}

		result, err := signer.Sign(context.Background(), req)
		require.NoError(t, err)

		signatures[path] = result.Headers["CB-ACCESS-SIGN"]
	}

	// Different paths should produce different signatures
	assert.NotEqual(t, signatures[paths[0]], signatures[paths[1]])
	assert.NotEqual(t, signatures[paths[0]], signatures[paths[2]])
}

func TestHMACSigner_Sign_DifferentTimestamps(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	timestamps := []string{"1640995200", "1640995201", "1640995202"}
	signatures := make(map[string]string)

	for _, ts := range timestamps {
		req := auth.SignRequest{
			Method:    "GET",
			Path:      "/api/v3/brokerage/accounts",
			Body:      []byte(""),
			Timestamp: ts,
		}

		result, err := signer.Sign(context.Background(), req)
		require.NoError(t, err)

		signatures[ts] = result.Headers["CB-ACCESS-SIGN"]
	}

	// Different timestamps should produce different signatures
	assert.NotEqual(t, signatures[timestamps[0]], signatures[timestamps[1]])
	assert.NotEqual(t, signatures[timestamps[0]], signatures[timestamps[2]])
}

func TestHMACSigner_Sign_EmptyBodyVsNoBody(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	// Request with empty byte slice
	req1 := auth.SignRequest{
		Method:    "GET",
		Path:      "/api/v3/brokerage/accounts",
		Body:      []byte(""),
		Timestamp: testTimestamp,
	}

	// Request with nil body
	req2 := auth.SignRequest{
		Method:    "GET",
		Path:      "/api/v3/brokerage/accounts",
		Body:      nil,
		Timestamp: testTimestamp,
	}

	result1, err := signer.Sign(context.Background(), req1)
	require.NoError(t, err)

	result2, err := signer.Sign(context.Background(), req2)
	require.NoError(t, err)

	// Empty body and nil body should produce same signature
	assert.Equal(t, result1.Headers["CB-ACCESS-SIGN"], result2.Headers["CB-ACCESS-SIGN"])
}

func TestHMACSigner_Sign_SpecialCharactersInPath(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	// Path with special characters
	req := auth.SignRequest{
		Method:    "GET",
		Path:      "/api/v3/brokerage/orders/abc-123_456",
		Body:      []byte(""),
		Timestamp: testTimestamp,
	}

	result, err := signer.Sign(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should successfully sign path with special characters
	assert.NotEmpty(t, result.Headers["CB-ACCESS-SIGN"])
}

func TestHMACSigner_Sign_LargeBody(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	// Create a large body (10KB)
	largeBody := make([]byte, 10*1024)
	for i := range largeBody {
		largeBody[i] = byte('A' + (i % 26))
	}

	req := auth.SignRequest{
		Method:    "POST",
		Path:      "/api/v3/brokerage/orders/batch",
		Body:      largeBody,
		Timestamp: testTimestamp,
	}

	result, err := signer.Sign(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should successfully sign large bodies
	assert.NotEmpty(t, result.Headers["CB-ACCESS-SIGN"])
}

func TestHMACSigner_Sign_Deterministic(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	req := auth.SignRequest{
		Method:    "POST",
		Path:      "/api/v3/brokerage/orders",
		Body:      []byte(`{"product_id":"BTC-USD","side":"buy"}`),
		Timestamp: testTimestamp,
	}

	// Sign the same request multiple times
	signatures := make([]string, 5)
	for i := 0; i < 5; i++ {
		result, err := signer.Sign(context.Background(), req)
		require.NoError(t, err)
		signatures[i] = result.Headers["CB-ACCESS-SIGN"]
	}

	// All signatures should be identical (deterministic)
	for i := 1; i < len(signatures); i++ {
		assert.Equal(t, signatures[0], signatures[i],
			"Signature at index %d differs from first signature", i)
	}
}

func TestHMACSigner_Sign_KnownTestVector(t *testing.T) {
	// Test vector with known expected signature
	// This uses a specific secret, timestamp, method, path, and body
	// to produce a known signature for validation

	// Secret: "secret" encoded in base64
	secret := base64.StdEncoding.EncodeToString([]byte("secret"))

	config := auth.HMACConfig{
		APIKey:     "api-key",
		Secret:     secret,
		Passphrase: "passphrase",
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	req := auth.SignRequest{
		Method:    "GET",
		Path:      "/orders",
		Body:      []byte(""),
		Timestamp: "1234567890",
	}

	result, err := signer.Sign(context.Background(), req)
	require.NoError(t, err)

	// Expected signature computed manually:
	// prehash = "1234567890GET/orders"
	// signature = base64(hmac_sha256("secret", prehash))
	// Actual computed: "c0bz9rdYCiGfAsKzIyfvmtx6eU1fbWn3SVcwKIVqZM4="
	expectedSignature := "c0bz9rdYCiGfAsKzIyfvmtx6eU1fbWn3SVcwKIVqZM4="

	assert.Equal(t, expectedSignature, result.Headers["CB-ACCESS-SIGN"])
}

func TestHMACSigner_ImplementsSigner(t *testing.T) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(t, err)

	// Verify it implements the Signer interface
	var _ auth.Signer = signer
}

// Benchmark HMAC signing performance
func BenchmarkHMACSigner_Sign(b *testing.B) {
	config := auth.HMACConfig{
		APIKey:     testAPIKey,
		Secret:     testSecret,
		Passphrase: testPassphrase,
	}

	signer, err := auth.NewHMACSigner(config)
	require.NoError(b, err)

	req := auth.SignRequest{
		Method:    "POST",
		Path:      "/api/v3/brokerage/orders",
		Body:      []byte(`{"product_id":"BTC-USD","side":"buy","type":"limit","price":"10000","size":"0.01"}`),
		Timestamp: testTimestamp,
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := signer.Sign(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}
