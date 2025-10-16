package mock_test

import (
	"context"
	"errors"
	"testing"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/Combine-Capital/cqvx/pkg/client"
	"github.com/Combine-Capital/cqvx/pkg/client/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockClientImplementsInterface verifies that the mock client satisfies the VenueClient interface.
func TestMockClientImplementsInterface(t *testing.T) {
	var _ client.VenueClient = (*mock.Client)(nil)
}

// TestPlaceOrder_DefaultBehavior tests the default behavior when OnPlaceOrder is not configured.
func TestPlaceOrder_DefaultBehavior(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()

	order := mock.NewOrderBuilder().
		WithSymbol("BTC-USD").
		WithQuantity(1.0).
		WithPrice(50000.0).
		Build()

	report, err := m.PlaceOrder(ctx, order)

	require.NoError(t, err)
	require.NotNil(t, report)
	assert.NotNil(t, report.ExecutionId)
	assert.Equal(t, 1, m.PlaceOrderCallCount())
}

// TestPlaceOrder_ConfiguredHandler tests PlaceOrder with a configured handler.
func TestPlaceOrder_ConfiguredHandler(t *testing.T) {
	m := &mock.Client{}
	expectedReport := mock.NewExecutionReportBuilder().
		WithOrderID("custom-order-123").
		WithSymbol("ETH-USD").
		Build()

	m.OnPlaceOrder = func(ctx context.Context, order *venuesv1.Order) (*venuesv1.ExecutionReport, error) {
		return expectedReport, nil
	}

	ctx := context.Background()
	order := mock.NewOrderBuilder().WithSymbol("ETH-USD").Build()

	report, err := m.PlaceOrder(ctx, order)

	require.NoError(t, err)
	assert.Equal(t, expectedReport, report)
	assert.Equal(t, 1, m.PlaceOrderCallCount())

	// Verify call arguments
	callCtx, callOrder := m.PlaceOrderCall(0)
	assert.Equal(t, ctx, callCtx)
	assert.Equal(t, order, callOrder)
}

// TestPlaceOrder_ErrorHandling tests PlaceOrder error handling.
func TestPlaceOrder_ErrorHandling(t *testing.T) {
	m := &mock.Client{}
	expectedErr := errors.New("order placement failed")

	m.OnPlaceOrder = func(ctx context.Context, order *venuesv1.Order) (*venuesv1.ExecutionReport, error) {
		return nil, expectedErr
	}

	ctx := context.Background()
	order := mock.NewOrderBuilder().Build()

	report, err := m.PlaceOrder(ctx, order)

	assert.ErrorIs(t, err, expectedErr)
	assert.Nil(t, report)
	assert.Equal(t, 1, m.PlaceOrderCallCount())
}

// TestCancelOrder_DefaultBehavior tests the default behavior when OnCancelOrder is not configured.
func TestCancelOrder_DefaultBehavior(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()
	orderID := "test-order-123"

	status, err := m.CancelOrder(ctx, orderID)

	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, venuesv1.OrderStatus_ORDER_STATUS_CANCELLED, *status)
	assert.Equal(t, 1, m.CancelOrderCallCount())
}

// TestCancelOrder_ConfiguredHandler tests CancelOrder with a configured handler.
func TestCancelOrder_ConfiguredHandler(t *testing.T) {
	m := &mock.Client{}
	expectedStatus := venuesv1.OrderStatus_ORDER_STATUS_FILLED

	m.OnCancelOrder = func(ctx context.Context, orderID string) (*venuesv1.OrderStatus, error) {
		return &expectedStatus, nil
	}

	ctx := context.Background()
	orderID := "test-order-123"

	status, err := m.CancelOrder(ctx, orderID)

	require.NoError(t, err)
	assert.Equal(t, expectedStatus, *status)
	assert.Equal(t, 1, m.CancelOrderCallCount())

	// Verify call arguments
	callCtx, callOrderID := m.CancelOrderCall(0)
	assert.Equal(t, ctx, callCtx)
	assert.Equal(t, orderID, callOrderID)
}

