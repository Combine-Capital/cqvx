package coinbase

import (
	"net/http"
	"testing"
)

func TestNormalizeError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantType   string // "permanent", "temporary", "ratelimit"
		wantCode   string
	}{
		{
			name:       "authentication failure",
			statusCode: http.StatusUnauthorized,
			body:       []byte(`{"error": "unauthorized", "message": "Invalid API key"}`),
			wantType:   "permanent",
			wantCode:   "AUTH_FAILURE",
		},
		{
			name:       "authorization failure",
			statusCode: http.StatusForbidden,
			body:       []byte(`{"error": "forbidden", "message": "Insufficient permissions"}`),
			wantType:   "permanent",
			wantCode:   "AUTH_FAILURE",
		},
		{
			name:       "rate limit error",
			statusCode: http.StatusTooManyRequests,
			body:       []byte(`{"error": "rate_limit", "message": "Too many requests"}`),
			wantType:   "ratelimit",
			wantCode:   "RATE_LIMIT",
		},
		{
			name:       "invalid request - client error",
			statusCode: http.StatusBadRequest,
			body:       []byte(`{"error": "invalid_request", "message": "Invalid order size"}`),
			wantType:   "permanent",
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "insufficient funds - client error",
			statusCode: http.StatusBadRequest,
			body:       []byte(`{"error": "insufficient_funds", "message": "Not enough balance"}`),
			wantType:   "permanent",
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "bad request - server issue",
			statusCode: http.StatusBadRequest,
			body:       []byte(`{"error": "processing_error", "message": "Request failed"}`),
			wantType:   "temporary",
			wantCode:   "BAD_REQUEST",
		},
		{
			name:       "not found error",
			statusCode: http.StatusNotFound,
			body:       []byte(`{"error": "not_found", "message": "Order not found"}`),
			wantType:   "permanent",
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			body:       []byte(`{"error": "internal_error", "message": "Server error"}`),
			wantType:   "temporary",
			wantCode:   "SERVER_ERROR",
		},
		{
			name:       "service unavailable",
			statusCode: http.StatusServiceUnavailable,
			body:       []byte(`{"error": "service_unavailable", "message": "Service down"}`),
			wantType:   "temporary",
			wantCode:   "SERVER_ERROR",
		},
		{
			name:       "empty body",
			statusCode: http.StatusInternalServerError,
			body:       []byte{},
			wantType:   "error", // Just check it returns an error
			wantCode:   "",
		},
		{
			name:       "malformed JSON",
			statusCode: http.StatusBadRequest,
			body:       []byte(`{invalid json`),
			wantType:   "error",
			wantCode:   "",
		},
		{
			name:       "error with details",
			statusCode: http.StatusBadRequest,
			body:       []byte(`{"error": "invalid_order", "message": "Order failed", "error_details": "Size too small", "preview_failure_reason": "Insufficient funds"}`),
			wantType:   "permanent",
			wantCode:   "INVALID_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NormalizeError(tt.statusCode, tt.body)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			// Check error type
			switch tt.wantType {
			case "permanent":
				if _, ok := err.(*PermanentError); !ok {
					t.Errorf("expected PermanentError, got %T: %v", err, err)
				}
				if perr, ok := err.(*PermanentError); ok && perr.Code != tt.wantCode {
					t.Errorf("expected code %q, got %q", tt.wantCode, perr.Code)
				}
			case "temporary":
				if _, ok := err.(*TemporaryError); !ok {
					t.Errorf("expected TemporaryError, got %T: %v", err, err)
				}
				if terr, ok := err.(*TemporaryError); ok && terr.Code != tt.wantCode {
					t.Errorf("expected code %q, got %q", tt.wantCode, terr.Code)
				}
			case "ratelimit":
				if _, ok := err.(*RateLimitError); !ok {
					t.Errorf("expected RateLimitError, got %T: %v", err, err)
				}
				if rerr, ok := err.(*RateLimitError); ok && rerr.Code != tt.wantCode {
					t.Errorf("expected code %q, got %q", tt.wantCode, rerr.Code)
				}
			case "error":
				// Just verify it's an error
			}

			// Error message should contain status code
			if tt.statusCode > 0 && !contains(err.Error(), "status") {
				t.Errorf("error message should contain status code: %v", err)
			}
		})
	}
}

