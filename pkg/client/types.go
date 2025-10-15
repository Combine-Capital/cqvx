package client

import (
	"errors"
	"time"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
)

// Error definitions for filter validation
var (
	ErrInvalidLimit     = errors.New("limit must be non-negative")
	ErrInvalidOffset    = errors.New("offset must be non-negative")
	ErrInvalidTimeRange = errors.New("start time must be before end time")
)

// OrderFilter defines filter criteria for querying orders.
// All fields are optional. If not specified, no filtering is applied for that field.
type OrderFilter struct {
	// Symbols filters orders by trading pair (e.g., "BTC-USD", "ETH-USD").
	// If empty, orders for all symbols are returned.
	Symbols []string

	// Statuses filters orders by their current status.
	// If empty, orders with any status are returned.
	Statuses []venuesv1.OrderStatus

	// StartTime filters orders created on or after this time.
	// If zero, no lower bound is applied.
	StartTime time.Time

	// EndTime filters orders created before this time.
	// If zero, no upper bound is applied.
	EndTime time.Time

	// Limit specifies the maximum number of orders to return.
	// If zero, venue-specific default limit is used.
	// Some venues may have maximum limits that override this value.
	Limit int

	// Offset specifies the number of orders to skip (for pagination).
	// If zero, starts from the first order.
	Offset int
}

// Validate checks if the filter has valid values.
// Returns an error if the filter configuration is invalid.
func (f *OrderFilter) Validate() error {
	if f.Limit < 0 {
		return ErrInvalidLimit
	}
	if f.Offset < 0 {
		return ErrInvalidOffset
	}
	if !f.StartTime.IsZero() && !f.EndTime.IsZero() && f.StartTime.After(f.EndTime) {
		return ErrInvalidTimeRange
	}
	return nil
}

// HasTimeRange returns true if the filter specifies a time range.
func (f *OrderFilter) HasTimeRange() bool {
	return !f.StartTime.IsZero() || !f.EndTime.IsZero()
}

// HasSymbolFilter returns true if the filter specifies symbols.
func (f *OrderFilter) HasSymbolFilter() bool {
	return len(f.Symbols) > 0
}

// HasStatusFilter returns true if the filter specifies statuses.
func (f *OrderFilter) HasStatusFilter() bool {
	return len(f.Statuses) > 0
}

// OrderBookHandler is a callback function for order book update events.
// Implementations receive order book snapshots or updates as they occur.
type OrderBookHandler func(orderBook *marketsv1.OrderBook) error

// TradeHandler is a callback function for trade events.
// Implementations receive trade notifications as they occur.
type TradeHandler func(trade *marketsv1.Trade) error
