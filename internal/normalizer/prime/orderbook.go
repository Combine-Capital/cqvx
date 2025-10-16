package prime

import (
	"context"
	"encoding/json"
	"fmt"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	"github.com/Combine-Capital/cqvx/internal/normalizer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// PrimeOrderBook represents a Coinbase Prime order book response (L2 data).
// Prime order books follow a similar structure to Advanced Trade but may include
// aggregated liquidity from multiple sources via SOR.
//
// Reference: https://docs.cdp.coinbase.com/prime/websocket-feed/ (market data)
type PrimeOrderBook struct {
	ProductID string          `json:"product_id"`
	Bids      [][]interface{} `json:"bids"` // [[price, size], ...]
	Asks      [][]interface{} `json:"asks"` // [[price, size], ...]
	Time      string          `json:"time"`
	Sequence  int64           `json:"sequence"` // Sequence number for ordering updates
}

// NormalizeOrderBook converts a Coinbase Prime order book JSON response to a CQC OrderBook protobuf.
//
// The function handles:
//   - Parsing JSON response
//   - Converting bid/ask arrays to OrderBookLevel protos
//   - Calculating best bid, best ask, spread, and mid price
//   - Parsing timestamps
//   - Handling Prime-specific sequencing
//
// Returns an error if JSON parsing fails or data is malformed.
func NormalizeOrderBook(ctx context.Context, raw []byte) (*marketsv1.OrderBook, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty orderbook response")
	}

	var primeBook PrimeOrderBook
	if err := json.Unmarshal(raw, &primeBook); err != nil {
		return nil, fmt.Errorf("failed to parse prime orderbook: %w", err)
	}

	// Parse timestamp
	var timestamp *timestamppb.Timestamp
	if primeBook.Time != "" {
		if ts, err := normalizer.ParseTimestamp(primeBook.Time); err == nil {
			timestamp = ts
		} else {
			timestamp = timestamppb.Now()
		}
	} else {
		timestamp = timestamppb.Now()
	}

	// Convert bids
	bids, err := parseOrderBookLevels(primeBook.Bids)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bids: %w", err)
	}

	// Convert asks
	asks, err := parseOrderBookLevels(primeBook.Asks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse asks: %w", err)
	}

	// Calculate best bid, best ask, spread, mid price
	var bestBid, bestAsk, spread, midPrice *float64
	if len(bids) > 0 && bids[0].Price != nil {
		bestBid = bids[0].Price
	}
	if len(asks) > 0 && asks[0].Price != nil {
		bestAsk = asks[0].Price
	}
	if bestBid != nil && bestAsk != nil {
		spreadVal := *bestAsk - *bestBid
		spread = &spreadVal
		midPriceVal := (*bestBid + *bestAsk) / 2.0
		midPrice = &midPriceVal
	}

	// Build CQC OrderBook
	venueId := "prime"
	orderBook := &marketsv1.OrderBook{
		VenueId:     &venueId,
		VenueSymbol: &primeBook.ProductID,
		Timestamp:   timestamp,
		Bids:        bids,
		Asks:        asks,
		BestBid:     bestBid,
		BestAsk:     bestAsk,
		Spread:      spread,
		MidPrice:    midPrice,
	}

	// Add sequence number if available (useful for maintaining order book state)
	if primeBook.Sequence > 0 {
		orderBook.Sequence = &primeBook.Sequence
	}

	return orderBook, nil
}

// parseOrderBookLevels converts raw bid/ask arrays to OrderBookLevel protos.
// Each level is expected to be [price, size] or [price, size, num_orders].
func parseOrderBookLevels(levels [][]interface{}) ([]*marketsv1.OrderBookLevel, error) {
	result := make([]*marketsv1.OrderBookLevel, 0, len(levels))

	for i, level := range levels {
		if len(level) < 2 {
			return nil, fmt.Errorf("level %d: expected at least 2 elements, got %d", i, len(level))
		}

		// Parse price (can be string or number)
		var price float64
		switch v := level[0].(type) {
		case string:
			price = normalizer.ParseDecimalOrZero(v)
		case float64:
			price = v
		case int:
			price = float64(v)
		case int64:
			price = float64(v)
		default:
			return nil, fmt.Errorf("level %d: invalid price type %T", i, v)
		}

		// Parse size (can be string or number)
		var size float64
		switch v := level[1].(type) {
		case string:
			size = normalizer.ParseDecimalOrZero(v)
		case float64:
			size = v
		case int:
			size = float64(v)
		case int64:
			size = float64(v)
		default:
			return nil, fmt.Errorf("level %d: invalid size type %T", i, v)
		}

		// Skip empty levels
		if price == 0 && size == 0 {
			continue
		}

		result = append(result, &marketsv1.OrderBookLevel{
			Price:    &price,
			Quantity: &size,
		})
	}

	return result, nil
}