// TestGetOrder_DefaultBehavior tests the default behavior when OnGetOrder is not configured.
func TestGetOrder_DefaultBehavior(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()
	orderID := "test-order-123"

	order, err := m.GetOrder(ctx, orderID)

	require.NoError(t, err)
	require.NotNil(t, order)
	assert.Equal(t, &orderID, order.OrderId)
	assert.Equal(t, 1, m.GetOrderCallCount())
}

// TestGetOrder_ConfiguredHandler tests GetOrder with a configured handler.
func TestGetOrder_ConfiguredHandler(t *testing.T) {
	m := &mock.Client{}
	expectedOrder := mock.NewOrderBuilder().
		WithOrderID("test-order-456").
		WithSymbol("BTC-USD").
		Build()

	m.OnGetOrder = func(ctx context.Context, orderID string) (*venuesv1.Order, error) {
		return expectedOrder, nil
	}

	ctx := context.Background()

	order, err := m.GetOrder(ctx, "test-order-456")

	require.NoError(t, err)
	assert.Equal(t, expectedOrder, order)
	assert.Equal(t, 1, m.GetOrderCallCount())
}

// TestGetOrders_DefaultBehavior tests the default behavior when OnGetOrders is not configured.
func TestGetOrders_DefaultBehavior(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()
	filter := client.OrderFilter{
		Symbols: []string{"BTC-USD"},
		Limit:   10,
	}

	orders, err := m.GetOrders(ctx, filter)

	require.NoError(t, err)
	require.NotNil(t, orders)
	assert.Empty(t, orders)
	assert.Equal(t, 1, m.GetOrdersCallCount())

	// Verify call arguments
	callCtx, callFilter := m.GetOrdersCall(0)
	assert.Equal(t, ctx, callCtx)
	assert.Equal(t, filter, callFilter)
}

// TestGetOrders_ConfiguredHandler tests GetOrders with a configured handler.
func TestGetOrders_ConfiguredHandler(t *testing.T) {
	m := &mock.Client{}
	expectedOrders := []*venuesv1.Order{
		mock.NewOrderBuilder().WithOrderID("order-1").Build(),
		mock.NewOrderBuilder().WithOrderID("order-2").Build(),
	}

	m.OnGetOrders = func(ctx context.Context, filter client.OrderFilter) ([]*venuesv1.Order, error) {
		return expectedOrders, nil
	}

	ctx := context.Background()
	filter := client.OrderFilter{Limit: 10}

	orders, err := m.GetOrders(ctx, filter)

	require.NoError(t, err)
	assert.Equal(t, expectedOrders, orders)
	assert.Equal(t, 2, len(orders))
}

// TestGetBalance_DefaultBehavior tests the default behavior when OnGetBalance is not configured.
func TestGetBalance_DefaultBehavior(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()

	balance, err := m.GetBalance(ctx)

	require.NoError(t, err)
	require.NotNil(t, balance)
	assert.Equal(t, 1, m.GetBalanceCallCount())
}

// TestGetBalance_ConfiguredHandler tests GetBalance with a configured handler.
func TestGetBalance_ConfiguredHandler(t *testing.T) {
	m := &mock.Client{}
	expectedBalance := mock.NewBalanceBuilder().
		WithAccountID("test-account").
		WithAssetID("BTC").
		WithAvailable(1.5).
		Build()

	m.OnGetBalance = func(ctx context.Context) (*venuesv1.Balance, error) {
		return expectedBalance, nil
	}

	ctx := context.Background()

	balance, err := m.GetBalance(ctx)

	require.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)
	assert.Equal(t, "BTC", *balance.AssetId)
}

