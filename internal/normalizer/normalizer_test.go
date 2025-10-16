package normalizer

import (
	"context"
	"testing"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/stretchr/testify/assert"
)

// MockNormalizer is a mock implementation of the Normalizer interface for testing.
// This demonstrates the contract that venue-specific normalizers must fulfill.
type MockNormalizer struct {
	// Allow configuration of mock responses
	OrderResponse           []byte
	ExecutionReportResponse []byte
	BalanceResponse         []byte
	OrderBookResponse       []byte
	TradeResponse           []byte
	ErrorResponse           error
}

// Ensure MockNormalizer implements Normalizer interface
var _ Normalizer = (*MockNormalizer)(nil)

// NormalizeOrder implements Normalizer.NormalizeOrder
func (m *MockNormalizer) NormalizeOrder(ctx context.Context, raw []byte) (*venuesv1.Order, error) {
	if m.ErrorResponse != nil {
		return nil, m.ErrorResponse
	}
	// Return a basic order for testing
	return &venuesv1.Order{
		OrderId:     StringPtr("mock-order-1"),
		VenueSymbol: StringPtr("BTC-USD"),
		OrderType:   venuesv1.OrderType_ORDER_TYPE_LIMIT.Enum(),
		Side:        venuesv1.OrderSide_ORDER_SIDE_BUY.Enum(),
		Status:      venuesv1.OrderStatus_ORDER_STATUS_OPEN.Enum(),
	}, nil
}

// NormalizeExecutionReport implements Normalizer.NormalizeExecutionReport
func (m *MockNormalizer) NormalizeExecutionReport(ctx context.Context, raw []byte) (*venuesv1.ExecutionReport, error) {
	if m.ErrorResponse != nil {
		return nil, m.ErrorResponse
	}
	return &venuesv1.ExecutionReport{
		OrderId: StringPtr("mock-order-1"),
	}, nil
}

// NormalizeBalance implements Normalizer.NormalizeBalance
func (m *MockNormalizer) NormalizeBalance(ctx context.Context, raw []byte) (*venuesv1.Balance, error) {
	if m.ErrorResponse != nil {
		return nil, m.ErrorResponse
	}
	return &venuesv1.Balance{
		VenueId: StringPtr("mock-venue"),
	}, nil
}

// NormalizeOrderBook implements Normalizer.NormalizeOrderBook
func (m *MockNormalizer) NormalizeOrderBook(ctx context.Context, raw []byte) (*marketsv1.OrderBook, error) {
	if m.ErrorResponse != nil {
		return nil, m.ErrorResponse
	}
	return &marketsv1.OrderBook{}, nil
}

// NormalizeTrade implements Normalizer.NormalizeTrade
func (m *MockNormalizer) NormalizeTrade(ctx context.Context, raw []byte) (*marketsv1.Trade, error) {
	if m.ErrorResponse != nil {
		return nil, m.ErrorResponse
	}
	return &marketsv1.Trade{}, nil
}

// NormalizeError implements Normalizer.NormalizeError
func (m *MockNormalizer) NormalizeError(ctx context.Context, raw []byte) error {
	if m.ErrorResponse != nil {
		return m.ErrorResponse
	}
	return nil
}

// TestNormalizerInterface tests the Normalizer interface contract
func TestNormalizerInterface(t *testing.T) {
	ctx := context.Background()
	normalizer := &MockNormalizer{}

	t.Run("NormalizeOrder returns Order", func(t *testing.T) {
		order, err := normalizer.NormalizeOrder(ctx, []byte(`{}`))
		assert.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, "mock-order-1", *order.OrderId)
	})

	t.Run("NormalizeExecutionReport returns ExecutionReport", func(t *testing.T) {
		report, err := normalizer.NormalizeExecutionReport(ctx, []byte(`{}`))
		assert.NoError(t, err)
		assert.NotNil(t, report)
		assert.Equal(t, "mock-order-1", *report.OrderId)
	})

	t.Run("NormalizeBalance returns Balance", func(t *testing.T) {
		balance, err := normalizer.NormalizeBalance(ctx, []byte(`{}`))
		assert.NoError(t, err)
		assert.NotNil(t, balance)
		assert.Equal(t, "mock-venue", *balance.VenueId)
	})

	t.Run("NormalizeOrderBook returns OrderBook", func(t *testing.T) {
		book, err := normalizer.NormalizeOrderBook(ctx, []byte(`{}`))
		assert.NoError(t, err)
		assert.NotNil(t, book)
	})

	t.Run("NormalizeTrade returns Trade", func(t *testing.T) {
		trade, err := normalizer.NormalizeTrade(ctx, []byte(`{}`))
		assert.NoError(t, err)
		assert.NotNil(t, trade)
	})

	t.Run("NormalizeError handles error responses", func(t *testing.T) {
		err := normalizer.NormalizeError(ctx, []byte(`{"error": "test"}`))
		assert.NoError(t, err) // Mock returns no error
	})
}

// TestMockNormalizerImplementsInterface ensures compile-time interface satisfaction
func TestMockNormalizerImplementsInterface(t *testing.T) {
	var _ Normalizer = (*MockNormalizer)(nil)
}
