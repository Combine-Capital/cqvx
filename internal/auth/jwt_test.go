package auth_test

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/Combine-Capital/cqvx/internal/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Generate a test EC private key for testing
func generateTestECKey(t *testing.T) string {
	t.Helper()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	x509Encoded, err := x509.MarshalECPrivateKey(privateKey)
	require.NoError(t, err)

	pemEncoded := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: x509Encoded,
	})

	return string(pemEncoded)
}

const (
	testKeyName   = "organizations/test-org/apiKeys/test-key-123"
	testExpiresIn = 120
)

func TestNewJWTSigner_Success(t *testing.T) {
	privateKey := generateTestECKey(t)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		ExpiresIn:  testExpiresIn,
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(t, err)
	require.NotNil(t, signer)
}

func TestNewJWTSigner_DefaultExpiration(t *testing.T) {
	privateKey := generateTestECKey(t)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		// ExpiresIn not set - should default to 120
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(t, err)
	require.NotNil(t, signer)
}

func TestNewJWTSigner_Validation(t *testing.T) {
	privateKey := generateTestECKey(t)

	tests := []struct {
		name        string
		config      auth.JWTConfig
		expectError string
	}{
		{
			name: "missing key name",
			config: auth.JWTConfig{
				KeyName:    "",
				PrivateKey: privateKey,
				ExpiresIn:  testExpiresIn,
			},
			expectError: "key name is required",
		},
		{
			name: "missing private key",
			config: auth.JWTConfig{
				KeyName:    testKeyName,
				PrivateKey: "",
				ExpiresIn:  testExpiresIn,
			},
			expectError: "private key is required",
		},
		{
			name: "invalid PEM format",
			config: auth.JWTConfig{
				KeyName:    testKeyName,
				PrivateKey: "not-a-valid-pem-key",
				ExpiresIn:  testExpiresIn,
			},
			expectError: "failed to parse private key",
		},
		{
			name: "invalid key content",
			config: auth.JWTConfig{
				KeyName:    testKeyName,
				PrivateKey: "-----BEGIN EC PRIVATE KEY-----\ninvalid-content\n-----END EC PRIVATE KEY-----",
				ExpiresIn:  testExpiresIn,
			},
			expectError: "failed to parse private key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signer, err := auth.NewJWTSigner(tt.config)
			assert.Error(t, err)
			assert.Nil(t, signer)
			assert.Contains(t, err.Error(), tt.expectError)
		})
	}
}

func TestJWTSigner_Sign_GeneratesValidJWT(t *testing.T) {
	privateKey := generateTestECKey(t)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		ExpiresIn:  testExpiresIn,
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(t, err)

	req := auth.SignRequest{
		Method: "GET",
		Path:   "/api/v3/brokerage/accounts",
	}

	result, err := signer.Sign(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Verify Authorization header is present
	authHeader, ok := result.Headers["Authorization"]
	assert.True(t, ok)
	assert.True(t, strings.HasPrefix(authHeader, "Bearer "))

	// Extract JWT token
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	assert.NotEmpty(t, tokenString)

	// Parse JWT token (without verification for structure inspection)
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	require.NoError(t, err)

	// Verify JWT header
	assert.Equal(t, "ES256", token.Method.Alg())
	assert.Equal(t, "JWT", token.Header["typ"])
	assert.Equal(t, testKeyName, token.Header["kid"])
	assert.NotEmpty(t, token.Header["nonce"])

	// Verify JWT claims
	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)
	assert.Equal(t, "cdp", claims["iss"])
	assert.Equal(t, testKeyName, claims["sub"])
	assert.NotNil(t, claims["nbf"])
	assert.NotNil(t, claims["exp"])
	assert.NotEmpty(t, claims["uri"])
}

func TestJWTSigner_Sign_URIConstruction(t *testing.T) {
	privateKey := generateTestECKey(t)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		ExpiresIn:  testExpiresIn,
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		method      string
		path        string
		expectedURI string
	}{
		{
			name:        "GET request",
			method:      "GET",
			path:        "/api/v3/brokerage/accounts",
			expectedURI: "GET api.coinbase.com/api/v3/brokerage/accounts",
		},
		{
			name:        "POST request",
			method:      "POST",
			path:        "/api/v3/brokerage/orders",
			expectedURI: "POST api.coinbase.com/api/v3/brokerage/orders",
		},
		{
			name:        "DELETE request",
			method:      "DELETE",
			path:        "/api/v3/brokerage/orders/123",
			expectedURI: "DELETE api.coinbase.com/api/v3/brokerage/orders/123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := auth.SignRequest{
				Method: tt.method,
				Path:   tt.path,
			}

			result, err := signer.Sign(context.Background(), req)
			require.NoError(t, err)

			// Extract and parse JWT
			tokenString := strings.TrimPrefix(result.Headers["Authorization"], "Bearer ")
			token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
			require.NoError(t, err)

			claims, ok := token.Claims.(jwt.MapClaims)
			require.True(t, ok)

			assert.Equal(t, tt.expectedURI, claims["uri"])
		})
	}
}

