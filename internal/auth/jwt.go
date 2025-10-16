// Package auth provides authentication interfaces and implementations for venue clients.
package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig contains configuration for JWT (Coinbase Prime) authentication.
type JWTConfig struct {
	// KeyName is the API key name in the format "organizations/{org_id}/apiKeys/{key_id}"
	KeyName string

	// PrivateKey is the PEM-encoded EC private key
	PrivateKey string

	// ExpiresIn is the JWT expiration time in seconds (default: 120)
	ExpiresIn int64
}

// JWTSigner implements Coinbase Prime JWT authentication using ES256.
// It generates JSON Web Tokens according to Coinbase's JWT specification:
//
// JWT Header:
//   - alg: "ES256"
//   - typ: "JWT"
//   - kid: API key name
//   - nonce: Random hex string for replay protection
//
// JWT Claims:
//   - iss: "cdp"
//   - nbf: Current Unix timestamp
//   - exp: Current Unix timestamp + expiration (default 120 seconds)
//   - sub: API key name
//   - uri: "{METHOD} {HOST}{PATH}" (e.g., "GET api.coinbase.com/api/v3/brokerage/accounts")
//
// Thread-safe: This implementation is safe for concurrent use.
type JWTSigner struct {
	config     JWTConfig
	privateKey *ecdsa.PrivateKey
}

// NewJWTSigner creates a new JWT signer for Coinbase Prime.
// The private key must be a PEM-encoded EC private key.
func NewJWTSigner(config JWTConfig) (*JWTSigner, error) {
	if config.KeyName == "" {
		return nil, fmt.Errorf("key name is required")
	}
	if config.PrivateKey == "" {
		return nil, fmt.Errorf("private key is required")
	}

	// Set default expiration if not provided
	if config.ExpiresIn <= 0 {
		config.ExpiresIn = 120 // 2 minutes default
	}

	// Parse PEM-encoded private key
	privateKey, err := parseECPrivateKey(config.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return &JWTSigner{
		config:     config,
		privateKey: privateKey,
	}, nil
}

// Sign generates a JWT for authenticating Coinbase Prime API requests.
// It returns an Authorization: Bearer <JWT> header.
//
// The URI is constructed from the request method, a host (derived from headers or default),
// and the request path. Format: "{METHOD} {HOST}{PATH}"
//
// Returns an error if JWT generation fails.
func (s *JWTSigner) Sign(ctx context.Context, req SignRequest) (*SignResult, error) {
	// Extract host from headers or use default
	host := req.Headers.Get("Host")
	if host == "" {
		host = "api.coinbase.com" // Default host for Coinbase Prime
	}

	// Construct URI: "METHOD HOST/PATH"
	uri := fmt.Sprintf("%s %s%s", req.Method, host, req.Path)

	// Generate random nonce for replay protection
	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Current time for nbf and exp claims
	now := time.Now().Unix()

	// Create JWT claims
	claims := jwt.MapClaims{
		"iss": "cdp",
		"nbf": now,
		"exp": now + s.config.ExpiresIn,
		"sub": s.config.KeyName,
		"uri": uri,
	}

	// Create JWT with custom headers
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["kid"] = s.config.KeyName
	token.Header["nonce"] = nonce
	token.Header["typ"] = "JWT"

	// Sign the token
	tokenString, err := token.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign JWT: %w", err)
	}

	// Return Authorization: Bearer header
	return &SignResult{
		Headers: map[string]string{
			"Authorization": "Bearer " + tokenString,
		},
	}, nil
}

// parseECPrivateKey parses a PEM-encoded EC private key.
func parseECPrivateKey(pemKey string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemKey))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Try parsing as PKCS8
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		ecKey, ok := key.(*ecdsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("key is not an EC private key")
		}
		return ecKey, nil
	}

	// Try parsing as EC private key
	ecKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse EC private key: %w", err)
	}

	return ecKey, nil
}

// generateNonce generates a random 16-byte hex string for nonce.
func generateNonce() (string, error) {
	// Generate 16 random bytes
	nonce := make([]byte, 16)
	_, err := rand.Read(nonce)
	if err != nil {
		return "", err
	}

	// Convert to hex string (32 characters)
	return hex.EncodeToString(nonce), nil
}

// generateNonceInt generates a random large integer for nonce (alternative implementation).
// This is kept for compatibility with some JWT examples that use integer nonces.
func generateNonceInt() (string, error) {
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(63), nil)

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return n.String(), nil
}

// Verify that JWTSigner implements the Signer interface
var _ Signer = (*JWTSigner)(nil)
