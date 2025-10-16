package prime

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// PrimeError represents an error response from the Coinbase Prime API.
// Prime returns errors with status codes and error messages.
type PrimeError struct {
	Message    string                 `json:"message"`
	Code       string                 `json:"code"`
	Details    map[string]interface{} `json:"details"`
	StatusCode int                    `json:"status_code"`
}

// NormalizeError converts a Coinbase Prime API error response to a structured error.
// It parses the error response and classifies it based on HTTP status code and error message.
//
// Error Classification:
//   - 401/403: Authentication/Authorization errors (Permanent)
//   - 429: Rate limit errors (RateLimit)
//   - 400: Invalid request errors (Permanent if client error, otherwise Temporary)
//   - 404: Not found errors (Permanent)
//   - 500/503: Server errors (Temporary)
//   - Other: Temporary by default
//
// Returns an error with appropriate classification and original error details.
func NormalizeError(statusCode int, body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("prime api error: status %d (no body)", statusCode)
	}

	var primeErr PrimeError
	if err := json.Unmarshal(body, &primeErr); err != nil {
		// If we can't parse the error, return raw body
		return fmt.Errorf("prime api error: status %d: %s", statusCode, string(body))
	}

	// Construct error message with all available details
	msg := formatErrorMessage(primeErr)

	// Classify error based on status code
	return classifyError(statusCode, msg, &primeErr)
}

// formatErrorMessage constructs a comprehensive error message from Prime error fields.
func formatErrorMessage(primeErr PrimeError) string {
	msg := "prime api error"

	if primeErr.Code != "" {
		msg = fmt.Sprintf("%s: %s", msg, primeErr.Code)
	}

	if primeErr.Message != "" {
		if primeErr.Code == "" {
			msg = fmt.Sprintf("%s: %s", msg, primeErr.Message)
		} else {
			msg = fmt.Sprintf("%s (%s)", msg, primeErr.Message)
		}
	}

	// Add details if available
	if len(primeErr.Details) > 0 {
		detailsJSON, err := json.Marshal(primeErr.Details)
		if err == nil {
			msg = fmt.Sprintf("%s [details: %s]", msg, string(detailsJSON))
		}
	}

	return msg
}

// classifyError determines the error type based on HTTP status code and error content.
func classifyError(statusCode int, msg string, primeErr *PrimeError) error {
	baseErr := fmt.Errorf("%s (status: %d)", msg, statusCode)

	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		// Authentication/authorization failures are permanent
		return &PermanentError{Err: baseErr, Code: primeErr.Code}

	case http.StatusTooManyRequests:
		// Rate limit errors
		return &RateLimitError{Err: baseErr, Code: primeErr.Code}

	case http.StatusBadRequest:
		// Bad request errors - check if it's a client error or validation error
		if isClientError(primeErr) {
			return &PermanentError{Err: baseErr, Code: primeErr.Code}
		}
		return &TemporaryError{Err: baseErr, Code: primeErr.Code}

	case http.StatusNotFound:
		// Not found errors are permanent
		return &PermanentError{Err: baseErr, Code: primeErr.Code}

	case http.StatusInternalServerError, http.StatusServiceUnavailable, http.StatusBadGateway:
		// Server errors are temporary (retry may succeed)
		return &TemporaryError{Err: baseErr, Code: primeErr.Code}

	default:
		// Unknown errors are treated as temporary by default
		return &TemporaryError{Err: baseErr, Code: primeErr.Code}
	}
}

// isClientError determines if an error is a client-side error that should not be retried.
func isClientError(primeErr *PrimeError) bool {
	// Common client error codes that indicate permanent failures
	clientErrorCodes := map[string]bool{
		"INVALID_ARGUMENT":     true,
		"INVALID_PRODUCT":      true,
		"INVALID_ORDER":        true,
		"INVALID_ORDER_ID":     true,
		"INVALID_PORTFOLIO":    true,
		"INVALID_PORTFOLIO_ID": true,
		"INSUFFICIENT_FUNDS":   true,
		"ORDER_NOT_FOUND":      true,
		"VALIDATION_ERROR":     true,
	}

	return clientErrorCodes[primeErr.Code]
}

// PermanentError represents an error that should not be retried.
type PermanentError struct {
	Err  error
	Code string
}

func (e *PermanentError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("permanent error [%s]: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("permanent error: %v", e.Err)
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

// TemporaryError represents an error that may succeed if retried.
type TemporaryError struct {
	Err  error
	Code string
}

func (e *TemporaryError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("temporary error [%s]: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("temporary error: %v", e.Err)
}

func (e *TemporaryError) Unwrap() error {
	return e.Err
}

// Temporary returns true to indicate this error is temporary.
func (e *TemporaryError) Temporary() bool {
	return true
}

// RateLimitError represents a rate limit error.
type RateLimitError struct {
	Err  error
	Code string
}

func (e *RateLimitError) Error() string {
	if e.Code != "" {
		return fmt.Sprintf("rate limit error [%s]: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("rate limit error: %v", e.Err)
}

func (e *RateLimitError) Unwrap() error {
	return e.Err
}

// Temporary returns true since rate limit errors can be retried after backoff.
func (e *RateLimitError) Temporary() bool {
	return true
}

// IsTemporary checks if an error is temporary and can be retried.
func IsTemporary(err error) bool {
	type temporary interface {
		Temporary() bool
	}

	if t, ok := err.(temporary); ok {
		return t.Temporary()
	}

	return false
}

// IsPermanent checks if an error is permanent and should not be retried.
func IsPermanent(err error) bool {
	_, ok := err.(*PermanentError)
	return ok
}

// IsRateLimit checks if an error is a rate limit error.
func IsRateLimit(err error) bool {
	_, ok := err.(*RateLimitError)
	return ok
}
