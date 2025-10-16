package coinbase

import (
	"context"
	"encoding/json"
	"fmt"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	"github.com/Combine-Capital/cqvx/internal/normalizer"
)

// CoinbaseTrade represents a Coinbase trade/match response.
// This matches the structure returned by the Coinbase Advanced Trade API ticker or trades endpoint.
//
// Reference: https://docs.cdp.coinbase.com/advanced-trade-api/
type CoinbaseTrade struct {
	TradeID   string `json:"trade_id"`
	ProductID string `json:"product_id"`
	Price     string `json:"price"`
	Size      string `json:"size"`
	Time      string `json:"time"`
	Side      string `json:"side"` // "BUY" or "SELL" from taker perspective
	Bid       string `json:"bid"`
	Ask       string `json:"ask"`
}

// CoinbaseTradesResponse represents the response from the trades endpoint.
type CoinbaseTradesResponse struct {
	Trades  []CoinbaseTrade `json:"trades"`
	BestBid string          `json:"best_bid"`
	BestAsk string          `json:"best_ask"`
}

// NormalizeTrade converts a Coinbase trade JSON response to a CQC Trade protobuf.
//
// The function handles:
//   - Parsing JSON response (single trade or trades list)
//   - Converting timestamps to protobuf format
//   - Parsing decimal values for prices and quantities
//   - Mapping trade side (BUY/SELL from taker perspective)
//   - Calculating trade value
//
// Returns an error if JSON parsing fails or required fields are missing.
func NormalizeTrade(ctx context.Context, raw []byte) (*marketsv1.Trade, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty trade response")
	}

	// Try parsing as trades list first
	var tradesResp CoinbaseTradesResponse
	if err := json.Unmarshal(raw, &tradesResp); err == nil && len(tradesResp.Trades) > 0 {
		// Return the first trade from the list
		return normalizeSingleTrade(tradesResp.Trades[0])
	}

	// Try parsing as single trade
	var trade CoinbaseTrade
	if err := json.Unmarshal(raw, &trade); err != nil {
		return nil, fmt.Errorf("failed to parse coinbase trade: %w", err)
	}

	return normalizeSingleTrade(trade)
}

// NormalizeTrades converts a Coinbase trades JSON response to multiple CQC Trade protos.
// This is useful when processing multiple trades from a single API response.
func NormalizeTrades(ctx context.Context, raw []byte) ([]*marketsv1.Trade, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty trades response")
	}

	var tradesResp CoinbaseTradesResponse
	if err := json.Unmarshal(raw, &tradesResp); err != nil {
		return nil, fmt.Errorf("failed to parse coinbase trades: %w", err)
	}

	if len(tradesResp.Trades) == 0 {
		return []*marketsv1.Trade{}, nil
	}

	trades := make([]*marketsv1.Trade, 0, len(tradesResp.Trades))
	for _, cbTrade := range tradesResp.Trades {
		trade, err := normalizeSingleTrade(cbTrade)
		if err != nil {
			// Log error but continue processing other trades
			continue
		}
		trades = append(trades, trade)
	}

	return trades, nil
}

// normalizeSingleTrade converts a single Coinbase trade to a CQC Trade protobuf.
func normalizeSingleTrade(cbTrade CoinbaseTrade) (*marketsv1.Trade, error) {
	// Parse timestamp
	timestamp, err := normalizer.ParseTimestamp(cbTrade.Time)
	if err != nil {
		return nil, fmt.Errorf("invalid trade time: %w", err)
	}

	// Parse decimal fields
	price := normalizer.ParseDecimalOrZero(cbTrade.Price)
	quantity := normalizer.ParseDecimalOrZero(cbTrade.Size)

	// Calculate value
	value := price * quantity

	// Parse side
	side := parseTradeSide(cbTrade.Side)

	// Build CQC Trade
	venueId := "coinbase"
	trade := &marketsv1.Trade{
		TradeId:     &cbTrade.TradeID,
		VenueId:     &venueId,
		VenueSymbol: &cbTrade.ProductID,
		Timestamp:   timestamp,
		Price:       &price,
		Quantity:    &quantity,
		Side:        &side,
		Value:       &value,
	}

	return trade, nil
}

// parseTradeSide converts Coinbase trade side string to CQC TradeSide enum.
// Coinbase returns "BUY" or "SELL" from the taker's perspective.
func parseTradeSide(side string) marketsv1.TradeSide {
	switch side {
	case "BUY":
		return marketsv1.TradeSide_TRADE_SIDE_BUY
	case "SELL":
		return marketsv1.TradeSide_TRADE_SIDE_SELL
	default:
		return marketsv1.TradeSide_TRADE_SIDE_UNSPECIFIED
	}
}
