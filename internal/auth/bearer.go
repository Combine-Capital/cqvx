// Package auth provides authentication interfaces and implementations for venue clients.
package auth

import (
	"context"
	"fmt"
)

// BearerConfig contains configuration for Bearer token authentication.
type BearerConfig struct {
	// Token is the static bearer token for API authentication
	Token string
}

// BearerSigner implements Bearer token authentication for FalconX and similar venues.
// It adds a static Authorization: Bearer <token> header to all requests.
//
// This is the simplest authentication method, using a pre-generated token
// that doesn't change per request. The token is typically long-lived and
// rotated infrequently.
//
// Required header:
//   - Authorization: Bearer <token>
//
// Thread-safe: This implementation is safe for concurrent use.
type BearerSigner struct {
	config BearerConfig
}

// NewBearerSigner creates a new Bearer token signer.
// The token should be a valid bearer token provided by the venue.
func NewBearerSigner(config BearerConfig) (*BearerSigner, error) {
	if config.Token == "" {
		return nil, fmt.Errorf("token is required")
	}

	return &BearerSigner{
		config: config,
	}, nil
}

// Sign generates Bearer token authentication header for an API request.
// It returns the Authorization: Bearer <token> header.
//
// Unlike HMAC or JWT signing, this method doesn't compute any signature.
// It simply returns the pre-configured token as a bearer token header.
//
// The context parameter is accepted for interface compliance and future
// extensibility (e.g., token rotation), but is not currently used.
//
// Returns an error only if the signer is misconfigured (which should be
// caught during initialization).
func (s *BearerSigner) Sign(ctx context.Context, req SignRequest) (*SignResult, error) {
	// Return Authorization: Bearer header with the configured token
	return &SignResult{
		Headers: map[string]string{
			"Authorization": "Bearer " + s.config.Token,
		},
	}, nil
}

// Verify that BearerSigner implements the Signer interface
var _ Signer = (*BearerSigner)(nil)
