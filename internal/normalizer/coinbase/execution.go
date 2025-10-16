package coinbase

import (
	"context"
	"encoding/json"
	"fmt"

	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/Combine-Capital/cqvx/internal/normalizer"
)

// CoinbaseFill represents a Coinbase fill/execution response.
// Fills represent individual trades that occurred to fill an order.
//
// Reference: https://docs.cdp.coinbase.com/advanced-trade-api/
type CoinbaseFill struct {
	EntryID            string `json:"entry_id"`
	TradeID            string `json:"trade_id"`
	OrderID            string `json:"order_id"`
	TradeTime          string `json:"trade_time"`
	TradeType          string `json:"trade_type"` // "FILL"
	Price              string `json:"price"`
	Size               string `json:"size"`
	Commission         string `json:"commission"`
	ProductID          string `json:"product_id"`
	SequenceTimestamp  string `json:"sequence_timestamp"`
	LiquidityIndicator string `json:"liquidity_indicator"` // "MAKER", "TAKER", "UNKNOWN"
	SizeInQuote        string `json:"size_in_quote"`
	UserID             string `json:"user_id"`
	Side               string `json:"side"`
	RetailPortfolioID  string `json:"retail_portfolio_id"`
}

// NormalizeExecutionReport converts a Coinbase fill JSON response to a CQC ExecutionReport protobuf.
//
// The function handles:
//   - Parsing JSON response
//   - Converting timestamps to protobuf format
//   - Parsing decimal values for prices, quantities, and fees
//   - Mapping liquidity indicators to maker/taker flags
//   - Determining execution type from trade type
//
// Returns an error if JSON parsing fails or required fields are missing.
func NormalizeExecutionReport(ctx context.Context, raw []byte) (*venuesv1.ExecutionReport, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty execution report response")
	}

	var cbFill CoinbaseFill
	if err := json.Unmarshal(raw, &cbFill); err != nil {
		return nil, fmt.Errorf("failed to parse coinbase fill: %w", err)
	}

	// Parse timestamp
	timestamp, err := normalizer.ParseTimestamp(cbFill.TradeTime)
	if err != nil {
		return nil, fmt.Errorf("invalid trade_time: %w", err)
	}

	// Parse decimal fields
	price := normalizer.ParseDecimalOrZero(cbFill.Price)
	quantity := normalizer.ParseDecimalOrZero(cbFill.Size)
	fee := normalizer.ParseDecimalOrZero(cbFill.Commission)

	// Calculate value
	value := price * quantity

	// Determine if maker or taker
	isMaker := cbFill.LiquidityIndicator == "MAKER"

	// Parse execution type
	executionType := parseExecutionType(cbFill.TradeType)

	// Parse order status (from fill, we know it's at least partially filled)
	orderStatus := "FILLED" // Default for fills

	// Build CQC ExecutionReport
	report := &venuesv1.ExecutionReport{
		ExecutionId:      &cbFill.EntryID,
		OrderId:          &cbFill.OrderID,
		VenueOrderId:     &cbFill.OrderID,
		VenueSymbol:      &cbFill.ProductID,
		ExecutionType:    &executionType,
		OrderStatus:      &orderStatus,
		Side:             &cbFill.Side,
		Timestamp:        timestamp,
		Price:            &price,
		Quantity:         &quantity,
		Fee:              &fee,
		TradeId:          &cbFill.TradeID,
		IsMaker:          &isMaker,
		Liquidity:        &cbFill.LiquidityIndicator,
		VenueExecutionId: &cbFill.EntryID,
		Value:            &value,
	}

	return report, nil
}

// parseExecutionType converts Coinbase trade type to CQC ExecutionType enum.
func parseExecutionType(tradeType string) venuesv1.ExecutionType {
	switch tradeType {
	case "FILL":
		return venuesv1.ExecutionType_EXECUTION_TYPE_TRADE
	default:
		return venuesv1.ExecutionType_EXECUTION_TYPE_UNSPECIFIED
	}
}
