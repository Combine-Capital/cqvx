package normalizer

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Timestamp Conversion Utilities

// ParseTimestamp attempts to parse a timestamp string in various common formats
// and returns a protobuf Timestamp. Handles RFC3339, ISO8601, and Unix timestamps.
//
// Supported formats:
//   - RFC3339: "2006-01-02T15:04:05Z07:00"
//   - ISO8601: "2006-01-02T15:04:05.999Z"
//   - Unix seconds: "1609459200"
//   - Unix milliseconds: "1609459200000"
//   - Unix microseconds: "1609459200000000"
//
// Returns nil for empty strings.
// Returns an error for malformed or unparseable timestamps.
func ParseTimestamp(s string) (*timestamppb.Timestamp, error) {
	// Handle empty/null cases
	s = strings.TrimSpace(s)
	if s == "" || s == "null" {
		return nil, nil
	}

	// Try parsing as numeric Unix timestamp (seconds, millis, or micros)
	if isNumeric(s) {
		return parseUnixTimestamp(s)
	}

	// Try common timestamp formats
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05.999Z",      // ISO8601 with millis
		"2006-01-02T15:04:05.999999Z",   // ISO8601 with micros
		"2006-01-02T15:04:05Z",          // ISO8601 without timezone
		"2006-01-02 15:04:05",           // Common SQL format
		"2006-01-02T15:04:05.999Z07:00", // ISO8601 with timezone
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return timestamppb.New(t), nil
		}
	}

	return nil, fmt.Errorf("unable to parse timestamp: %q", s)
}

// ParseTimestampOrNow parses a timestamp string, returning the current time if parsing fails.
// This is useful for non-critical timestamp fields where a default is acceptable.
func ParseTimestampOrNow(s string) *timestamppb.Timestamp {
	ts, err := ParseTimestamp(s)
	if err != nil || ts == nil {
		return timestamppb.Now()
	}
	return ts
}

// parseUnixTimestamp parses a Unix timestamp string that could be in seconds,
// milliseconds, or microseconds.
func parseUnixTimestamp(s string) (*timestamppb.Timestamp, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid unix timestamp: %w", err)
	}

	// Determine the unit based on magnitude
	// Unix seconds: ~10 digits (e.g., 1609459200)
	// Unix millis:  ~13 digits (e.g., 1609459200000)
	// Unix micros:  ~16 digits (e.g., 1609459200000000)
	var t time.Time
	switch {
	case n < 1e11: // Seconds (less than ~3000 years from epoch)
		t = time.Unix(n, 0)
	case n < 1e14: // Milliseconds
		t = time.Unix(n/1000, (n%1000)*1e6)
	case n < 1e17: // Microseconds
		t = time.Unix(n/1e6, (n%1e6)*1e3)
	default:
		return nil, fmt.Errorf("unix timestamp out of reasonable range: %d", n)
	}

	return timestamppb.New(t), nil
}

// isNumeric returns true if the string contains only digits (and optionally a leading minus).
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '-' {
		s = s[1:]
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// Decimal Conversion Utilities

// ParseDecimal converts a string or number to a float64 for use in CQC protobuf types.
// Handles various decimal formats including scientific notation.
//
// Examples:
//   - "123.45" -> 123.45
//   - "1.23e5" -> 123000.0
//   - "0.00000001" -> 0.00000001
//   - "" -> 0.0
//   - "null" -> 0.0
//
// Returns an error for malformed decimal strings.
func ParseDecimal(s string) (float64, error) {
	// Handle empty/null cases
	s = strings.TrimSpace(s)
	if s == "" || s == "null" {
		return 0.0, nil
	}

	// Parse as float64
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, fmt.Errorf("invalid decimal: %w", err)
	}

	// Check for special values
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return 0.0, fmt.Errorf("invalid decimal: %s (NaN or Inf)", s)
	}

	return f, nil
}

// ParseDecimalOrZero parses a decimal string, returning 0.0 if parsing fails.
// This is useful for optional numeric fields where a default is acceptable.
func ParseDecimalOrZero(s string) float64 {
	f, err := ParseDecimal(s)
	if err != nil {
		return 0.0
	}
	return f
}

// MustParseDecimal parses a decimal string and panics if parsing fails.
// This should only be used in tests or when the input is guaranteed to be valid.
func MustParseDecimal(s string) float64 {
	f, err := ParseDecimal(s)
	if err != nil {
		panic(fmt.Sprintf("MustParseDecimal: %v", err))
	}
	return f
}

// FormatDecimal converts a float64 to a string with appropriate precision.
// Removes trailing zeros and unnecessary decimal points.
func FormatDecimal(f float64) string {
	// Use strconv for proper formatting
	s := strconv.FormatFloat(f, 'f', -1, 64)
	return s
}

// Enum Mapping Utilities

