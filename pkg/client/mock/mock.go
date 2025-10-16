// Package mock provides a deterministic mock implementation of VenueClient
// for testing consuming services without requiring live venue connections.
package mock

import (
	"context"
	"fmt"
	"sync"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/Combine-Capital/cqvx/pkg/client"
)

// Ensure Client implements the VenueClient interface at compile time
var _ client.VenueClient = (*Client)(nil)

// Client is a mock implementation of the VenueClient interface.
// It provides configurable behaviors for each method and tracks call counts
// for assertion in tests.
//
// Thread-safe: All methods can be called concurrently.
//
// Example usage:
//
//	mock := &mock.Client{}
//	mock.OnPlaceOrder = func(ctx context.Context, order *venuesv1.Order) (*venuesv1.ExecutionReport, error) {
//	    return &venuesv1.ExecutionReport{
//	        OrderId: "test-order-123",
//	        Status:  venuesv1.OrderStatus_ORDER_STATUS_NEW,
//	    }, nil
//	}
//
//	report, err := mock.PlaceOrder(ctx, order)
//	assert.NoError(t, err)
//	assert.Equal(t, 1, mock.PlaceOrderCallCount())
type Client struct {
	mu sync.RWMutex

	// Configurable method behaviors - set these to control mock responses
	OnPlaceOrder         func(ctx context.Context, order *venuesv1.Order) (*venuesv1.ExecutionReport, error)
	OnCancelOrder        func(ctx context.Context, orderID string) (*venuesv1.OrderStatus, error)
	OnGetOrder           func(ctx context.Context, orderID string) (*venuesv1.Order, error)
	OnGetOrders          func(ctx context.Context, filter client.OrderFilter) ([]*venuesv1.Order, error)
	OnGetBalance         func(ctx context.Context) (*venuesv1.Balance, error)
	OnGetOrderBook       func(ctx context.Context, symbol string) (*marketsv1.OrderBook, error)
	OnSubscribeOrderBook func(ctx context.Context, symbol string, handler client.OrderBookHandler) error
	OnSubscribeTrades    func(ctx context.Context, symbol string, handler client.TradeHandler) error
	OnHealth             func(ctx context.Context) error

	// Call tracking - tracks arguments for each call
	placeOrderCalls         []placeOrderCall
	cancelOrderCalls        []cancelOrderCall
	getOrderCalls           []getOrderCall
	getOrdersCalls          []getOrdersCall
	getBalanceCalls         []getBalanceCall
	getOrderBookCalls       []getOrderBookCall
	subscribeOrderBookCalls []subscribeOrderBookCall
	subscribeTradesCalls    []subscribeTradesCall
	healthCalls             []healthCall
}

// Call tracking types
type placeOrderCall struct {
	ctx   context.Context
	order *venuesv1.Order
}

type cancelOrderCall struct {
	ctx     context.Context
	orderID string
}

type getOrderCall struct {
	ctx     context.Context
	orderID string
}

type getOrdersCall struct {
	ctx    context.Context
	filter client.OrderFilter
}

type getBalanceCall struct {
	ctx context.Context
}

type getOrderBookCall struct {
	ctx    context.Context
	symbol string
}

type subscribeOrderBookCall struct {
	ctx     context.Context
	symbol  string
	handler client.OrderBookHandler
}

type subscribeTradesCall struct {
	ctx     context.Context
	symbol  string
	handler client.TradeHandler
}

type healthCall struct {
	ctx context.Context
}

// PlaceOrder submits a new order. Calls the configured OnPlaceOrder handler if set.
// If OnPlaceOrder is not set, returns a default ExecutionReport.
func (c *Client) PlaceOrder(ctx context.Context, order *venuesv1.Order) (*venuesv1.ExecutionReport, error) {
	c.mu.Lock()
	c.placeOrderCalls = append(c.placeOrderCalls, placeOrderCall{ctx: ctx, order: order})
	handler := c.OnPlaceOrder
	c.mu.Unlock()

	if handler != nil {
		return handler(ctx, order)
	}

	// Default behavior: return a successful execution report
	orderID := fmt.Sprintf("mock-order-%d", len(c.placeOrderCalls))
	return &venuesv1.ExecutionReport{
		ExecutionId:   &orderID,
		OrderId:       order.OrderId,
		VenueSymbol:   order.VenueSymbol,
		ExecutionType: venuesv1.ExecutionType_EXECUTION_TYPE_NEW.Enum(),
		OrderStatus:   stringPtr("NEW"),
		Price:         order.Price,
		Quantity:      order.Quantity,
	}, nil
}

