// Package normalizer provides interfaces and utilities for converting
// venue-specific responses to CQC protocol buffer types.
package normalizer

import (
	"context"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
)

// Normalizer defines the interface for converting venue-specific responses
// to CQC protocol buffer types. Each venue implementation should provide
// a concrete implementation of this interface.
//
// Thread-safety: Implementations must be safe for concurrent use.
//
// Example implementations:
//   - Coinbase Exchange normalizer (internal/normalizer/coinbase)
//   - Coinbase Prime normalizer (internal/normalizer/prime)
//   - FalconX normalizer (internal/normalizer/falconx)
//   - Fordefi normalizer (internal/normalizer/fordefi)
type Normalizer interface {
	// NormalizeOrder converts a venue-specific order response to a CQC Order.
	//
	// The raw parameter contains the venue's JSON/XML response bytes.
	// Returns a fully populated CQC Order or an error if normalization fails.
	//
	// Common normalization tasks:
	//   - Map venue field names to CQC field names
	//   - Convert venue order types/statuses to CQC enums
	//   - Parse timestamps to protobuf Timestamp
	//   - Parse decimal values to appropriate numeric types
	//
	// Example:
	//   order, err := normalizer.NormalizeOrder(ctx, coinbaseOrderJSON)
	NormalizeOrder(ctx context.Context, raw []byte) (*venuesv1.Order, error)

	// NormalizeExecutionReport converts a venue-specific execution/fill response
	// to a CQC ExecutionReport.
	//
	// ExecutionReports represent order state changes, including:
	//   - Order acknowledgements
	//   - Partial and full fills
	//   - Cancellations
	//   - Rejections
	//
	// The raw parameter contains the venue's JSON/XML response bytes.
	// Returns a fully populated CQC ExecutionReport or an error if normalization fails.
	NormalizeExecutionReport(ctx context.Context, raw []byte) (*venuesv1.ExecutionReport, error)

	// NormalizeBalance converts a venue-specific balance/account response
	// to a CQC Balance.
	//
	// Balance information includes:
	//   - Available balances per asset
	//   - Held/locked balances
	//   - Total balances
	//
	// The raw parameter contains the venue's JSON/XML response bytes.
	// Returns a fully populated CQC Balance or an error if normalization fails.
	//
	// Note: Some venues return balance per asset, others return all assets.
	// Implementations should handle both patterns.
	NormalizeBalance(ctx context.Context, raw []byte) (*venuesv1.Balance, error)

	// NormalizeOrderBook converts a venue-specific order book response
	// to a CQC OrderBook.
	//
	// Order books contain:
	//   - Bids (buy orders) with price and quantity
	//   - Asks (sell orders) with price and quantity
	//   - Timestamp of the snapshot
	//
	// The raw parameter contains the venue's JSON/XML response bytes.
	// Returns a fully populated CQC OrderBook or an error if normalization fails.
	//
	// Common variations:
	//   - L2 order books (aggregated by price level)
	//   - L3 order books (individual orders)
	//   - Different depth levels (top 10, 50, 100, etc.)
	NormalizeOrderBook(ctx context.Context, raw []byte) (*marketsv1.OrderBook, error)

	// NormalizeTrade converts a venue-specific trade/match response
	// to a CQC Trade.
	//
	// Trades represent executed transactions with:
	//   - Price at which the trade occurred
	//   - Quantity/size of the trade
	//   - Timestamp of the trade
	//   - Trade direction (buy/sell from taker perspective)
	//
	// The raw parameter contains the venue's JSON/XML response bytes.
	// Returns a fully populated CQC Trade or an error if normalization fails.
	NormalizeTrade(ctx context.Context, raw []byte) (*marketsv1.Trade, error)

	// NormalizeError converts a venue-specific error response to a structured error.
	//
	// This method should parse venue error codes and messages, mapping them to
	// appropriate error types (e.g., rate limit, invalid request, authentication failure).
	//
	// The raw parameter contains the venue's error response bytes.
	// Returns an error with proper context and type information.
	//
	// Common error types to map:
	//   - Rate limit errors (temporary, should retry with backoff)
	//   - Authentication errors (permanent, credentials invalid)
	//   - Invalid request errors (permanent, bad parameters)
	//   - Temporary errors (network issues, venue downtime)
	//
	// Example:
	//   err := normalizer.NormalizeError(ctx, errorResponseJSON)
	//   if errors.Is(err, ErrRateLimit) {
	//       // Apply backoff and retry
	//   }
	NormalizeError(ctx context.Context, raw []byte) error
}
