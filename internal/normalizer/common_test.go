package normalizer

import (
	"testing"
	"time"

	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseTimestamp tests timestamp parsing with various formats
func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		wantNil   bool
		checkTime func(t *testing.T, ts int64) // Check Unix timestamp
	}{
		{
			name:    "empty string",
			input:   "",
			wantNil: true,
		},
		{
			name:    "null string",
			input:   "null",
			wantNil: true,
		},
		{
			name:    "whitespace only",
			input:   "   ",
			wantNil: true,
		},
		{
			name:  "RFC3339 format",
			input: "2021-01-01T00:00:00Z",
			checkTime: func(t *testing.T, ts int64) {
				expected := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
				assert.Equal(t, expected, ts)
			},
		},
		{
			name:  "RFC3339 with timezone",
			input: "2021-01-01T00:00:00-05:00",
			checkTime: func(t *testing.T, ts int64) {
				loc, _ := time.LoadLocation("America/New_York")
				expected := time.Date(2021, 1, 1, 0, 0, 0, 0, loc).Unix()
				assert.Equal(t, expected, ts)
			},
		},
		{
			name:  "RFC3339Nano format",
			input: "2021-01-01T00:00:00.123456789Z",
			checkTime: func(t *testing.T, ts int64) {
				expected := time.Date(2021, 1, 1, 0, 0, 0, 123456789, time.UTC).Unix()
				assert.Equal(t, expected, ts)
			},
		},
		{
			name:  "ISO8601 with millis",
			input: "2021-01-01T00:00:00.123Z",
			checkTime: func(t *testing.T, ts int64) {
				expected := time.Date(2021, 1, 1, 0, 0, 0, 123000000, time.UTC).Unix()
				assert.Equal(t, expected, ts)
			},
		},
		{
			name:  "Unix seconds",
			input: "1609459200",
			checkTime: func(t *testing.T, ts int64) {
				assert.Equal(t, int64(1609459200), ts)
			},
		},
		{
			name:  "Unix milliseconds",
			input: "1609459200000",
			checkTime: func(t *testing.T, ts int64) {
				assert.Equal(t, int64(1609459200), ts)
			},
		},
		{
			name:  "Unix microseconds",
			input: "1609459200000000",
			checkTime: func(t *testing.T, ts int64) {
				assert.Equal(t, int64(1609459200), ts)
			},
		},
		{
			name:    "invalid format",
			input:   "not-a-timestamp",
			wantErr: true,
		},
		{
			name:    "invalid unix timestamp",
			input:   "abc123",
			wantErr: true,
		},
		{
			name:    "unix timestamp out of range",
			input:   "99999999999999999999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := ParseTimestamp(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.wantNil {
				assert.Nil(t, ts)
				return
			}

			require.NotNil(t, ts)
			if tt.checkTime != nil {
				tt.checkTime(t, ts.Seconds)
			}
		})
	}
}

// TestParseTimestampOrNow tests the fallback to current time
func TestParseTimestampOrNow(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectNow bool
	}{
		{
			name:      "invalid timestamp returns now",
			input:     "invalid",
			expectNow: true,
		},
		{
			name:      "empty string returns now",
			input:     "",
			expectNow: true,
		},
		{
			name:      "valid timestamp returns parsed time",
			input:     "2021-01-01T00:00:00Z",
			expectNow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now().Unix()
			ts := ParseTimestampOrNow(tt.input)
			after := time.Now().Unix()

			require.NotNil(t, ts)

			if tt.expectNow {
				// Should be within a few seconds of now
				assert.GreaterOrEqual(t, ts.Seconds, before-1)
				assert.LessOrEqual(t, ts.Seconds, after+1)
			} else {
				// Should be the parsed time (2021-01-01)
				assert.Less(t, ts.Seconds, before)
			}
		})
	}
}