// CancelOrder cancels an existing order. Calls the configured OnCancelOrder handler if set.
func (c *Client) CancelOrder(ctx context.Context, orderID string) (*venuesv1.OrderStatus, error) {
	c.mu.Lock()
	c.cancelOrderCalls = append(c.cancelOrderCalls, cancelOrderCall{ctx: ctx, orderID: orderID})
	handler := c.OnCancelOrder
	c.mu.Unlock()

	if handler != nil {
		return handler(ctx, orderID)
	}

	// Default behavior: return cancelled status
	status := venuesv1.OrderStatus_ORDER_STATUS_CANCELLED
	return &status, nil
}

// GetOrder retrieves order details. Calls the configured OnGetOrder handler if set.
func (c *Client) GetOrder(ctx context.Context, orderID string) (*venuesv1.Order, error) {
	c.mu.Lock()
	c.getOrderCalls = append(c.getOrderCalls, getOrderCall{ctx: ctx, orderID: orderID})
	handler := c.OnGetOrder
	c.mu.Unlock()

	if handler != nil {
		return handler(ctx, orderID)
	}

	// Default behavior: return a basic order
	return &venuesv1.Order{
		OrderId:     &orderID,
		Status:      venuesv1.OrderStatus_ORDER_STATUS_OPEN.Enum(),
		VenueSymbol: stringPtr("BTC-USD"),
	}, nil
}

// GetOrders retrieves multiple orders. Calls the configured OnGetOrders handler if set.
func (c *Client) GetOrders(ctx context.Context, filter client.OrderFilter) ([]*venuesv1.Order, error) {
	c.mu.Lock()
	c.getOrdersCalls = append(c.getOrdersCalls, getOrdersCall{ctx: ctx, filter: filter})
	handler := c.OnGetOrders
	c.mu.Unlock()

	if handler != nil {
		return handler(ctx, filter)
	}

	// Default behavior: return an empty slice
	return []*venuesv1.Order{}, nil
}

// GetBalance retrieves account balance. Calls the configured OnGetBalance handler if set.
func (c *Client) GetBalance(ctx context.Context) (*venuesv1.Balance, error) {
	c.mu.Lock()
	c.getBalanceCalls = append(c.getBalanceCalls, getBalanceCall{ctx: ctx})
	handler := c.OnGetBalance
	c.mu.Unlock()

	if handler != nil {
		return handler(ctx)
	}

	// Default behavior: return an empty balance
	return &venuesv1.Balance{}, nil
}

// GetOrderBook retrieves order book snapshot. Calls the configured OnGetOrderBook handler if set.
func (c *Client) GetOrderBook(ctx context.Context, symbol string) (*marketsv1.OrderBook, error) {
	c.mu.Lock()
	c.getOrderBookCalls = append(c.getOrderBookCalls, getOrderBookCall{ctx: ctx, symbol: symbol})
	handler := c.OnGetOrderBook
	c.mu.Unlock()

	if handler != nil {
		return handler(ctx, symbol)
	}

	// Default behavior: return an empty order book
	return &marketsv1.OrderBook{
		VenueSymbol: stringPtr(symbol),
		Bids:        []*marketsv1.OrderBookLevel{},
		Asks:        []*marketsv1.OrderBookLevel{},
	}, nil
}

