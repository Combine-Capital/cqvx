package coinbase

import (
	"context"
	"encoding/json"
	"fmt"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	"github.com/Combine-Capital/cqvx/internal/normalizer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CoinbaseOrderBook represents a Coinbase order book response (L2 data).
// This matches the structure returned by the Coinbase Advanced Trade API product book endpoint.
//
// Reference: https://docs.cdp.coinbase.com/advanced-trade-api/
type CoinbaseOrderBook struct {
	PriceBook PriceBook `json:"pricebook"`
	Time      string    `json:"time"`
}

// PriceBook contains the actual order book data.
type PriceBook struct {
	ProductID string          `json:"product_id"`
	Bids      [][]interface{} `json:"bids"` // [[price, size], ...]
	Asks      [][]interface{} `json:"asks"` // [[price, size], ...]
	Time      string          `json:"time"`
}

// NormalizeOrderBook converts a Coinbase order book JSON response to a CQC OrderBook protobuf.
//
// The function handles:
//   - Parsing JSON response
//   - Converting bid/ask arrays to OrderBookLevel protos
//   - Calculating best bid, best ask, spread, and mid price
//   - Parsing timestamps
//   - Sorting bids descending and asks ascending
//
// Returns an error if JSON parsing fails or data is malformed.
func NormalizeOrderBook(ctx context.Context, raw []byte) (*marketsv1.OrderBook, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty orderbook response")
	}

	var cbBook CoinbaseOrderBook
	if err := json.Unmarshal(raw, &cbBook); err != nil {
		return nil, fmt.Errorf("failed to parse coinbase orderbook: %w", err)
	}

	// Parse timestamp
	var timestamp *timestamppb.Timestamp
	timeStr := cbBook.Time
	if timeStr == "" {
		timeStr = cbBook.PriceBook.Time
	}
	if timeStr != "" {
		if ts, err := normalizer.ParseTimestamp(timeStr); err == nil {
			timestamp = ts
		} else {
			timestamp = timestamppb.Now()
		}
	} else {
		timestamp = timestamppb.Now()
	}

	// Convert bids
	bids, err := parseOrderBookLevels(cbBook.PriceBook.Bids)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bids: %w", err)
	}

	// Convert asks
	asks, err := parseOrderBookLevels(cbBook.PriceBook.Asks)
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
	venueId := "coinbase"
	orderBook := &marketsv1.OrderBook{
		VenueId:     &venueId,
		VenueSymbol: &cbBook.PriceBook.ProductID,
		Timestamp:   timestamp,
		Bids:        bids,
		Asks:        asks,
		BestBid:     bestBid,
		BestAsk:     bestAsk,
		Spread:      spread,
		MidPrice:    midPrice,
	}

	return orderBook, nil
}

// parseOrderBookLevels converts Coinbase order book level arrays to OrderBookLevel protos.
// Coinbase returns levels as arrays: [[price, size], [price, size], ...]
func parseOrderBookLevels(levels [][]interface{}) ([]*marketsv1.OrderBookLevel, error) {
	if levels == nil {
		return []*marketsv1.OrderBookLevel{}, nil
	}

	result := make([]*marketsv1.OrderBookLevel, 0, len(levels))

	for i, level := range levels {
		if len(level) < 2 {
			return nil, fmt.Errorf("invalid level at index %d: expected at least 2 elements, got %d", i, len(level))
		}

		// Parse price (first element)
		priceStr, ok := level[0].(string)
		if !ok {
			// Try float64
			if priceFloat, ok := level[0].(float64); ok {
				priceStr = fmt.Sprintf("%f", priceFloat)
			} else {
				return nil, fmt.Errorf("invalid price type at index %d: %T", i, level[0])
			}
		}
		price := normalizer.ParseDecimalOrZero(priceStr)

		// Parse quantity (second element)
		quantityStr, ok := level[1].(string)
		if !ok {
			// Try float64
			if quantityFloat, ok := level[1].(float64); ok {
				quantityStr = fmt.Sprintf("%f", quantityFloat)
			} else {
				return nil, fmt.Errorf("invalid quantity type at index %d: %T", i, level[1])
			}
		}
		quantity := normalizer.ParseDecimalOrZero(quantityStr)

		// Create level
		obLevel := &marketsv1.OrderBookLevel{
			Price:    &price,
			Quantity: &quantity,
		}

		// Some venues provide order count as third element
		if len(level) > 2 {
			if countFloat, ok := level[2].(float64); ok {
				count := int32(countFloat)
				obLevel.OrderCount = &count
			} else if countStr, ok := level[2].(string); ok {
				if countFloat := normalizer.ParseDecimalOrZero(countStr); countFloat > 0 {
					count := int32(countFloat)
					obLevel.OrderCount = &count
				}
			}
		}

		result = append(result, obLevel)
	}

	return result, nil
}
