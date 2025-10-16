package prime

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNormalizeOrder tests order normalization with various order types.
func TestNormalizeOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("limit order", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("testdata", "order_limit.json"))
		require.NoError(t, err)

		order, err := NormalizeOrder(ctx, data)
		require.NoError(t, err)
		assert.NotNil(t, order)

		// Verify order fields
		assert.Equal(t, "test-order-123", *order.OrderId)
		assert.Equal(t, "client-abc", *order.ClientOrderId)
		assert.Equal(t, "BTC-USD", *order.VenueSymbol)
		assert.Equal(t, "ORDER_SIDE_BUY", order.Side.String())
		assert.Equal(t, "ORDER_TYPE_LIMIT", order.OrderType.String())
		assert.Equal(t, "ORDER_STATUS_OPEN", order.Status.String())
		assert.Equal(t, "TIME_IN_FORCE_GTC", order.TimeInForce.String())

		// Verify quantities and prices
		assert.Equal(t, 1.5, *order.Quantity)
		assert.Equal(t, 50000.00, *order.Price)
		assert.Equal(t, 0.5, *order.FilledQuantity)
		assert.Equal(t, 49950.00, *order.AverageFillPrice)
		assert.Equal(t, 25.00, *order.TotalFees)
	})

	t.Run("TWAP order", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("testdata", "order_twap.json"))
		require.NoError(t, err)

		order, err := NormalizeOrder(ctx, data)
		require.NoError(t, err)
		assert.NotNil(t, order)

		// Verify TWAP-specific fields
		assert.Equal(t, "twap-order-456", *order.OrderId)
		assert.Equal(t, "ETH-USD", *order.VenueSymbol)
		assert.Equal(t, "ORDER_SIDE_SELL", order.Side.String())

		// TWAP orders are mapped to LIMIT type in CQC
		// The actual type "TWAP" is preserved in the original order data
		assert.Equal(t, "ORDER_TYPE_LIMIT", order.OrderType.String())
		assert.Equal(t, "ORDER_STATUS_OPEN", order.Status.String()) // WORKING maps to OPEN
		assert.Equal(t, "TIME_IN_FORCE_GTD", order.TimeInForce.String())

		// Verify quantities
		assert.Equal(t, 10.0, *order.Quantity)
		assert.Equal(t, 3.2, *order.FilledQuantity)
		assert.Equal(t, 3010.50, *order.AverageFillPrice)
	})

	t.Run("empty response", func(t *testing.T) {
		_, err := NormalizeOrder(ctx, []byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty order response")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := NormalizeOrder(ctx, []byte("invalid json"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse prime order")
	})
}

// TestNormalizeExecutionReport tests execution report (fill) normalization.
func TestNormalizeExecutionReport(t *testing.T) {
	ctx := context.Background()

	t.Run("valid fill", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("testdata", "fill.json"))
		require.NoError(t, err)

		report, err := NormalizeExecutionReport(ctx, data)
		require.NoError(t, err)
		assert.NotNil(t, report)

		// Verify execution report fields
		assert.Equal(t, "fill-789", *report.ExecutionId)
		assert.Equal(t, "order-123", *report.OrderId)
		assert.Equal(t, "BTC-USD", *report.VenueSymbol)
		assert.Equal(t, "EXECUTION_TYPE_FILL", report.ExecutionType.String())
		assert.Equal(t, "PARTIALLY_FILLED", *report.OrderStatus)
		assert.Equal(t, "BUY", *report.Side)

		// Verify prices and quantities
		assert.Equal(t, 49950.00, *report.Price)
		assert.Equal(t, 0.5, *report.Quantity)
		assert.Equal(t, 25.00, *report.Fee)
		assert.Equal(t, "match-abc", *report.TradeId)

		// Verify calculated value
		expectedValue := 49950.00 * 0.5
		assert.InDelta(t, expectedValue, *report.Value, 0.01)

		// Verify client order ID
		assert.Equal(t, "client-abc", *report.ClientOrderId)
	})

	t.Run("empty response", func(t *testing.T) {
		_, err := NormalizeExecutionReport(ctx, []byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty execution report response")
	})
}