// SubscribeOrderBook subscribes to order book updates. Calls the configured OnSubscribeOrderBook handler if set.
func (c *Client) SubscribeOrderBook(ctx context.Context, symbol string, handler client.OrderBookHandler) error {
	c.mu.Lock()
	c.subscribeOrderBookCalls = append(c.subscribeOrderBookCalls, subscribeOrderBookCall{
		ctx:     ctx,
		symbol:  symbol,
		handler: handler,
	})
	onSubscribe := c.OnSubscribeOrderBook
	c.mu.Unlock()

	if onSubscribe != nil {
		return onSubscribe(ctx, symbol, handler)
	}

	// Default behavior: do nothing (subscription succeeds but no updates are sent)
	return nil
}

// SubscribeTrades subscribes to trade updates. Calls the configured OnSubscribeTrades handler if set.
func (c *Client) SubscribeTrades(ctx context.Context, symbol string, handler client.TradeHandler) error {
	c.mu.Lock()
	c.subscribeTradesCalls = append(c.subscribeTradesCalls, subscribeTradesCall{
		ctx:     ctx,
		symbol:  symbol,
		handler: handler,
	})
	onSubscribe := c.OnSubscribeTrades
	c.mu.Unlock()

	if onSubscribe != nil {
		return onSubscribe(ctx, symbol, handler)
	}

	// Default behavior: do nothing (subscription succeeds but no updates are sent)
	return nil
}

// Health performs a health check. Calls the configured OnHealth handler if set.
func (c *Client) Health(ctx context.Context) error {
	c.mu.Lock()
	c.healthCalls = append(c.healthCalls, healthCall{ctx: ctx})
	handler := c.OnHealth
	c.mu.Unlock()

	if handler != nil {
		return handler(ctx)
	}

	// Default behavior: return nil (healthy)
	return nil
}

// Call count methods - return the number of times each method was called

// PlaceOrderCallCount returns the number of times PlaceOrder was called.
func (c *Client) PlaceOrderCallCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.placeOrderCalls)
}

// CancelOrderCallCount returns the number of times CancelOrder was called.
func (c *Client) CancelOrderCallCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cancelOrderCalls)
}

// GetOrderCallCount returns the number of times GetOrder was called.
func (c *Client) GetOrderCallCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.getOrderCalls)
}

// GetOrdersCallCount returns the number of times GetOrders was called.
func (c *Client) GetOrdersCallCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.getOrdersCalls)
}

// GetBalanceCallCount returns the number of times GetBalance was called.
func (c *Client) GetBalanceCallCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.getBalanceCalls)
}

// GetOrderBookCallCount returns the number of times GetOrderBook was called.
func (c *Client) GetOrderBookCallCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.getOrderBookCalls)
}

// SubscribeOrderBookCallCount returns the number of times SubscribeOrderBook was called.
func (c *Client) SubscribeOrderBookCallCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.subscribeOrderBookCalls)
}

// SubscribeTradesCallCount returns the number of times SubscribeTrades was called.
func (c *Client) SubscribeTradesCallCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.subscribeTradesCalls)
}

// HealthCallCount returns the number of times Health was called.
func (c *Client) HealthCallCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.healthCalls)
}

// Call argument retrieval methods - return arguments from specific calls

// PlaceOrderCall returns the arguments from the nth PlaceOrder call (0-indexed).
// Panics if n is out of bounds.
func (c *Client) PlaceOrderCall(n int) (context.Context, *venuesv1.Order) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n < 0 || n >= len(c.placeOrderCalls) {
		panic(fmt.Sprintf("PlaceOrderCall: index %d out of bounds (0-%d)", n, len(c.placeOrderCalls)-1))
	}
	call := c.placeOrderCalls[n]
	return call.ctx, call.order
}

// CancelOrderCall returns the arguments from the nth CancelOrder call (0-indexed).
func (c *Client) CancelOrderCall(n int) (context.Context, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n < 0 || n >= len(c.cancelOrderCalls) {
		panic(fmt.Sprintf("CancelOrderCall: index %d out of bounds (0-%d)", n, len(c.cancelOrderCalls)-1))
	}
	call := c.cancelOrderCalls[n]
	return call.ctx, call.orderID
}