func TestIsClientError(t *testing.T) {
	tests := []struct {
		name   string
		cbErr  *CoinbaseError
		expect bool
	}{
		{
			name:   "invalid request",
			cbErr:  &CoinbaseError{Error: "invalid_size"},
			expect: true,
		},
		{
			name:   "missing field",
			cbErr:  &CoinbaseError{Message: "missing required field"},
			expect: true,
		},
		{
			name:   "insufficient funds",
			cbErr:  &CoinbaseError{ErrorDetails: "insufficient balance"},
			expect: true,
		},
		{
			name:   "size too small",
			cbErr:  &CoinbaseError{Message: "Order size too small"},
			expect: true,
		},
		{
			name:   "exceeds limit",
			cbErr:  &CoinbaseError{Error: "exceeds maximum"},
			expect: true,
		},
		{
			name:   "not allowed",
			cbErr:  &CoinbaseError{Message: "operation not allowed"},
			expect: true,
		},
		{
			name:   "unsupported",
			cbErr:  &CoinbaseError{Error: "unsupported order type"},
			expect: true,
		},
		{
			name:   "duplicate",
			cbErr:  &CoinbaseError{Message: "duplicate client_order_id"},
			expect: true,
		},
		{
			name:   "malformed",
			cbErr:  &CoinbaseError{ErrorDetails: "malformed request"},
			expect: true,
		},
		{
			name:   "server error",
			cbErr:  &CoinbaseError{Error: "internal_server_error"},
			expect: false,
		},
		{
			name:   "processing error",
			cbErr:  &CoinbaseError{Message: "processing failed"},
			expect: false,
		},
		{
			name:   "timeout",
			cbErr:  &CoinbaseError{Error: "request timeout"},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClientError(tt.cbErr)
			if got != tt.expect {
				t.Errorf("isClientError() = %v, want %v for error: %+v", got, tt.expect, tt.cbErr)
			}
		})
	}
}

func TestFormatErrorMessage(t *testing.T) {
	tests := []struct {
		name   string
		cbErr  CoinbaseError
		expect string
	}{
		{
			name: "error only",
			cbErr: CoinbaseError{
				Error: "invalid_request",
			},
			expect: "coinbase api error: invalid_request",
		},
		{
			name: "error and message",
			cbErr: CoinbaseError{
				Error:   "invalid_request",
				Message: "Order size too small",
			},
			expect: "coinbase api error: invalid_request (Order size too small)",
		},
		{
			name: "message only",
			cbErr: CoinbaseError{
				Message: "Order failed",
			},
			expect: "coinbase api error: Order failed",
		},
		{
			name: "with error details",
			cbErr: CoinbaseError{
				Error:        "invalid_request",
				Message:      "Validation failed",
				ErrorDetails: "Size must be greater than 0.001",
			},
			expect: "coinbase api error: invalid_request (Validation failed) [Size must be greater than 0.001]",
		},
		{
			name: "with multiple details",
			cbErr: CoinbaseError{
				Error:           "order_failed",
				Message:         "Order could not be placed",
				PreviewFailure:  "Insufficient funds",
				NewOrderFailure: "Market closed",
			},
			expect: "coinbase api error: order_failed (Order could not be placed) [preview: Insufficient funds; order: Market closed]",
		},
		{
			name:   "empty error",
			cbErr:  CoinbaseError{},
			expect: "coinbase api error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatErrorMessage(tt.cbErr)
			if got != tt.expect {
				t.Errorf("formatErrorMessage() = %q, want %q", got, tt.expect)
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	t.Run("PermanentError", func(t *testing.T) {
		baseErr := &PermanentError{
			Err:  http.ErrBodyNotAllowed,
			Code: "INVALID_REQUEST",
		}

		if !contains(baseErr.Error(), "permanent") {
			t.Errorf("PermanentError.Error() should contain 'permanent': %v", baseErr.Error())
		}

		if !contains(baseErr.Error(), "INVALID_REQUEST") {
			t.Errorf("PermanentError.Error() should contain code: %v", baseErr.Error())
		}

		if baseErr.Unwrap() != http.ErrBodyNotAllowed {
			t.Errorf("PermanentError.Unwrap() failed")
		}
	})

	t.Run("TemporaryError", func(t *testing.T) {
		baseErr := &TemporaryError{
			Err:  http.ErrServerClosed,
			Code: "SERVER_ERROR",
		}

		if !contains(baseErr.Error(), "temporary") {
			t.Errorf("TemporaryError.Error() should contain 'temporary': %v", baseErr.Error())
		}

		if !baseErr.Temporary() {
			t.Error("TemporaryError.Temporary() should return true")
		}

		if baseErr.Unwrap() != http.ErrServerClosed {
			t.Errorf("TemporaryError.Unwrap() failed")
		}
	})

	t.Run("RateLimitError", func(t *testing.T) {
		baseErr := &RateLimitError{
			Err:  http.ErrHandlerTimeout,
			Code: "RATE_LIMIT",
		}

		if !contains(baseErr.Error(), "rate limit") {
			t.Errorf("RateLimitError.Error() should contain 'rate limit': %v", baseErr.Error())
		}

		if !baseErr.RateLimit() {
			t.Error("RateLimitError.RateLimit() should return true")
		}

		if baseErr.Unwrap() != http.ErrHandlerTimeout {
			t.Errorf("RateLimitError.Unwrap() failed")
		}
	})
}

func TestContains(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"hello world", "world", true},
		{"hello world", "WORLD", true}, // case-insensitive
		{"hello world", "foo", false},
		{"HELLO WORLD", "hello", true}, // case-insensitive
		{"", "anything", false},
		{"something", "", true}, // empty substring always matches
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			if got != tt.want {
				t.Errorf("contains(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Hello", "hello"},
		{"WORLD", "world"},
		{"MixedCase", "mixedcase"},
		{"123ABC", "123abc"},
		{"already-lower", "already-lower"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := toLower(tt.input)
			if got != tt.want {
				t.Errorf("toLower(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
