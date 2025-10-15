// Package client defines the core VenueClient interface and common types
// for interacting with cryptocurrency trading venues.
package client

import (
	"context"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
)

// VenueClient defines the unified interface that all venue implementations must satisfy.
// This interface provides a consistent API for trading operations, account management,
// market data retrieval, and streaming subscriptions across different venues.
//
// All methods accept a context.Context for cancellation and timeout support.
// All methods return CQC protocol buffer types for type safety and consistency.
// Errors are CQI-typed for structured error handling.
//
// Implementations must handle venue-specific authentication, rate limiting,
// and response normalization internally.
type VenueClient interface {
	// Trading Operations

	// PlaceOrder submits a new order to the venue.
	// Returns an ExecutionReport containing the order status and details.
	// The order parameter must contain valid symbol, side, type, and quantity fields.
	//
	// Example:
	//   order := &venuesv1.Order{
	//       Symbol:   "BTC-USD",
	//       Side:     venuesv1.Side_SIDE_BUY,
	//       Type:     venuesv1.OrderType_ORDER_TYPE_LIMIT,
	//       Quantity: "0.01",
	//       Price:    "50000.00",
	//   }
	//   report, err := client.PlaceOrder(ctx, order)
	PlaceOrder(ctx context.Context, order *venuesv1.Order) (*venuesv1.ExecutionReport, error)

	// CancelOrder cancels an existing order by ID.
	// Returns the final order status after cancellation.
	// If the order is already filled or cancelled, may return an error.
	CancelOrder(ctx context.Context, orderID string) (*venuesv1.OrderStatus, error)

	// GetOrder retrieves the current state of a specific order by ID.
	// Returns the complete order details including fills and status.
	GetOrder(ctx context.Context, orderID string) (*venuesv1.Order, error)

	// GetOrders retrieves multiple orders based on the provided filter criteria.
	// The filter can specify time ranges, symbols, statuses, and pagination.
	// Returns a slice of orders matching the filter.
	GetOrders(ctx context.Context, filter OrderFilter) ([]*venuesv1.Order, error)

	// Account Operations

	// GetBalance retrieves the current account balance.
	// Returns balance information for all assets held at the venue.
	GetBalance(ctx context.Context) (*venuesv1.Balance, error)

	// Market Data Operations

	// GetOrderBook retrieves the current order book snapshot for a symbol.
	// Returns bids and asks with price and quantity information.
	GetOrderBook(ctx context.Context, symbol string) (*marketsv1.OrderBook, error)

	// Streaming Operations

	// SubscribeOrderBook establishes a streaming subscription to order book updates.
	// The handler callback is invoked for each order book update.
	// The subscription remains active until the context is cancelled or an error occurs.
	//
	// Note: Not all venues support streaming. Implementations may return an error
	// indicating unsupported operation (e.g., FalconX).
	SubscribeOrderBook(ctx context.Context, symbol string, handler OrderBookHandler) error

	// SubscribeTrades establishes a streaming subscription to trade updates.
	// The handler callback is invoked for each trade that occurs.
	// The subscription remains active until the context is cancelled or an error occurs.
	//
	// Note: Not all venues support streaming. Implementations may return an error
	// indicating unsupported operation.
	SubscribeTrades(ctx context.Context, symbol string, handler TradeHandler) error

	// Health Operations

	// Health performs a health check on the venue connection.
	// Returns nil if the venue is reachable and operational.
	// Returns an error if the venue is unreachable or experiencing issues.
	Health(ctx context.Context) error
}