// GetOrderCall returns the arguments from the nth GetOrder call (0-indexed).
func (c *Client) GetOrderCall(n int) (context.Context, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n < 0 || n >= len(c.getOrderCalls) {
		panic(fmt.Sprintf("GetOrderCall: index %d out of bounds (0-%d)", n, len(c.getOrderCalls)-1))
	}
	call := c.getOrderCalls[n]
	return call.ctx, call.orderID
}

// GetOrdersCall returns the arguments from the nth GetOrders call (0-indexed).
func (c *Client) GetOrdersCall(n int) (context.Context, client.OrderFilter) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n < 0 || n >= len(c.getOrdersCalls) {
		panic(fmt.Sprintf("GetOrdersCall: index %d out of bounds (0-%d)", n, len(c.getOrdersCalls)-1))
	}
	call := c.getOrdersCalls[n]
	return call.ctx, call.filter
}

// GetBalanceCall returns the arguments from the nth GetBalance call (0-indexed).
func (c *Client) GetBalanceCall(n int) context.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n < 0 || n >= len(c.getBalanceCalls) {
		panic(fmt.Sprintf("GetBalanceCall: index %d out of bounds (0-%d)", n, len(c.getBalanceCalls)-1))
	}
	return c.getBalanceCalls[n].ctx
}

// GetOrderBookCall returns the arguments from the nth GetOrderBook call (0-indexed).
func (c *Client) GetOrderBookCall(n int) (context.Context, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n < 0 || n >= len(c.getOrderBookCalls) {
		panic(fmt.Sprintf("GetOrderBookCall: index %d out of bounds (0-%d)", n, len(c.getOrderBookCalls)-1))
	}
	call := c.getOrderBookCalls[n]
	return call.ctx, call.symbol
}

// SubscribeOrderBookCall returns the arguments from the nth SubscribeOrderBook call (0-indexed).
func (c *Client) SubscribeOrderBookCall(n int) (context.Context, string, client.OrderBookHandler) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n < 0 || n >= len(c.subscribeOrderBookCalls) {
		panic(fmt.Sprintf("SubscribeOrderBookCall: index %d out of bounds (0-%d)", n, len(c.subscribeOrderBookCalls)-1))
	}
	call := c.subscribeOrderBookCalls[n]
	return call.ctx, call.symbol, call.handler
}

// SubscribeTradesCall returns the arguments from the nth SubscribeTrades call (0-indexed).
func (c *Client) SubscribeTradesCall(n int) (context.Context, string, client.TradeHandler) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n < 0 || n >= len(c.subscribeTradesCalls) {
		panic(fmt.Sprintf("SubscribeTradesCall: index %d out of bounds (0-%d)", n, len(c.subscribeTradesCalls)-1))
	}
	call := c.subscribeTradesCalls[n]
	return call.ctx, call.symbol, call.handler
}

// HealthCall returns the arguments from the nth Health call (0-indexed).
func (c *Client) HealthCall(n int) context.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if n < 0 || n >= len(c.healthCalls) {
		panic(fmt.Sprintf("HealthCall: index %d out of bounds (0-%d)", n, len(c.healthCalls)-1))
	}
	return c.healthCalls[n].ctx
}

// Reset clears all call history and configured handlers.
// Useful for reusing the same mock instance across multiple tests.
func (c *Client) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Clear handlers
	c.OnPlaceOrder = nil
	c.OnCancelOrder = nil
	c.OnGetOrder = nil
	c.OnGetOrders = nil
	c.OnGetBalance = nil
	c.OnGetOrderBook = nil
	c.OnSubscribeOrderBook = nil
	c.OnSubscribeTrades = nil
	c.OnHealth = nil

	// Clear call history
	c.placeOrderCalls = nil
	c.cancelOrderCalls = nil
	c.getOrderCalls = nil
	c.getOrdersCalls = nil
	c.getBalanceCalls = nil
	c.getOrderBookCalls = nil
	c.subscribeOrderBookCalls = nil
	c.subscribeTradesCalls = nil
	c.healthCalls = nil
}
