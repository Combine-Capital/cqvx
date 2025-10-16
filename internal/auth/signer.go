// Package auth provides authentication interfaces and implementations for venue clients.
// It defines the Signer interface and common authentication patterns for HTTP requests.
package auth

import (
	"bytes"
	"context"
	"io"
	"net/http"
)

// SignRequest represents an HTTP request to be signed.
// It contains all information needed for various authentication schemes.
type SignRequest struct {
	// Method is the HTTP method (GET, POST, etc.)
	Method string

	// Path is the request path (e.g., "/orders")
	Path string

	// Body is the request body (may be empty for GET requests)
	Body []byte

	// Timestamp is the request timestamp (Unix seconds or RFC3339, depending on venue)
	// Signers should generate this if not provided
	Timestamp string

	// Headers contains existing request headers that may be needed for signing
	Headers http.Header
}

// SignResult contains the authentication information to be added to the request.
type SignResult struct {
	// Headers contains authentication headers to add to the request
	Headers map[string]string

	// QueryParams contains authentication query parameters to add to the request
	QueryParams map[string]string
}

// Signer defines the interface for request authentication.
// Implementations sign HTTP requests according to venue-specific requirements.
//
// Thread-safety: Implementations must be safe for concurrent use.
//
// Example implementations:
//   - HMAC-SHA256 for Coinbase Exchange
//   - JWT for Coinbase Prime
//   - Bearer token for FalconX
//   - MPC signing for Fordefi
type Signer interface {
	// Sign generates authentication information for an HTTP request.
	// It returns SignResult containing headers and/or query parameters to add.
	//
	// The context can be used for cancellation and to pass request-scoped values.
	//
	// Returns an error if signing fails (e.g., invalid credentials, key not found).
	Sign(ctx context.Context, req SignRequest) (*SignResult, error)
}

// Middleware creates an HTTP middleware function that applies authentication
// to outgoing requests using the provided Signer.
//
// This middleware can be used with standard http.Client through RoundTripper,
// or with CQI HTTP client if it supports middleware functions.
//
// Example usage:
//
//	signer := NewHMACSigner(config)
//	client := &http.Client{
//	    Transport: Middleware(signer, http.DefaultTransport),
//	}
func Middleware(signer Signer, next http.RoundTripper) http.RoundTripper {
	if next == nil {
		next = http.DefaultTransport
	}
	return &authTransport{
		signer: signer,
		next:   next,
	}
}

// authTransport is an http.RoundTripper that applies authentication to requests.
type authTransport struct {
	signer Signer
	next   http.RoundTripper
}

// RoundTrip implements http.RoundTripper by signing the request before forwarding it.
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read the request body if present (for signing)
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, err
		}
		// Restore the body so it can be read again
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	// Build sign request
	signReq := SignRequest{
		Method:    req.Method,
		Path:      req.URL.Path,
		Body:      body,
		Headers:   req.Header,
		Timestamp: "", // Let signer generate timestamp
	}

	// Sign the request
	result, err := t.signer.Sign(req.Context(), signReq)
	if err != nil {
		return nil, err
	}

	// Apply authentication headers to the original request
	for key, value := range result.Headers {
		req.Header.Set(key, value)
	}

	// Apply authentication query parameters
	if len(result.QueryParams) > 0 {
		q := req.URL.Query()
		for key, value := range result.QueryParams {
			q.Set(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	// Forward the signed request
	return t.next.RoundTrip(req)
}
