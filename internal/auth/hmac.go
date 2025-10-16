// Package auth provides authentication interfaces and implementations for venue clients.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"
)

// HMACConfig contains configuration for HMAC-SHA256 authentication.
type HMACConfig struct {
	// APIKey is the Coinbase API key (CB-ACCESS-KEY header)
	APIKey string

	// Secret is the base64-encoded secret key for HMAC signing
	Secret string

	// Passphrase is the API passphrase (CB-ACCESS-PASSPHRASE header)
	Passphrase string
}

// HMACSigner implements Coinbase Exchange HMAC-SHA256 authentication.
// It generates signatures according to Coinbase's authentication specification:
//
//	CB-ACCESS-SIGN = base64(hmac_sha256(base64_decode(secret), timestamp + method + path + body))
//
// Required headers:
//   - CB-ACCESS-KEY: The API key
//   - CB-ACCESS-SIGN: The base64-encoded signature
//   - CB-ACCESS-TIMESTAMP: Unix timestamp in seconds
//   - CB-ACCESS-PASSPHRASE: The API passphrase
//
// Thread-safe: This implementation is safe for concurrent use.
type HMACSigner struct {
	config HMACConfig
}

// NewHMACSigner creates a new HMAC-SHA256 signer for Coinbase Exchange.
// The secret must be base64-encoded as provided by Coinbase.
func NewHMACSigner(config HMACConfig) (*HMACSigner, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if config.Secret == "" {
		return nil, fmt.Errorf("secret is required")
	}
	if config.Passphrase == "" {
		return nil, fmt.Errorf("passphrase is required")
	}

	// Validate that secret is valid base64
	_, err := base64.StdEncoding.DecodeString(config.Secret)
	if err != nil {
		return nil, fmt.Errorf("secret must be valid base64: %w", err)
	}

	return &HMACSigner{
		config: config,
	}, nil
}

// Sign generates HMAC-SHA256 authentication headers for a Coinbase API request.
// It returns headers that must be added to the HTTP request.
//
// The signature is computed as:
//
//	prehash = timestamp + method + path + body
//	signature = base64(HMAC-SHA256(base64_decode(secret), prehash))
//
// Returns an error if signature generation fails.
func (s *HMACSigner) Sign(ctx context.Context, req SignRequest) (*SignResult, error) {
	// Generate timestamp if not provided (Unix seconds as string)
	timestamp := req.Timestamp
	if timestamp == "" {
		timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	}

	// Construct prehash string: timestamp + method + path + body
	// For GET requests with no body, body should be empty string
	body := string(req.Body)
	prehash := timestamp + req.Method + req.Path + body

	// Decode the base64-encoded secret
	decodedSecret, err := base64.StdEncoding.DecodeString(s.config.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to decode secret: %w", err)
	}

	// Compute HMAC-SHA256
	h := hmac.New(sha256.New, decodedSecret)
	h.Write([]byte(prehash))
	signature := h.Sum(nil)

	// Encode signature as base64
	signatureB64 := base64.StdEncoding.EncodeToString(signature)

	// Return authentication headers
	return &SignResult{
		Headers: map[string]string{
			"CB-ACCESS-KEY":        s.config.APIKey,
			"CB-ACCESS-SIGN":       signatureB64,
			"CB-ACCESS-TIMESTAMP":  timestamp,
			"CB-ACCESS-PASSPHRASE": s.config.Passphrase,
		},
	}, nil
}

// Verify that HMACSigner implements the Signer interface
var _ Signer = (*HMACSigner)(nil)