// TestParseDecimal tests decimal parsing
func TestParseDecimal(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    float64
		wantErr bool
	}{
		{
			name:  "simple decimal",
			input: "123.45",
			want:  123.45,
		},
		{
			name:  "integer",
			input: "100",
			want:  100.0,
		},
		{
			name:  "scientific notation",
			input: "1.23e5",
			want:  123000.0,
		},
		{
			name:  "small decimal",
			input: "0.00000001",
			want:  0.00000001,
		},
		{
			name:  "negative decimal",
			input: "-123.45",
			want:  -123.45,
		},
		{
			name:  "zero",
			input: "0",
			want:  0.0,
		},
		{
			name:  "empty string",
			input: "",
			want:  0.0,
		},
		{
			name:  "null string",
			input: "null",
			want:  0.0,
		},
		{
			name:  "whitespace",
			input: "  123.45  ",
			want:  123.45,
		},
		{
			name:    "invalid decimal",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "NaN",
			input:   "NaN",
			wantErr: true,
		},
		{
			name:    "Infinity",
			input:   "Inf",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDecimal(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseDecimalOrZero tests fallback to zero
func TestParseDecimalOrZero(t *testing.T) {
	assert.Equal(t, 0.0, ParseDecimalOrZero("invalid"))
	assert.Equal(t, 0.0, ParseDecimalOrZero(""))
	assert.Equal(t, 123.45, ParseDecimalOrZero("123.45"))
}

// TestMustParseDecimal tests panic behavior
func TestMustParseDecimal(t *testing.T) {
	// Valid input should not panic
	assert.NotPanics(t, func() {
		result := MustParseDecimal("123.45")
		assert.Equal(t, 123.45, result)
	})

	// Invalid input should panic
	assert.Panics(t, func() {
		MustParseDecimal("invalid")
	})
}

// TestFormatDecimal tests decimal formatting
func TestFormatDecimal(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  string
	}{
		{
			name:  "simple decimal",
			input: 123.45,
			want:  "123.45",
		},
		{
			name:  "integer",
			input: 100.0,
			want:  "100",
		},
		{
			name:  "small decimal",
			input: 0.00000001,
			want:  "0.00000001",
		},
		{
			name:  "zero",
			input: 0.0,
			want:  "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDecimal(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseOrderStatus tests order status parsing
func TestParseOrderStatus(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  venuesv1.OrderStatus
	}{
		// Open variations
		{name: "open", input: "open", want: venuesv1.OrderStatus_ORDER_STATUS_OPEN},
		{name: "new", input: "new", want: venuesv1.OrderStatus_ORDER_STATUS_OPEN},
		{name: "active", input: "active", want: venuesv1.OrderStatus_ORDER_STATUS_OPEN},
		{name: "pending", input: "pending", want: venuesv1.OrderStatus_ORDER_STATUS_OPEN},
		{name: "accepted", input: "accepted", want: venuesv1.OrderStatus_ORDER_STATUS_OPEN},
		{name: "OPEN uppercase", input: "OPEN", want: venuesv1.OrderStatus_ORDER_STATUS_OPEN},

		// Filled variations
		{name: "filled", input: "filled", want: venuesv1.OrderStatus_ORDER_STATUS_FILLED},
		{name: "done", input: "done", want: venuesv1.OrderStatus_ORDER_STATUS_FILLED},
		{name: "closed", input: "closed", want: venuesv1.OrderStatus_ORDER_STATUS_FILLED},
		{name: "complete", input: "complete", want: venuesv1.OrderStatus_ORDER_STATUS_FILLED},

		// Cancelled variations
		{name: "cancelled", input: "cancelled", want: venuesv1.OrderStatus_ORDER_STATUS_CANCELLED},
		{name: "canceled", input: "canceled", want: venuesv1.OrderStatus_ORDER_STATUS_CANCELLED},
		{name: "cancelled_by_user", input: "cancelled_by_user", want: venuesv1.OrderStatus_ORDER_STATUS_CANCELLED},

		// Rejected variations
		{name: "rejected", input: "rejected", want: venuesv1.OrderStatus_ORDER_STATUS_REJECTED},
		{name: "failed", input: "failed", want: venuesv1.OrderStatus_ORDER_STATUS_REJECTED},
		{name: "invalid", input: "invalid", want: venuesv1.OrderStatus_ORDER_STATUS_REJECTED},
		{name: "expired", input: "expired", want: venuesv1.OrderStatus_ORDER_STATUS_REJECTED},

		// Partially filled variations
		{name: "partially_filled", input: "partially_filled", want: venuesv1.OrderStatus_ORDER_STATUS_PARTIALLY_FILLED},
		{name: "partial", input: "partial", want: venuesv1.OrderStatus_ORDER_STATUS_PARTIALLY_FILLED},
		{name: "partial-fill", input: "partial-fill", want: venuesv1.OrderStatus_ORDER_STATUS_PARTIALLY_FILLED},

		// Edge cases
		{name: "whitespace", input: "  open  ", want: venuesv1.OrderStatus_ORDER_STATUS_OPEN},
		{name: "mixed case", input: "OpEn", want: venuesv1.OrderStatus_ORDER_STATUS_OPEN},
		{name: "unknown", input: "unknown_status", want: venuesv1.OrderStatus_ORDER_STATUS_UNSPECIFIED},
		{name: "empty", input: "", want: venuesv1.OrderStatus_ORDER_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseOrderStatus(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseOrderType tests order type parsing
func TestParseOrderType(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  venuesv1.OrderType
	}{
		{name: "limit", input: "limit", want: venuesv1.OrderType_ORDER_TYPE_LIMIT},
		{name: "market", input: "market", want: venuesv1.OrderType_ORDER_TYPE_MARKET},
		{name: "stop", input: "stop", want: venuesv1.OrderType_ORDER_TYPE_STOP_LOSS},
		{name: "stop_loss", input: "stop_loss", want: venuesv1.OrderType_ORDER_TYPE_STOP_LOSS},
		{name: "stop-loss", input: "stop-loss", want: venuesv1.OrderType_ORDER_TYPE_STOP_LOSS},
		{name: "stop_limit", input: "stop_limit", want: venuesv1.OrderType_ORDER_TYPE_STOP_LIMIT},
		{name: "trailing_stop", input: "trailing_stop", want: venuesv1.OrderType_ORDER_TYPE_TRAILING_STOP},
		{name: "post_only", input: "post_only", want: venuesv1.OrderType_ORDER_TYPE_POST_ONLY},
		{name: "ioc", input: "ioc", want: venuesv1.OrderType_ORDER_TYPE_IOC},
		{name: "fok", input: "fok", want: venuesv1.OrderType_ORDER_TYPE_FOK},
		{name: "gtc", input: "gtc", want: venuesv1.OrderType_ORDER_TYPE_GTC},
		{name: "LIMIT uppercase", input: "LIMIT", want: venuesv1.OrderType_ORDER_TYPE_LIMIT},
		{name: "unknown", input: "unknown_type", want: venuesv1.OrderType_ORDER_TYPE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseOrderType(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseOrderSide tests order side parsing
func TestParseOrderSide(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  venuesv1.OrderSide
	}{
		{name: "buy", input: "buy", want: venuesv1.OrderSide_ORDER_SIDE_BUY},
		{name: "bid", input: "bid", want: venuesv1.OrderSide_ORDER_SIDE_BUY},
		{name: "sell", input: "sell", want: venuesv1.OrderSide_ORDER_SIDE_SELL},
		{name: "ask", input: "ask", want: venuesv1.OrderSide_ORDER_SIDE_SELL},
		{name: "BUY uppercase", input: "BUY", want: venuesv1.OrderSide_ORDER_SIDE_BUY},
		{name: "whitespace", input: "  buy  ", want: venuesv1.OrderSide_ORDER_SIDE_BUY},
		{name: "unknown", input: "unknown", want: venuesv1.OrderSide_ORDER_SIDE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseOrderSide(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestParseTimeInForce tests time-in-force parsing
func TestParseTimeInForce(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  venuesv1.TimeInForce
	}{
		{name: "GTC", input: "GTC", want: venuesv1.TimeInForce_TIME_IN_FORCE_GTC},
		{name: "gtc lowercase", input: "gtc", want: venuesv1.TimeInForce_TIME_IN_FORCE_GTC},
		{name: "good_til_cancelled", input: "good_til_cancelled", want: venuesv1.TimeInForce_TIME_IN_FORCE_GTC},
		{name: "IOC", input: "IOC", want: venuesv1.TimeInForce_TIME_IN_FORCE_IOC},
		{name: "immediate_or_cancel", input: "immediate_or_cancel", want: venuesv1.TimeInForce_TIME_IN_FORCE_IOC},
		{name: "FOK", input: "FOK", want: venuesv1.TimeInForce_TIME_IN_FORCE_FOK},
		{name: "fill_or_kill", input: "fill_or_kill", want: venuesv1.TimeInForce_TIME_IN_FORCE_FOK},
		{name: "GTD", input: "GTD", want: venuesv1.TimeInForce_TIME_IN_FORCE_GTD},
		{name: "good_til_date", input: "good_til_date", want: venuesv1.TimeInForce_TIME_IN_FORCE_GTD},
		{name: "unknown", input: "unknown", want: venuesv1.TimeInForce_TIME_IN_FORCE_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTimeInForce(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestStringPtr tests string pointer utility
func TestStringPtr(t *testing.T) {
	s := "test"
	ptr := StringPtr(s)
	assert.NotNil(t, ptr)
	assert.Equal(t, s, *ptr)
}

// TestFloat64Ptr tests float64 pointer utility
func TestFloat64Ptr(t *testing.T) {
	f := 123.45
	ptr := Float64Ptr(f)
	assert.NotNil(t, ptr)
	assert.Equal(t, f, *ptr)
}

// TestInt64Ptr tests int64 pointer utility
func TestInt64Ptr(t *testing.T) {
	i := int64(123)
	ptr := Int64Ptr(i)
	assert.NotNil(t, ptr)
	assert.Equal(t, i, *ptr)
}

// TestSafeString tests safe string utility
func TestSafeString(t *testing.T) {
	s := "test"
	assert.Equal(t, "test", SafeString(&s))
	assert.Equal(t, "", SafeString(nil))
}

// TestSafeFloat64 tests safe float64 utility
func TestSafeFloat64(t *testing.T) {
	f := 123.45
	assert.Equal(t, 123.45, SafeFloat64(&f))
	assert.Equal(t, 0.0, SafeFloat64(nil))
}

// TestSafeInt64 tests safe int64 utility
func TestSafeInt64(t *testing.T) {
	i := int64(123)
	assert.Equal(t, int64(123), SafeInt64(&i))
	assert.Equal(t, int64(0), SafeInt64(nil))
}

// BenchmarkParseTimestamp benchmarks timestamp parsing
func BenchmarkParseTimestamp(b *testing.B) {
	timestamps := []string{
		"2021-01-01T00:00:00Z",
		"1609459200",
		"1609459200000",
	}

	for i := 0; i < b.N; i++ {
		for _, ts := range timestamps {
			_, _ = ParseTimestamp(ts)
		}
	}
}

// BenchmarkParseDecimal benchmarks decimal parsing
func BenchmarkParseDecimal(b *testing.B) {
	decimals := []string{
		"123.45",
		"1.23e5",
		"0.00000001",
	}

	for i := 0; i < b.N; i++ {
		for _, dec := range decimals {
			_, _ = ParseDecimal(dec)
		}
	}
}

// BenchmarkParseOrderStatus benchmarks enum parsing
func BenchmarkParseOrderStatus(b *testing.B) {
	statuses := []string{
		"open",
		"filled",
		"cancelled",
		"rejected",
		"partially_filled",
	}

	for i := 0; i < b.N; i++ {
		for _, status := range statuses {
			_ = ParseOrderStatus(status)
		}
	}
}