// TestGetOrderBook_DefaultBehavior tests the default behavior when OnGetOrderBook is not configured.
func TestGetOrderBook_DefaultBehavior(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()
	symbol := "BTC-USD"

	orderBook, err := m.GetOrderBook(ctx, symbol)

	require.NoError(t, err)
	require.NotNil(t, orderBook)
	assert.Equal(t, &symbol, orderBook.VenueSymbol)
	assert.Empty(t, orderBook.Bids)
	assert.Empty(t, orderBook.Asks)
	assert.Equal(t, 1, m.GetOrderBookCallCount())
}

// TestGetOrderBook_ConfiguredHandler tests GetOrderBook with a configured handler.
func TestGetOrderBook_ConfiguredHandler(t *testing.T) {
	m := &mock.Client{}
	expectedOrderBook := mock.NewOrderBookBuilder().
		WithSymbol("ETH-USD").
		WithBid(3000.0, 1.5).
		WithAsk(3001.0, 2.0).
		Build()

	m.OnGetOrderBook = func(ctx context.Context, symbol string) (*marketsv1.OrderBook, error) {
		return expectedOrderBook, nil
	}

	ctx := context.Background()

	orderBook, err := m.GetOrderBook(ctx, "ETH-USD")

	require.NoError(t, err)
	assert.Equal(t, expectedOrderBook, orderBook)
	assert.Equal(t, 1, len(orderBook.Bids))
	assert.Equal(t, 1, len(orderBook.Asks))
}

// TestSubscribeOrderBook_DefaultBehavior tests the default behavior when OnSubscribeOrderBook is not configured.
func TestSubscribeOrderBook_DefaultBehavior(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()
	symbol := "BTC-USD"
	handlerCalled := false

	handler := func(ob *marketsv1.OrderBook) error {
		handlerCalled = true
		return nil
	}

	err := m.SubscribeOrderBook(ctx, symbol, handler)

	require.NoError(t, err)
	assert.Equal(t, 1, m.SubscribeOrderBookCallCount())
	assert.False(t, handlerCalled, "Handler should not be called by default behavior")
}

// TestSubscribeOrderBook_ConfiguredHandler tests SubscribeOrderBook with a configured handler.
func TestSubscribeOrderBook_ConfiguredHandler(t *testing.T) {
	m := &mock.Client{}
	handlerCalled := false

	m.OnSubscribeOrderBook = func(ctx context.Context, symbol string, handler client.OrderBookHandler) error {
		// Simulate sending an order book update
		orderBook := mock.NewOrderBookBuilder().WithSymbol(symbol).Build()
		return handler(orderBook)
	}

	ctx := context.Background()

	handler := func(ob *marketsv1.OrderBook) error {
		handlerCalled = true
		assert.Equal(t, "BTC-USD", *ob.VenueSymbol)
		return nil
	}

	err := m.SubscribeOrderBook(ctx, "BTC-USD", handler)

	require.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, 1, m.SubscribeOrderBookCallCount())
}

// TestSubscribeTrades_DefaultBehavior tests the default behavior when OnSubscribeTrades is not configured.
func TestSubscribeTrades_DefaultBehavior(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()
	symbol := "BTC-USD"
	handlerCalled := false

	handler := func(trade *marketsv1.Trade) error {
		handlerCalled = true
		return nil
	}

	err := m.SubscribeTrades(ctx, symbol, handler)

	require.NoError(t, err)
	assert.Equal(t, 1, m.SubscribeTradesCallCount())
	assert.False(t, handlerCalled, "Handler should not be called by default behavior")
}

// TestSubscribeTrades_ConfiguredHandler tests SubscribeTrades with a configured handler.
func TestSubscribeTrades_ConfiguredHandler(t *testing.T) {
	m := &mock.Client{}
	tradeCount := 0

	m.OnSubscribeTrades = func(ctx context.Context, symbol string, handler client.TradeHandler) error {
		// Simulate sending multiple trade updates
		trade1 := mock.NewTradeBuilder().WithTradeID("trade-1").WithSymbol(symbol).Build()
		trade2 := mock.NewTradeBuilder().WithTradeID("trade-2").WithSymbol(symbol).Build()

		if err := handler(trade1); err != nil {
			return err
		}
		return handler(trade2)
	}

	ctx := context.Background()

	handler := func(trade *marketsv1.Trade) error {
		tradeCount++
		assert.Equal(t, "ETH-USD", *trade.VenueSymbol)
		return nil
	}

	err := m.SubscribeTrades(ctx, "ETH-USD", handler)

	require.NoError(t, err)
	assert.Equal(t, 2, tradeCount)
	assert.Equal(t, 1, m.SubscribeTradesCallCount())
}

