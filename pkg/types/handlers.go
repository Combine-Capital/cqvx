package types

import (
	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
)

// OrderBookHandler is a callback function invoked when an order book update is received.
// The handler receives a complete order book snapshot or incremental update.
//
// Implementations should:
//   - Process the update quickly to avoid blocking the streaming connection
//   - Return an error to signal that the subscription should be terminated
//   - Handle nil checks if the venue may send empty updates
//
// Example:
//
//	handler := func(book *marketsv1.OrderBook) error {
//	    log.Printf("Order book for %s: %d bids, %d asks",
//	        book.Symbol, len(book.Bids), len(book.Asks))
//	    return nil
//	}
type OrderBookHandler func(orderBook *marketsv1.OrderBook) error

// TradeHandler is a callback function invoked when a trade event is received.
// The handler receives information about trades that have executed on the venue.
//
// Implementations should:
//   - Process trades quickly to avoid blocking the streaming connection
//   - Return an error to signal that the subscription should be terminated
//   - Handle venue-specific trade types appropriately
//
// Example:
//
//	handler := func(trade *marketsv1.Trade) error {
//	    log.Printf("Trade: %s %s @ %s",
//	        trade.Quantity, trade.Symbol, trade.Price)
//	    return nil
//	}
type TradeHandler func(trade *marketsv1.Trade) error

// ExecutionHandler is a callback function for execution report updates.
// This is used for streaming order status updates from venues that support it.
//
// Example:
//
//	handler := func(report *venuesv1.ExecutionReport) error {
//	    log.Printf("Order %s: %s", report.OrderId, report.Status)
//	    return nil
//	}
//
// Note: Not used in MVP, but included for future venue implementations.
type ExecutionHandler func(report interface{}) error

// BalanceHandler is a callback function for balance updates.
// Some venues provide real-time balance updates via streaming.
//
// Note: Not used in MVP, but included for future venue implementations.
type BalanceHandler func(balance interface{}) error

// ErrorHandler is a callback function for handling errors during streaming.
// This allows consumers to implement custom error handling logic.
//
// If the handler returns true, the stream will attempt to reconnect.
// If it returns false, the stream will terminate.
type ErrorHandler func(err error) bool