func TestJWTSigner_Sign_ExpirationTime(t *testing.T) {
	privateKey := generateTestECKey(t)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		ExpiresIn:  60, // 1 minute
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(t, err)

	req := auth.SignRequest{
		Method: "GET",
		Path:   "/api/v3/brokerage/accounts",
	}

	beforeSign := time.Now().Unix()
	result, err := signer.Sign(context.Background(), req)
	afterSign := time.Now().Unix()
	require.NoError(t, err)

	// Extract and parse JWT
	tokenString := strings.TrimPrefix(result.Headers["Authorization"], "Bearer ")
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// Verify nbf (not before) is current time
	nbf, ok := claims["nbf"].(float64)
	require.True(t, ok)
	assert.GreaterOrEqual(t, int64(nbf), beforeSign)
	assert.LessOrEqual(t, int64(nbf), afterSign)

	// Verify exp (expiration) is nbf + ExpiresIn
	exp, ok := claims["exp"].(float64)
	require.True(t, ok)
	assert.Equal(t, int64(nbf)+60, int64(exp))
}

func TestJWTSigner_Sign_NonceUniqueness(t *testing.T) {
	privateKey := generateTestECKey(t)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		ExpiresIn:  testExpiresIn,
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(t, err)

	req := auth.SignRequest{
		Method: "GET",
		Path:   "/api/v3/brokerage/accounts",
	}

	// Generate multiple tokens
	nonces := make(map[string]bool)
	for i := 0; i < 100; i++ {
		result, err := signer.Sign(context.Background(), req)
		require.NoError(t, err)

		// Extract and parse JWT
		tokenString := strings.TrimPrefix(result.Headers["Authorization"], "Bearer ")
		token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
		require.NoError(t, err)

		nonce, ok := token.Header["nonce"].(string)
		require.True(t, ok)
		require.NotEmpty(t, nonce)

		// Verify nonce is unique
		assert.False(t, nonces[nonce], "Duplicate nonce generated: %s", nonce)
		nonces[nonce] = true
	}

	// All nonces should be unique
	assert.Len(t, nonces, 100)
}

func TestJWTSigner_Sign_DifferentRequests(t *testing.T) {
	privateKey := generateTestECKey(t)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		ExpiresIn:  testExpiresIn,
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(t, err)

	requests := []auth.SignRequest{
		{Method: "GET", Path: "/api/v3/brokerage/accounts"},
		{Method: "POST", Path: "/api/v3/brokerage/orders"},
		{Method: "DELETE", Path: "/api/v3/brokerage/orders/123"},
	}

	tokens := make([]string, len(requests))
	for i, req := range requests {
		result, err := signer.Sign(context.Background(), req)
		require.NoError(t, err)

		tokens[i] = result.Headers["Authorization"]
	}

	// All tokens should be different
	assert.NotEqual(t, tokens[0], tokens[1])
	assert.NotEqual(t, tokens[0], tokens[2])
	assert.NotEqual(t, tokens[1], tokens[2])
}

func TestJWTSigner_Sign_WithCustomHost(t *testing.T) {
	privateKey := generateTestECKey(t)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		ExpiresIn:  testExpiresIn,
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(t, err)

	// Create request with custom Host header
	req := auth.SignRequest{
		Method:  "GET",
		Path:    "/api/v1/orders",
		Headers: make(map[string][]string),
	}
	req.Headers.Set("Host", "prime.coinbase.com")

	result, err := signer.Sign(context.Background(), req)
	require.NoError(t, err)

	// Extract and parse JWT
	tokenString := strings.TrimPrefix(result.Headers["Authorization"], "Bearer ")
	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	require.NoError(t, err)

	claims, ok := token.Claims.(jwt.MapClaims)
	require.True(t, ok)

	// URI should use custom host
	expectedURI := "GET prime.coinbase.com/api/v1/orders"
	assert.Equal(t, expectedURI, claims["uri"])
}

func TestJWTSigner_Sign_TokenStructure(t *testing.T) {
	privateKey := generateTestECKey(t)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		ExpiresIn:  testExpiresIn,
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(t, err)

	req := auth.SignRequest{
		Method: "GET",
		Path:   "/api/v3/brokerage/accounts",
	}

	result, err := signer.Sign(context.Background(), req)
	require.NoError(t, err)

	tokenString := strings.TrimPrefix(result.Headers["Authorization"], "Bearer ")

	// JWT should have 3 parts: header.payload.signature
	parts := strings.Split(tokenString, ".")
	assert.Len(t, parts, 3, "JWT should have 3 parts")

	// All parts should be non-empty
	for i, part := range parts {
		assert.NotEmpty(t, part, "JWT part %d should not be empty", i)
	}
}

func TestJWTSigner_ImplementsSigner(t *testing.T) {
	privateKey := generateTestECKey(t)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		ExpiresIn:  testExpiresIn,
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(t, err)

	// Verify it implements the Signer interface
	var _ auth.Signer = signer
}

// Benchmark JWT signing performance
func BenchmarkJWTSigner_Sign(b *testing.B) {
	// Generate test key inline for benchmark
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(b, err)

	x509Encoded, err := x509.MarshalECPrivateKey(ecKey)
	require.NoError(b, err)

	pemEncoded := pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: x509Encoded,
	})
	privateKey := string(pemEncoded)

	config := auth.JWTConfig{
		KeyName:    testKeyName,
		PrivateKey: privateKey,
		ExpiresIn:  testExpiresIn,
	}

	signer, err := auth.NewJWTSigner(config)
	require.NoError(b, err)

	req := auth.SignRequest{
		Method: "GET",
		Path:   "/api/v3/brokerage/accounts",
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
