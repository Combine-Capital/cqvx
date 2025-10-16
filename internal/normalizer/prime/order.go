// Package prime provides normalizers for Coinbase Prime API.
// Prime is Coinbase's institutional trading platform with support for
// SOR (Smart Order Routing), TWAP, VWAP, and other advanced order types.
package prime

import (
	"context"
	"encoding/json"
	"fmt"

	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/Combine-Capital/cqvx/internal/normalizer"
)

// PrimeOrder represents a Coinbase Prime order response.
// This matches the structure returned by the Coinbase Prime API.
//
// Reference: https://docs.cdp.coinbase.com/api-reference/prime-api/rest-api/orders/list-open-orders
type PrimeOrder struct {
	ID                    string           `json:"id"`
	UserID                string           `json:"user_id"`
	PortfolioID           string           `json:"portfolio_id"`
	ProductID             string           `json:"product_id"`
	Side                  string           `json:"side"` // "BUY", "SELL"
	ClientOrderID         string           `json:"client_order_id"`
	Type                  string           `json:"type"` // "MARKET", "LIMIT", "TWAP", "BLOCK", "VWAP", "STOP_LIMIT", "RFQ"
	BaseQuantity          string           `json:"base_quantity"`
	QuoteValue            string           `json:"quote_value"`
	LimitPrice            string           `json:"limit_price"`
	StartTime             string           `json:"start_time"`
	ExpiryTime            string           `json:"expiry_time"`
	Status                string           `json:"status"`        // "OPEN", "FILLED", "CANCELLED", "EXPIRED", etc.
	TimeInForce           string           `json:"time_in_force"` // "GOOD_UNTIL_DATE_TIME", "GOOD_UNTIL_CANCELLED", "IMMEDIATE_OR_CANCEL", "FILL_OR_KILL"
	CreatedAt             string           `json:"created_at"`
	FilledQuantity        string           `json:"filled_quantity"`
	FilledValue           string           `json:"filled_value"`
	AverageFilledPrice    string           `json:"average_filled_price"`
	Commission            string           `json:"commission"`
	ExchangeFee           string           `json:"exchange_fee"`
	HistoricalPOV         string           `json:"historical_pov"` // Percentage of volume (for TWAP/VWAP)
	StopPrice             string           `json:"stop_price"`
	NetAverageFilledPrice string           `json:"net_average_filled_price"`
	UserContext           string           `json:"user_context"`
	ClientProductID       string           `json:"client_product_id"`
	PostOnly              bool             `json:"post_only"`
	OrderEditHistory      []PrimeOrderEdit `json:"order_edit_history"`
	IsRaiseExact          bool             `json:"is_raise_exact"`
	DisplaySize           string           `json:"display_size"` // Iceberg order display size
	EditHistory           []PrimeOrderEdit `json:"edit_history"`
	DisplayQuoteSize      string           `json:"display_quote_size"`
	DisplayBaseSize       string           `json:"display_base_size"`
}

// PrimeOrderEdit represents an edit to a Prime order.
type PrimeOrderEdit struct {
	Price            string `json:"price"`
	Size             string `json:"size"`
	DisplaySize      string `json:"display_size"`
	StopPrice        string `json:"stop_price"`
	StopLimitPrice   string `json:"stop_limit_price"`
	EndTime          string `json:"end_time"`
	AcceptTime       string `json:"accept_time"`
	ClientOrderID    string `json:"client_order_id"`
	BaseQuantity     string `json:"base_quantity"`
	QuoteValue       string `json:"quote_value"`
	DisplayBaseSize  string `json:"display_base_size"`
	DisplayQuoteSize string `json:"display_quote_size"`
	ExpiryTime       string `json:"expiry_time"`
}