// TestHealth_DefaultBehavior tests the default behavior when OnHealth is not configured.
func TestHealth_DefaultBehavior(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()

	err := m.Health(ctx)

	require.NoError(t, err)
	assert.Equal(t, 1, m.HealthCallCount())
}

// TestHealth_ConfiguredHandler tests Health with a configured handler.
func TestHealth_ConfiguredHandler(t *testing.T) {
	m := &mock.Client{}
	expectedErr := errors.New("venue unavailable")

	m.OnHealth = func(ctx context.Context) error {
		return expectedErr
	}

	ctx := context.Background()

	err := m.Health(ctx)

	assert.ErrorIs(t, err, expectedErr)
	assert.Equal(t, 1, m.HealthCallCount())
}

// TestReset tests that Reset clears all call history and handlers.
func TestReset(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()

	// Configure handlers and make calls
	m.OnPlaceOrder = func(ctx context.Context, order *venuesv1.Order) (*venuesv1.ExecutionReport, error) {
		return mock.NewExecutionReportBuilder().Build(), nil
	}
	_, _ = m.PlaceOrder(ctx, mock.NewOrderBuilder().Build())
	_, _ = m.GetBalance(ctx)

	assert.Equal(t, 1, m.PlaceOrderCallCount())
	assert.Equal(t, 1, m.GetBalanceCallCount())
	assert.NotNil(t, m.OnPlaceOrder)

	// Reset
	m.Reset()

	// Verify everything is cleared
	assert.Equal(t, 0, m.PlaceOrderCallCount())
	assert.Equal(t, 0, m.GetBalanceCallCount())
	assert.Nil(t, m.OnPlaceOrder)

	// Verify default behavior still works after reset
	_, err := m.PlaceOrder(ctx, mock.NewOrderBuilder().Build())
	assert.NoError(t, err)
	assert.Equal(t, 1, m.PlaceOrderCallCount())
}

// TestMultipleCalls verifies that call tracking works correctly for multiple calls.
func TestMultipleCalls(t *testing.T) {
	m := &mock.Client{}
	ctx := context.Background()

	// Make multiple calls
	_, _ = m.PlaceOrder(ctx, mock.NewOrderBuilder().WithOrderID("order-1").Build())
	_, _ = m.PlaceOrder(ctx, mock.NewOrderBuilder().WithOrderID("order-2").Build())
	_, _ = m.PlaceOrder(ctx, mock.NewOrderBuilder().WithOrderID("order-3").Build())

	assert.Equal(t, 3, m.PlaceOrderCallCount())

	// Verify individual call arguments
	_, order1 := m.PlaceOrderCall(0)
	_, order2 := m.PlaceOrderCall(1)
	_, order3 := m.PlaceOrderCall(2)

	assert.Equal(t, "order-1", *order1.OrderId)
	assert.Equal(t, "order-2", *order2.OrderId)
	assert.Equal(t, "order-3", *order3.OrderId)
}

// TestCallArgumentRetrieval_OutOfBounds verifies that out-of-bounds access panics.
func TestCallArgumentRetrieval_OutOfBounds(t *testing.T) {
	m := &mock.Client{}

	assert.Panics(t, func() {
		m.PlaceOrderCall(0)
	}, "Should panic when accessing call that doesn't exist")
}

