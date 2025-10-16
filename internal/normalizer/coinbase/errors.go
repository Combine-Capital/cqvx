package coinbase

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// CoinbaseError represents an error response from the Coinbase API.
// Coinbase returns errors in a consistent format with error codes and messages.
type CoinbaseError struct {
	Error           string `json:"error"`
	Message         string `json:"message"`
	ErrorDetails    string `json:"error_details"`
	PreviewFailure  string `json:"preview_failure_reason"`
	NewOrderFailure string `json:"new_order_failure_reason"`
	EditFailure     string `json:"edit_failure_reason"`
}

// NormalizeError converts a Coinbase API error response to a structured error.
// It parses the error response and classifies it based on HTTP status code and error message.
//
// Error Classification:
//   - 401/403: Authentication/Authorization errors (Permanent)
//   - 429: Rate limit errors (RateLimit)
//   - 400: Invalid request errors (Permanent if client error, otherwise Temporary)
//   - 500/503: Server errors (Temporary)
//   - Other: Temporary by default
//
// Returns an error with appropriate classification and original error details.
func NormalizeError(statusCode int, body []byte) error {
	if len(body) == 0 {
		return fmt.Errorf("coinbase api error: status %d (no body)", statusCode)
	}

	var cbErr CoinbaseError
	if err := json.Unmarshal(body, &cbErr); err != nil {
		// If we can't parse the error, return raw body
		return fmt.Errorf("coinbase api error: status %d: %s", statusCode, string(body))
	}

	// Construct error message with all available details
	msg := formatErrorMessage(cbErr)

	// Classify error based on status code
	return classifyError(statusCode, msg, &cbErr)
}

// formatErrorMessage constructs a comprehensive error message from Coinbase error fields.
func formatErrorMessage(cbErr CoinbaseError) string {
	msg := "coinbase api error"

	if cbErr.Error != "" {
		msg = fmt.Sprintf("%s: %s", msg, cbErr.Error)
	}

	if cbErr.Message != "" {
		if cbErr.Error == "" {
			msg = fmt.Sprintf("%s: %s", msg, cbErr.Message)
		} else {
			msg = fmt.Sprintf("%s (%s)", msg, cbErr.Message)
		}
	}

	// Add additional details if available
	details := []string{}
	if cbErr.ErrorDetails != "" {
		details = append(details, cbErr.ErrorDetails)
	}
	if cbErr.PreviewFailure != "" {
		details = append(details, fmt.Sprintf("preview: %s", cbErr.PreviewFailure))
	}
	if cbErr.NewOrderFailure != "" {
		details = append(details, fmt.Sprintf("order: %s", cbErr.NewOrderFailure))
	}
	if cbErr.EditFailure != "" {
		details = append(details, fmt.Sprintf("edit: %s", cbErr.EditFailure))
	}

	if len(details) > 0 {
		detailStr := details[0]
		for i := 1; i < len(details); i++ {
			detailStr = detailStr + "; " + details[i]
		}
		msg = fmt.Sprintf("%s [%s]", msg, detailStr)
	}

	return msg
}

// classifyError determines the error type based on HTTP status code and error content.
func classifyError(statusCode int, msg string, cbErr *CoinbaseError) error {
	baseErr := fmt.Errorf("%s (status: %d)", msg, statusCode)

	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		// Authentication/authorization failures are permanent
		return &PermanentError{Err: baseErr, Code: "AUTH_FAILURE"}

	case http.StatusTooManyRequests:
		// Rate limit errors
		return &RateLimitError{Err: baseErr, Code: "RATE_LIMIT"}

	case http.StatusBadRequest:
		// Bad request - check if it's a client error (permanent) or server issue (temporary)
		if isClientError(cbErr) {
			return &PermanentError{Err: baseErr, Code: "INVALID_REQUEST"}
		}
		return &TemporaryError{Err: baseErr, Code: "BAD_REQUEST"}

	case http.StatusNotFound:
		// Resource not found - permanent
		return &PermanentError{Err: baseErr, Code: "NOT_FOUND"}

	case http.StatusInternalServerError, http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
		// Server errors are temporary
		return &TemporaryError{Err: baseErr, Code: "SERVER_ERROR"}

	default:
		// Unknown errors default to temporary
		return &TemporaryError{Err: baseErr, Code: "UNKNOWN"}
	}
}

// isClientError determines if an error is caused by client input (permanent) vs server issues (temporary).
func isClientError(cbErr *CoinbaseError) bool {
	// Check for error messages that indicate client-side issues
	clientErrors := []string{
		"invalid",
		"missing",
		"insufficient",
		"exceed",
		"too small",
		"too large",
		"not allowed",
		"unsupported",
		"duplicate",
		"malformed",
	}

	msg := cbErr.Error + " " + cbErr.Message + " " + cbErr.ErrorDetails
	for _, pattern := range clientErrors {
		if contains(msg, pattern) {
			return true
		}
	}

	return false
}

// contains checks if s contains substr (case-insensitive).
func contains(s, substr string) bool {
	// Simple case-insensitive check
	s = toLower(s)
	substr = toLower(substr)
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toLower converts a string to lowercase.
func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

// Error types for classification

// PermanentError represents an error that won't succeed on retry (e.g., invalid request, auth failure).
type PermanentError struct {
	Err  error
	Code string
}

func (e *PermanentError) Error() string {
	return fmt.Sprintf("permanent error [%s]: %v", e.Code, e.Err)
}

func (e *PermanentError) Unwrap() error {
	return e.Err
}

// TemporaryError represents an error that might succeed on retry (e.g., server error, timeout).
type TemporaryError struct {
	Err  error
	Code string
}

func (e *TemporaryError) Error() string {
	return fmt.Sprintf("temporary error [%s]: %v", e.Code, e.Err)
}

func (e *TemporaryError) Unwrap() error {
	return e.Err
}

func (e *TemporaryError) Temporary() bool {
	return true
}

// RateLimitError represents a rate limit error.
type RateLimitError struct {
	Err  error
	Code string
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit error [%s]: %v", e.Code, e.Err)
}

func (e *RateLimitError) Unwrap() error {
	return e.Err
}

func (e *RateLimitError) RateLimit() bool {
	return true
}