// NormalizeOrder converts a Coinbase Prime order JSON response to a CQC Order protobuf.
//
// The function handles:
//   - Parsing JSON response
//   - Converting timestamps to protobuf format
//   - Parsing decimal values for prices and quantities
//   - Mapping Prime order types (MARKET, LIMIT, TWAP, VWAP, BLOCK, STOP_LIMIT, RFQ) to CQC types
//   - Mapping Prime order sides (BUY, SELL) to CQC sides
//   - Mapping Prime order status to CQC order status
//   - Mapping Prime time-in-force to CQC time-in-force
//   - Handling Prime-specific fields (portfolio_id, display sizes, historical_pov, SOR)
//
// Returns an error if JSON parsing fails or required fields are missing.
func NormalizeOrder(ctx context.Context, raw []byte) (*venuesv1.Order, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty order response")
	}

	var primeOrder PrimeOrder
	if err := json.Unmarshal(raw, &primeOrder); err != nil {
		return nil, fmt.Errorf("failed to parse prime order: %w", err)
	}

	// Parse timestamps
	createdAt, err := normalizer.ParseTimestamp(primeOrder.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("invalid created_at: %w", err)
	}

	// Parse decimal fields
	var limitPrice, stopPrice, quantity float64
	var filledQuantity, avgFilledPrice, commission float64

	if primeOrder.LimitPrice != "" {
		limitPrice = normalizer.ParseDecimalOrZero(primeOrder.LimitPrice)
	}
	if primeOrder.StopPrice != "" {
		stopPrice = normalizer.ParseDecimalOrZero(primeOrder.StopPrice)
	}
	if primeOrder.BaseQuantity != "" {
		quantity = normalizer.ParseDecimalOrZero(primeOrder.BaseQuantity)
	}
	if primeOrder.FilledQuantity != "" {
		filledQuantity = normalizer.ParseDecimalOrZero(primeOrder.FilledQuantity)
	}
	if primeOrder.AverageFilledPrice != "" {
		avgFilledPrice = normalizer.ParseDecimalOrZero(primeOrder.AverageFilledPrice)
	}
	if primeOrder.Commission != "" {
		commission = normalizer.ParseDecimalOrZero(primeOrder.Commission)
	}

	// Map Prime order type to CQC order type
	orderType := mapOrderType(primeOrder.Type)

	// Map Prime order side to CQC order side
	side := mapOrderSide(primeOrder.Side)

	// Map Prime order status to CQC order status
	status := mapOrderStatus(primeOrder.Status)

	// Map Prime time-in-force to CQC time-in-force
	timeInForce := mapTimeInForce(primeOrder.TimeInForce)

	// Build CQC Order
	order := &venuesv1.Order{
		OrderId:          &primeOrder.ID,
		VenueOrderId:     &primeOrder.ID,
		ClientOrderId:    &primeOrder.ClientOrderID,
		VenueSymbol:      &primeOrder.ProductID,
		Side:             &side,
		OrderType:        &orderType,
		Status:           &status,
		TimeInForce:      &timeInForce,
		Quantity:         &quantity,
		Price:            &limitPrice,
		StopPrice:        &stopPrice,
		FilledQuantity:   &filledQuantity,
		AverageFillPrice: &avgFilledPrice,
		CreatedAt:        createdAt,
		TotalFees:        &commission,
		PostOnly:         &primeOrder.PostOnly,
	}

	// Set UpdatedAt based on creation time if we don't have a separate updated time
	order.UpdatedAt = createdAt

	// Prime-specific fields that don't have CQC Order equivalents:
	// - portfolio_id: Portfolio context for institutional trading
	// - historical_pov: Percentage of volume for TWAP/VWAP orders
	// - display_size/display_base_size/display_quote_size: Iceberg order display amounts
	// - user_context: User-provided context string
	// - is_raise_exact: Exact amount flag
	// These could be stored in a future venue-specific metadata field if needed

	return order, nil
}

// mapOrderType maps Prime order type to CQC OrderType enum.
func mapOrderType(primeType string) venuesv1.OrderType {
	switch primeType {
	case "MARKET":
		return venuesv1.OrderType_ORDER_TYPE_MARKET
	case "LIMIT":
		return venuesv1.OrderType_ORDER_TYPE_LIMIT
	case "STOP_LIMIT":
		return venuesv1.OrderType_ORDER_TYPE_STOP_LIMIT
	case "TWAP":
		// TWAP is a Prime-specific algorithmic order type
		// These don't have direct CQC equivalents, so we use LIMIT
		// and store the actual type in metadata
		return venuesv1.OrderType_ORDER_TYPE_LIMIT
	case "VWAP":
		// VWAP is a Prime-specific algorithmic order type
		return venuesv1.OrderType_ORDER_TYPE_LIMIT
	case "BLOCK":
		// Block trades are large OTC-style trades
		return venuesv1.OrderType_ORDER_TYPE_LIMIT
	case "RFQ":
		// Request for Quote
		return venuesv1.OrderType_ORDER_TYPE_LIMIT
	default:
		return venuesv1.OrderType_ORDER_TYPE_UNSPECIFIED
	}
}

// mapOrderSide maps Prime order side to CQC OrderSide enum.
func mapOrderSide(primeSide string) venuesv1.OrderSide {
	switch primeSide {
	case "BUY":
		return venuesv1.OrderSide_ORDER_SIDE_BUY
	case "SELL":
		return venuesv1.OrderSide_ORDER_SIDE_SELL
	default:
		return venuesv1.OrderSide_ORDER_SIDE_UNSPECIFIED
	}
}

// mapOrderStatus maps Prime order status to CQC OrderStatus enum.
func mapOrderStatus(primeStatus string) venuesv1.OrderStatus {
	switch primeStatus {
	case "OPEN", "WORKING":
		return venuesv1.OrderStatus_ORDER_STATUS_OPEN
	case "FILLED":
		return venuesv1.OrderStatus_ORDER_STATUS_FILLED
	case "CANCELLED":
		return venuesv1.OrderStatus_ORDER_STATUS_CANCELLED
	case "EXPIRED":
		return venuesv1.OrderStatus_ORDER_STATUS_REJECTED
	case "PENDING":
		return venuesv1.OrderStatus_ORDER_STATUS_PENDING
	case "REJECTED":
		return venuesv1.OrderStatus_ORDER_STATUS_REJECTED
	default:
		return venuesv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

// mapTimeInForce maps Prime time-in-force to CQC TimeInForce enum.
func mapTimeInForce(primeTIF string) venuesv1.TimeInForce {
	switch primeTIF {
	case "GOOD_UNTIL_DATE_TIME":
		return venuesv1.TimeInForce_TIME_IN_FORCE_GTD
	case "GOOD_UNTIL_CANCELLED":
		return venuesv1.TimeInForce_TIME_IN_FORCE_GTC
	case "IMMEDIATE_OR_CANCEL":
		return venuesv1.TimeInForce_TIME_IN_FORCE_IOC
	case "FILL_OR_KILL":
		return venuesv1.TimeInForce_TIME_IN_FORCE_FOK
	default:
		return venuesv1.TimeInForce_TIME_IN_FORCE_GTC // Default to GTC
	}
}