// TestBuilders_OrderBuilder tests the OrderBuilder functionality.
func TestBuilders_OrderBuilder(t *testing.T) {
	order := mock.NewOrderBuilder().
		WithOrderID("test-order-1").
		WithSymbol("BTC-USD").
		WithOrderType(venuesv1.OrderType_ORDER_TYPE_MARKET).
		WithSide(venuesv1.OrderSide_ORDER_SIDE_SELL).
		WithQuantity(2.5).
		WithPrice(55000.0).
		WithFilledQuantity(1.0).
		Build()

	assert.Equal(t, "test-order-1", *order.OrderId)
	assert.Equal(t, "BTC-USD", *order.VenueSymbol)
	assert.Equal(t, venuesv1.OrderType_ORDER_TYPE_MARKET, *order.OrderType)
	assert.Equal(t, venuesv1.OrderSide_ORDER_SIDE_SELL, *order.Side)
	assert.Equal(t, 2.5, *order.Quantity)
	assert.Equal(t, 55000.0, *order.Price)
	assert.Equal(t, 1.0, *order.FilledQuantity)
}

// TestBuilders_ExecutionReportBuilder tests the ExecutionReportBuilder functionality.
func TestBuilders_ExecutionReportBuilder(t *testing.T) {
	report := mock.NewExecutionReportBuilder().
		WithExecutionID("exec-1").
		WithOrderID("order-1").
		WithSymbol("ETH-USD").
		WithQuantity(1.0).
		WithPrice(3000.0).
		Build()

	assert.Equal(t, "exec-1", *report.ExecutionId)
	assert.Equal(t, "order-1", *report.OrderId)
	assert.Equal(t, "ETH-USD", *report.VenueSymbol)
	assert.Equal(t, 1.0, *report.Quantity)
	assert.Equal(t, 3000.0, *report.Price)
}

// TestBuilders_BalanceBuilder tests the BalanceBuilder functionality.
func TestBuilders_BalanceBuilder(t *testing.T) {
	balance := mock.NewBalanceBuilder().
		WithAccountID("account-1").
		WithAssetID("ETH").
		WithAvailable(5.0).
		WithTotal(10.0).
		WithLocked(5.0).
		Build()

	assert.Equal(t, "account-1", *balance.AccountId)
	assert.Equal(t, "ETH", *balance.AssetId)
	assert.Equal(t, 5.0, *balance.Available)
	assert.Equal(t, 10.0, *balance.Total)
	assert.Equal(t, 5.0, *balance.Locked)
}

// TestBuilders_OrderBookBuilder tests the OrderBookBuilder functionality.
func TestBuilders_OrderBookBuilder(t *testing.T) {
	orderBook := mock.NewOrderBookBuilder().
		WithSymbol("BTC-USD").
		WithBid(49999.0, 1.5).
		WithBid(49998.0, 2.0).
		WithAsk(50001.0, 1.0).
		WithAsk(50002.0, 1.5).
		Build()

	assert.Equal(t, "BTC-USD", *orderBook.VenueSymbol)
	assert.Equal(t, 2, len(orderBook.Bids))
	assert.Equal(t, 2, len(orderBook.Asks))
	assert.Equal(t, 49999.0, *orderBook.Bids[0].Price)
	assert.Equal(t, 50001.0, *orderBook.Asks[0].Price)
}

// TestBuilders_TradeBuilder tests the TradeBuilder functionality.
func TestBuilders_TradeBuilder(t *testing.T) {
	trade := mock.NewTradeBuilder().
		WithTradeID("trade-1").
		WithSymbol("BTC-USD").
		WithPrice(50000.0).
		WithQuantity(0.5).
		WithSide(marketsv1.TradeSide_TRADE_SIDE_SELL).
		Build()

	assert.Equal(t, "trade-1", *trade.TradeId)
	assert.Equal(t, "BTC-USD", *trade.VenueSymbol)
	assert.Equal(t, 50000.0, *trade.Price)
	assert.Equal(t, 0.5, *trade.Quantity)
	assert.Equal(t, marketsv1.TradeSide_TRADE_SIDE_SELL, *trade.Side)
}