// ParseOrderStatus converts a venue-specific order status string to a CQC OrderStatus enum.
//
// Common mappings (case-insensitive):
//   - "open", "new", "active", "pending" -> ORDER_STATUS_OPEN
//   - "filled", "done", "closed" -> ORDER_STATUS_FILLED
//   - "cancelled", "canceled" -> ORDER_STATUS_CANCELLED
//   - "rejected", "failed" -> ORDER_STATUS_REJECTED
//   - "partially_filled", "partial", "partial-fill" -> ORDER_STATUS_PARTIALLY_FILLED
//
// Returns ORDER_STATUS_UNSPECIFIED for unrecognized statuses.
func ParseOrderStatus(s string) venuesv1.OrderStatus {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	switch s {
	case "open", "new", "active", "pending", "accepted":
		return venuesv1.OrderStatus_ORDER_STATUS_OPEN
	case "filled", "done", "closed", "complete":
		return venuesv1.OrderStatus_ORDER_STATUS_FILLED
	case "cancelled", "canceled", "cancelled_by_user", "canceled_by_user":
		return venuesv1.OrderStatus_ORDER_STATUS_CANCELLED
	case "rejected", "failed", "invalid", "expired":
		return venuesv1.OrderStatus_ORDER_STATUS_REJECTED
	case "partially_filled", "partial", "partial_fill", "partially_filled_active":
		return venuesv1.OrderStatus_ORDER_STATUS_PARTIALLY_FILLED
	default:
		return venuesv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

// ParseOrderType converts a venue-specific order type string to a CQC OrderType enum.
//
// Common mappings (case-insensitive):
//   - "limit" -> ORDER_TYPE_LIMIT
//   - "market" -> ORDER_TYPE_MARKET
//   - "stop", "stop_loss" -> ORDER_TYPE_STOP
//   - "stop_limit" -> ORDER_TYPE_STOP_LIMIT
//
// Returns ORDER_TYPE_UNSPECIFIED for unrecognized types.
func ParseOrderType(s string) venuesv1.OrderType {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	switch s {
	case "limit":
		return venuesv1.OrderType_ORDER_TYPE_LIMIT
	case "market":
		return venuesv1.OrderType_ORDER_TYPE_MARKET
	case "stop", "stop_loss", "stop_market":
		return venuesv1.OrderType_ORDER_TYPE_STOP_LOSS
	case "stop_limit", "stop_loss_limit":
		return venuesv1.OrderType_ORDER_TYPE_STOP_LIMIT
	case "trailing_stop", "trailing_stop_loss":
		return venuesv1.OrderType_ORDER_TYPE_TRAILING_STOP
	case "post_only", "maker_only":
		return venuesv1.OrderType_ORDER_TYPE_POST_ONLY
	case "ioc", "immediate_or_cancel":
		return venuesv1.OrderType_ORDER_TYPE_IOC
	case "fok", "fill_or_kill":
		return venuesv1.OrderType_ORDER_TYPE_FOK
	case "gtc", "good_til_cancelled":
		return venuesv1.OrderType_ORDER_TYPE_GTC
	default:
		return venuesv1.OrderType_ORDER_TYPE_UNSPECIFIED
	}
}

// ParseOrderSide converts a venue-specific order side string to a CQC OrderSide enum.
//
// Common mappings (case-insensitive):
//   - "buy", "bid" -> ORDER_SIDE_BUY
//   - "sell", "ask" -> ORDER_SIDE_SELL
//
// Returns ORDER_SIDE_UNSPECIFIED for unrecognized sides.
func ParseOrderSide(s string) venuesv1.OrderSide {
	s = strings.ToLower(strings.TrimSpace(s))

	switch s {
	case "buy", "bid":
		return venuesv1.OrderSide_ORDER_SIDE_BUY
	case "sell", "ask":
		return venuesv1.OrderSide_ORDER_SIDE_SELL
	default:
		return venuesv1.OrderSide_ORDER_SIDE_UNSPECIFIED
	}
}

// ParseTimeInForce converts a venue-specific time-in-force string to a CQC TimeInForce enum.
//
// Common mappings (case-insensitive):
//   - "GTC", "good_til_cancelled" -> TIME_IN_FORCE_GTC
//   - "IOC", "immediate_or_cancel" -> TIME_IN_FORCE_IOC
//   - "FOK", "fill_or_kill" -> TIME_IN_FORCE_FOK
//   - "GTD", "good_til_date" -> TIME_IN_FORCE_GTD
//
// Returns TIME_IN_FORCE_UNSPECIFIED for unrecognized values.
func ParseTimeInForce(s string) venuesv1.TimeInForce {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")

	switch s {
	case "gtc", "good_til_cancelled", "good_til_canceled", "good_till_cancelled":
		return venuesv1.TimeInForce_TIME_IN_FORCE_GTC
	case "ioc", "immediate_or_cancel":
		return venuesv1.TimeInForce_TIME_IN_FORCE_IOC
	case "fok", "fill_or_kill":
		return venuesv1.TimeInForce_TIME_IN_FORCE_FOK
	case "gtd", "good_til_date", "good_til_time":
		return venuesv1.TimeInForce_TIME_IN_FORCE_GTD
	default:
		return venuesv1.TimeInForce_TIME_IN_FORCE_UNSPECIFIED
	}
}

// String Utilities

// StringPtr returns a pointer to the string value.
// This is useful for populating optional protobuf fields.
func StringPtr(s string) *string {
	return &s
}

// Float64Ptr returns a pointer to the float64 value.
// This is useful for populating optional protobuf fields.
func Float64Ptr(f float64) *float64 {
	return &f
}

// Int64Ptr returns a pointer to the int64 value.
// This is useful for populating optional protobuf fields.
func Int64Ptr(i int64) *int64 {
	return &i
}

// SafeString returns the string value or empty string if the pointer is nil.
func SafeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// SafeFloat64 returns the float64 value or 0.0 if the pointer is nil.
func SafeFloat64(f *float64) float64 {
	if f == nil {
		return 0.0
	}
	return *f
}

// SafeInt64 returns the int64 value or 0 if the pointer is nil.
func SafeInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
