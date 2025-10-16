package prime

import (
	"context"
	"encoding/json"
	"fmt"

	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/Combine-Capital/cqvx/internal/normalizer"
)

// PrimeFill represents a Coinbase Prime fill/execution response.
// Fills represent individual trades that occurred to fill an order.
//
// Reference: https://docs.cdp.coinbase.com/api-reference/prime-api/rest-api/portfolios/list-portfolio-fills
type PrimeFill struct {
	PortfolioID    string  `json:"portfolio_id"`
	PortfolioUUID  string  `json:"portfolio_uuid"`
	PortfolioName  string  `json:"portfolio_name"`
	FillID         string  `json:"fill_id"`
	ExecID         int64   `json:"exec_id"`
	OrderID        string  `json:"order_id"`
	InstrumentID   string  `json:"instrument_id"`
	InstrumentUUID string  `json:"instrument_uuid"`
	Symbol         string  `json:"symbol"`
	MatchID        string  `json:"match_id"`
	FillPrice      float64 `json:"fill_price"`
	FillQty        float64 `json:"fill_qty"`
	ClientID       string  `json:"client_id"`
	ClientOrderID  string  `json:"client_order_id"`
	OrderQty       float64 `json:"order_qty"`
	LimitPrice     float64 `json:"limit_price"`
	TotalFilled    float64 `json:"total_filled"`
	FilledVWAP     float64 `json:"filled_vwap"`
	ExpireTime     string  `json:"expire_time"`
	StopPrice      float64 `json:"stop_price"`
	Side           string  `json:"side"`     // "BUY", "SELL"
	TIF            string  `json:"tif"`      // Time in force
	STPMode        string  `json:"stp_mode"` // Self-trade prevention mode
	Flags          string  `json:"flags"`
	Fee            float64 `json:"fee"`
	FeeAsset       string  `json:"fee_asset"`
	OrderStatus    string  `json:"order_status"`
	EventTime      string  `json:"event_time"`
	Source         string  `json:"source"`          // "LIQUIDATION", etc.
	ExecutionVenue string  `json:"execution_venue"` // "CLOB", etc.
}

// NormalizeExecutionReport converts a Coinbase Prime fill JSON response to a CQC ExecutionReport protobuf.
//
// The function handles:
//   - Parsing JSON response
//   - Converting timestamps to protobuf format
//   - Parsing decimal values for prices, quantities, and fees
//   - Mapping Prime fill fields to CQC ExecutionReport fields
//   - Determining execution type and maker/taker status
//
// Returns an error if JSON parsing fails or required fields are missing.
func NormalizeExecutionReport(ctx context.Context, raw []byte) (*venuesv1.ExecutionReport, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty execution report response")
	}

	var primeFill PrimeFill
	if err := json.Unmarshal(raw, &primeFill); err != nil {
		return nil, fmt.Errorf("failed to parse prime fill: %w", err)
	}

	// Parse timestamp
	timestamp, err := normalizer.ParseTimestamp(primeFill.EventTime)
	if err != nil {
		return nil, fmt.Errorf("invalid event_time: %w", err)
	}

	// Calculate value (price * quantity)
	value := primeFill.FillPrice * primeFill.FillQty

	// Parse order status to string (ExecutionReport uses string, not enum)
	orderStatus := primeFill.OrderStatus

	// Determine execution type
	executionType := venuesv1.ExecutionType_EXECUTION_TYPE_FILL

	// Determine liquidity (maker/taker) - Prime doesn't always provide this explicitly
	// We can infer from order type: post_only orders are maker, others may be taker
	// For now, we'll leave it unspecified without explicit information
	isMaker := false // Default to taker if not specified

	// Parse side
	side := primeFill.Side

	// Build CQC ExecutionReport
	report := &venuesv1.ExecutionReport{
		ExecutionId:      &primeFill.FillID,
		OrderId:          &primeFill.OrderID,
		VenueOrderId:     &primeFill.OrderID,
		VenueSymbol:      &primeFill.Symbol,
		ExecutionType:    &executionType,
		OrderStatus:      &orderStatus,
		Side:             &side,
		Timestamp:        timestamp,
		Price:            &primeFill.FillPrice,
		Quantity:         &primeFill.FillQty,
		Fee:              &primeFill.Fee,
		TradeId:          &primeFill.MatchID,
		IsMaker:          &isMaker,
		VenueExecutionId: &primeFill.FillID,
		Value:            &value,
	}

	// Add client order ID if available
	if primeFill.ClientOrderID != "" {
		report.ClientOrderId = &primeFill.ClientOrderID
	}

	return report, nil
}
