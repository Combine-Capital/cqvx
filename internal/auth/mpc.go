// Package auth provides authentication interfaces and implementations for venue clients.
package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

// MPCConfig contains configuration for MPC (Multi-Party Computation) authentication.
type MPCConfig struct {
	// APIKey is the API key identifier for Fordefi
	APIKey string

	// SignerFunc is a function that performs MPC signing.
	// It takes the message to sign and returns the signature.
	// For MVP, this can be a stub that returns a pre-configured signature.
	// Post-MVP, this will integrate with an actual MPC signing service.
	//
	// The function signature is:
	//   func(ctx context.Context, message []byte) (signature string, error)
	SignerFunc func(ctx context.Context, message []byte) (string, error)
}

// MPCSigner implements MPC (Multi-Party Computation) authentication for Fordefi.
// MPC signing involves distributed key management where multiple parties
// collaboratively sign requests without any single party having access to
// the complete private key.
//
// For MVP, this is a stub implementation that uses a configurable signing
// function. Post-MVP enhancements will integrate with actual MPC providers
// like Fireblocks, Fordefi's native MPC, or other custody solutions.
//
// Required headers (Fordefi-specific):
//   - X-API-KEY: The API key identifier
//   - X-TIMESTAMP: Unix timestamp in milliseconds
//   - X-SIGNATURE: The MPC-generated signature
//
// The signature is computed over:
//
//	message = timestamp + method + path + body
//
// Thread-safe: This implementation is safe for concurrent use if the
// provided SignerFunc is thread-safe.
type MPCSigner struct {
	config MPCConfig
}

// NewMPCSigner creates a new MPC signer for Fordefi.
// The SignerFunc must be provided to perform actual signing operations.
//
// Example stub SignerFunc for testing:
//
//	signerFunc := func(ctx context.Context, message []byte) (string, error) {
//	    // For MVP, return a deterministic "signature" (e.g., hash of message)
//	    hash := sha256.Sum256(message)
//	    return hex.EncodeToString(hash[:]), nil
//	}
func NewMPCSigner(config MPCConfig) (*MPCSigner, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}
	if config.SignerFunc == nil {
		return nil, fmt.Errorf("signer function is required")
	}

	return &MPCSigner{
		config: config,
	}, nil
}

// Sign generates MPC authentication headers for a Fordefi API request.
// It constructs the message to sign, calls the configured SignerFunc to
// obtain the signature, and returns the required authentication headers.
//
// The message format is:
//
//	message = timestamp + method + path + body
//
// Where timestamp is Unix time in milliseconds.
//
// Returns an error if signature generation fails.
func (s *MPCSigner) Sign(ctx context.Context, req SignRequest) (*SignResult, error) {
	// Generate timestamp if not provided (Unix milliseconds as string)
	timestamp := req.Timestamp
	if timestamp == "" {
		timestamp = strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	// Construct message to sign: timestamp + method + path + body
	body := string(req.Body)
	message := timestamp + req.Method + req.Path + body

	// Call the MPC signer function to get the signature
	signature, err := s.config.SignerFunc(ctx, []byte(message))
	if err != nil {
		return nil, fmt.Errorf("MPC signing failed: %w", err)
	}

	// Return authentication headers
	return &SignResult{
		Headers: map[string]string{
			"X-API-KEY":   s.config.APIKey,
			"X-TIMESTAMP": timestamp,
			"X-SIGNATURE": signature,
		},
	}, nil
}

// DefaultMPCSignerFunc is a default stub implementation for testing.
// It returns the SHA256 hash of the message as a hex string.
// This should NOT be used in production - it's only for testing and MVP demonstration.
func DefaultMPCSignerFunc(ctx context.Context, message []byte) (string, error) {
	hash := sha256.Sum256(message)
	return hex.EncodeToString(hash[:]), nil
}

// Verify that MPCSigner implements the Signer interface
var _ Signer = (*MPCSigner)(nil)