// TestNormalizeBalance tests balance normalization.
func TestNormalizeBalance(t *testing.T) {
	ctx := context.Background()

	t.Run("valid balance", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("testdata", "balance.json"))
		require.NoError(t, err)

		balance, err := NormalizeBalance(ctx, data)
		require.NoError(t, err)
		assert.NotNil(t, balance)

		// Verify balance fields
		assert.Equal(t, "BTC", *balance.AssetId)
		assert.Equal(t, 10.5, *balance.Total)
		assert.Equal(t, 2.0, *balance.Locked)

		// Verify available = total - locked
		expectedAvailable := 10.5 - 2.0
		assert.InDelta(t, expectedAvailable, *balance.Available, 0.0001)
	})

	t.Run("empty response", func(t *testing.T) {
		_, err := NormalizeBalance(ctx, []byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty balance response")
	})
}

// TestNormalizeOrderBook tests order book normalization.
func TestNormalizeOrderBook(t *testing.T) {
	ctx := context.Background()

	t.Run("valid orderbook", func(t *testing.T) {
		data, err := os.ReadFile(filepath.Join("testdata", "orderbook.json"))
		require.NoError(t, err)

		book, err := NormalizeOrderBook(ctx, data)
		require.NoError(t, err)
		assert.NotNil(t, book)

		// Verify orderbook fields
		assert.Equal(t, "prime", *book.VenueId)
		assert.Equal(t, "BTC-USD", *book.VenueSymbol)
		assert.NotNil(t, book.Timestamp)
		assert.Equal(t, int64(12345), *book.Sequence)

		// Verify bids
		require.Len(t, book.Bids, 3)
		assert.Equal(t, 50000.00, *book.Bids[0].Price)
		assert.Equal(t, 1.5, *book.Bids[0].Quantity)

		// Verify asks
		require.Len(t, book.Asks, 3)
		assert.Equal(t, 50010.00, *book.Asks[0].Price)
		assert.Equal(t, 1.2, *book.Asks[0].Quantity)

		// Verify best bid/ask
		assert.Equal(t, 50000.00, *book.BestBid)
		assert.Equal(t, 50010.00, *book.BestAsk)

		// Verify spread and mid price
		expectedSpread := 50010.00 - 50000.00
		assert.InDelta(t, expectedSpread, *book.Spread, 0.01)

		expectedMid := (50000.00 + 50010.00) / 2.0
		assert.InDelta(t, expectedMid, *book.MidPrice, 0.01)
	})

	t.Run("empty response", func(t *testing.T) {
		_, err := NormalizeOrderBook(ctx, []byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty orderbook response")
	})
}

// TestNormalizeError tests error normalization and classification.
func TestNormalizeError(t *testing.T) {
	t.Run("unauthorized error", func(t *testing.T) {
		body := []byte(`{"message": "Invalid credentials", "code": "UNAUTHORIZED"}`)
		err := NormalizeError(401, body)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "prime api error")
		assert.Contains(t, err.Error(), "UNAUTHORIZED")

		// Should be a permanent error
		permErr, ok := err.(*PermanentError)
		assert.True(t, ok, "expected PermanentError")
		assert.NotNil(t, permErr)
	})

	t.Run("rate limit error", func(t *testing.T) {
		body := []byte(`{"message": "Too many requests", "code": "RATE_LIMIT_EXCEEDED"}`)
		err := NormalizeError(429, body)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "rate limit error")

		// Should be a rate limit error
		rateLimitErr, ok := err.(*RateLimitError)
		assert.True(t, ok, "expected RateLimitError")
		assert.NotNil(t, rateLimitErr)
		assert.True(t, rateLimitErr.Temporary())
	})

	t.Run("validation error", func(t *testing.T) {
		body := []byte(`{"message": "Invalid order", "code": "INVALID_ORDER"}`)
		err := NormalizeError(400, body)

		assert.Error(t, err)

		// Should be a permanent error (client error)
		permErr, ok := err.(*PermanentError)
		assert.True(t, ok, "expected PermanentError for validation error")
		assert.NotNil(t, permErr)
	})

	t.Run("server error", func(t *testing.T) {
		body := []byte(`{"message": "Internal server error", "code": "INTERNAL_ERROR"}`)
		err := NormalizeError(500, body)

		assert.Error(t, err)

		// Should be a temporary error
		tempErr, ok := err.(*TemporaryError)
		assert.True(t, ok, "expected TemporaryError")
		assert.NotNil(t, tempErr)
		assert.True(t, tempErr.Temporary())
	})

	t.Run("empty body", func(t *testing.T) {
		err := NormalizeError(500, []byte{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no body")
	})
}

// TestOrderTypeMapping tests the mapping of Prime order types to CQC types.
func TestOrderTypeMapping(t *testing.T) {
	tests := []struct {
		primeType string
		expected  string
	}{
		{"MARKET", "ORDER_TYPE_MARKET"},
		{"LIMIT", "ORDER_TYPE_LIMIT"},
		{"STOP_LIMIT", "ORDER_TYPE_STOP_LIMIT"},
		{"TWAP", "ORDER_TYPE_LIMIT"},  // Mapped to LIMIT
		{"VWAP", "ORDER_TYPE_LIMIT"},  // Mapped to LIMIT
		{"BLOCK", "ORDER_TYPE_LIMIT"}, // Mapped to LIMIT
		{"RFQ", "ORDER_TYPE_LIMIT"},   // Mapped to LIMIT
		{"UNKNOWN", "ORDER_TYPE_UNSPECIFIED"},
	}

	for _, tt := range tests {
		t.Run(tt.primeType, func(t *testing.T) {
			result := mapOrderType(tt.primeType)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestOrderStatusMapping tests the mapping of Prime order statuses to CQC statuses.
func TestOrderStatusMapping(t *testing.T) {
	tests := []struct {
		primeStatus string
		expected    string
	}{
		{"OPEN", "ORDER_STATUS_OPEN"},
		{"WORKING", "ORDER_STATUS_OPEN"},
		{"FILLED", "ORDER_STATUS_FILLED"},
		{"CANCELLED", "ORDER_STATUS_CANCELLED"},
		{"EXPIRED", "ORDER_STATUS_REJECTED"},
		{"PENDING", "ORDER_STATUS_PENDING"},
		{"REJECTED", "ORDER_STATUS_REJECTED"},
		{"UNKNOWN", "ORDER_STATUS_UNSPECIFIED"},
	}

	for _, tt := range tests {
		t.Run(tt.primeStatus, func(t *testing.T) {
			result := mapOrderStatus(tt.primeStatus)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}

// TestTimeInForceMapping tests the mapping of Prime TIF to CQC TIF.
func TestTimeInForceMapping(t *testing.T) {
	tests := []struct {
		primeTIF string
		expected string
	}{
		{"GOOD_UNTIL_DATE_TIME", "TIME_IN_FORCE_GTD"},
		{"GOOD_UNTIL_CANCELLED", "TIME_IN_FORCE_GTC"},
		{"IMMEDIATE_OR_CANCEL", "TIME_IN_FORCE_IOC"},
		{"FILL_OR_KILL", "TIME_IN_FORCE_FOK"},
		{"UNKNOWN", "TIME_IN_FORCE_GTC"}, // Default to GTC
	}

	for _, tt := range tests {
		t.Run(tt.primeTIF, func(t *testing.T) {
			result := mapTimeInForce(tt.primeTIF)
			assert.Equal(t, tt.expected, result.String())
		})
	}
}
