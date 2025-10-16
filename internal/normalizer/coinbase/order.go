// Package coinbase provides normalizers for Coinbase Advanced Trade API (v3).
// This is for the modern Advanced Trade API, not the legacy Coinbase Exchange v2 API.
package coinbase

import (
	"context"
	"encoding/json"
	"fmt"

	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/Combine-Capital/cqvx/internal/normalizer"
)

// CoinbaseOrder represents a Coinbase order response.
// This matches the structure returned by the Coinbase Advanced Trade API.
//
// Reference: https://docs.cdp.coinbase.com/advanced-trade-api/
type CoinbaseOrder struct {
	OrderID               string                     `json:"order_id"`
	ProductID             string                     `json:"product_id"`
	UserID                string                     `json:"user_id"`
	OrderConfiguration    CoinbaseOrderConfiguration `json:"order_configuration"`
	Side                  string                     `json:"side"`
	ClientOrderID         string                     `json:"client_order_id"`
	Status                string                     `json:"status"`
	TimeInForce           string                     `json:"time_in_force"`
	CreatedTime           string                     `json:"created_time"`
	CompletionPercentage  string                     `json:"completion_percentage"`
	FilledSize            string                     `json:"filled_size"`
	AverageFilledPrice    string                     `json:"average_filled_price"`
	Fee                   string                     `json:"fee"`
	NumberOfFills         string                     `json:"number_of_fills"`
	FilledValue           string                     `json:"filled_value"`
	PendingCancel         bool                       `json:"pending_cancel"`
	SizeInQuote           bool                       `json:"size_in_quote"`
	TotalFees             string                     `json:"total_fees"`
	SizeInclusiveOfFees   bool                       `json:"size_inclusive_of_fees"`
	TotalValueAfterFees   string                     `json:"total_value_after_fees"`
	TriggerStatus         string                     `json:"trigger_status"`
	OrderType             string                     `json:"order_type"`
	RejectReason          string                     `json:"reject_reason"`
	Settled               bool                       `json:"settled"`
	ProductType           string                     `json:"product_type"`
	RejectMessage         string                     `json:"reject_message"`
	CancelMessage         string                     `json:"cancel_message"`
	OrderPlacementSource  string                     `json:"order_placement_source"`
	OutstandingHoldAmount string                     `json:"outstanding_hold_amount"`
	IsLiquidation         bool                       `json:"is_liquidation"`
	LastFillTime          string                     `json:"last_fill_time"`
	EditHistory           []CoinbaseEditHistory      `json:"edit_history"`
	Leverage              string                     `json:"leverage"`
	MarginType            string                     `json:"margin_type"`
	RetailPortfolioID     string                     `json:"retail_portfolio_id"`
	OriginatingOrderID    string                     `json:"originating_order_id"`
	AttachedOrderID       string                     `json:"attached_order_id"`
}

// CoinbaseOrderConfiguration represents the order configuration from Coinbase.
// Only one field should be populated depending on the order type.
type CoinbaseOrderConfiguration struct {
	MarketMarketIOC       *CoinbaseMarketIOC      `json:"market_market_ioc"`
	SorLimitIOC           *CoinbaseSorLimitIOC    `json:"sor_limit_ioc"`
	LimitLimitGTC         *CoinbaseLimitGTC       `json:"limit_limit_gtc"`
	LimitLimitGTD         *CoinbaseLimitGTD       `json:"limit_limit_gtd"`
	LimitLimitFOK         *CoinbaseLimitFOK       `json:"limit_limit_fok"`
	StopLimitStopLimitGTC *CoinbaseStopLimitGTC   `json:"stop_limit_stop_limit_gtc"`
	StopLimitStopLimitGTD *CoinbaseStopLimitGTD   `json:"stop_limit_stop_limit_gtd"`
	TriggerBracketGTC     *CoinbaseTriggerBracket `json:"trigger_bracket_gtc"`
	TriggerBracketGTD     *CoinbaseTriggerBracket `json:"trigger_bracket_gtd"`
}

type CoinbaseMarketIOC struct {
	QuoteSize string `json:"quote_size"`
	BaseSize  string `json:"base_size"`
}

type CoinbaseSorLimitIOC struct {
	BaseSize   string `json:"base_size"`
	LimitPrice string `json:"limit_price"`
}

type CoinbaseLimitGTC struct {
	BaseSize   string `json:"base_size"`
	LimitPrice string `json:"limit_price"`
	PostOnly   bool   `json:"post_only"`
}

type CoinbaseLimitGTD struct {
	BaseSize   string `json:"base_size"`
	LimitPrice string `json:"limit_price"`
	EndTime    string `json:"end_time"`
	PostOnly   bool   `json:"post_only"`
}

type CoinbaseLimitFOK struct {
	BaseSize   string `json:"base_size"`
	LimitPrice string `json:"limit_price"`
}

type CoinbaseStopLimitGTC struct {
	BaseSize      string `json:"base_size"`
	LimitPrice    string `json:"limit_price"`
	StopPrice     string `json:"stop_price"`
	StopDirection string `json:"stop_direction"`
}

type CoinbaseStopLimitGTD struct {
	BaseSize      string `json:"base_size"`
	LimitPrice    string `json:"limit_price"`
	StopPrice     string `json:"stop_price"`
	EndTime       string `json:"end_time"`
	StopDirection string `json:"stop_direction"`
}

type CoinbaseTriggerBracket struct {
	BaseSize         string `json:"base_size"`
	LimitPrice       string `json:"limit_price"`
	StopTriggerPrice string `json:"stop_trigger_price"`
}

type CoinbaseEditHistory struct {
	Price                  string `json:"price"`
	Size                   string `json:"size"`
	ReplaceAcceptTimestamp string `json:"replace_accept_timestamp"`
}

// NormalizeOrder converts a Coinbase order JSON response to a CQC Order protobuf.
//
// The function handles:
//   - Parsing JSON response
//   - Mapping Coinbase order statuses to CQC enums
//   - Mapping Coinbase order types to CQC enums
//   - Converting timestamps to protobuf format
//   - Parsing decimal values for prices and quantities
//   - Extracting order configuration details (price, quantity)
//
// Returns an error if JSON parsing fails or required fields are missing.
func NormalizeOrder(ctx context.Context, raw []byte) (*venuesv1.Order, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty order response")
	}

	var cbOrder CoinbaseOrder
	if err := json.Unmarshal(raw, &cbOrder); err != nil {
		return nil, fmt.Errorf("failed to parse coinbase order: %w", err)
	}

	// Parse timestamps
	createdTime, err := normalizer.ParseTimestamp(cbOrder.CreatedTime)
	if err != nil {
		return nil, fmt.Errorf("invalid created_time: %w", err)
	}

	// Parse decimal fields
	filledSize := normalizer.ParseDecimalOrZero(cbOrder.FilledSize)
	avgPrice := normalizer.ParseDecimalOrZero(cbOrder.AverageFilledPrice)
	totalFees := normalizer.ParseDecimalOrZero(cbOrder.TotalFees)

	// Extract price and quantity from order configuration
	price, quantity := extractOrderConfiguration(cbOrder.OrderConfiguration)

	// Map Coinbase status to CQC status
	status := normalizer.ParseOrderStatus(cbOrder.Status)

	// Determine order type from configuration
	orderType := determineOrderType(cbOrder.OrderConfiguration, cbOrder.OrderType)

	// Parse side
	side := normalizer.ParseOrderSide(cbOrder.Side)

	// Parse time in force
	tif := normalizer.ParseTimeInForce(cbOrder.TimeInForce)

	// Build CQC Order with proper pointer types
	order := &venuesv1.Order{
		OrderId:          &cbOrder.OrderID,
		ClientOrderId:    &cbOrder.ClientOrderID,
		VenueOrderId:     &cbOrder.OrderID,
		VenueSymbol:      &cbOrder.ProductID,
		Side:             &side,
		OrderType:        &orderType,
		Quantity:         &quantity,
		Price:            &price,
		Status:           &status,
		TimeInForce:      &tif,
		FilledQuantity:   &filledSize,
		AverageFillPrice: &avgPrice,
		TotalFees:        &totalFees,
		CreatedAt:        createdTime,
	}

	// Add optional fields
	if cbOrder.RejectReason != "" {
		order.RejectionReason = &cbOrder.RejectReason
	}

	// Add post_only flag from configuration
	if config := cbOrder.OrderConfiguration; config.LimitLimitGTC != nil && config.LimitLimitGTC.PostOnly {
		postOnly := true
		order.PostOnly = &postOnly
	} else if config.LimitLimitGTD != nil && config.LimitLimitGTD.PostOnly {
		postOnly := true
		order.PostOnly = &postOnly
	}

	// Add last fill time if available (use as updated_at since Coinbase doesn't provide it)
	if cbOrder.LastFillTime != "" {
		if lastFillTime, err := normalizer.ParseTimestamp(cbOrder.LastFillTime); err == nil {
			order.UpdatedAt = lastFillTime
		}
	}

	return order, nil
}

// extractOrderConfiguration extracts price and quantity from the Coinbase order configuration.
// Different order types have different configuration structures.
func extractOrderConfiguration(config CoinbaseOrderConfiguration) (price, quantity float64) {
	switch {
	case config.MarketMarketIOC != nil:
		// Market orders may have quote_size (in quote currency) or base_size
		if config.MarketMarketIOC.QuoteSize != "" {
			quantity = normalizer.ParseDecimalOrZero(config.MarketMarketIOC.QuoteSize)
		} else if config.MarketMarketIOC.BaseSize != "" {
			quantity = normalizer.ParseDecimalOrZero(config.MarketMarketIOC.BaseSize)
		}
		price = 0.0 // Market orders don't have a price

	case config.SorLimitIOC != nil:
		price = normalizer.ParseDecimalOrZero(config.SorLimitIOC.LimitPrice)
		quantity = normalizer.ParseDecimalOrZero(config.SorLimitIOC.BaseSize)

	case config.LimitLimitGTC != nil:
		price = normalizer.ParseDecimalOrZero(config.LimitLimitGTC.LimitPrice)
		quantity = normalizer.ParseDecimalOrZero(config.LimitLimitGTC.BaseSize)

	case config.LimitLimitGTD != nil:
		price = normalizer.ParseDecimalOrZero(config.LimitLimitGTD.LimitPrice)
		quantity = normalizer.ParseDecimalOrZero(config.LimitLimitGTD.BaseSize)

	case config.LimitLimitFOK != nil:
		price = normalizer.ParseDecimalOrZero(config.LimitLimitFOK.LimitPrice)
		quantity = normalizer.ParseDecimalOrZero(config.LimitLimitFOK.BaseSize)

	case config.StopLimitStopLimitGTC != nil:
		price = normalizer.ParseDecimalOrZero(config.StopLimitStopLimitGTC.LimitPrice)
		quantity = normalizer.ParseDecimalOrZero(config.StopLimitStopLimitGTC.BaseSize)

	case config.StopLimitStopLimitGTD != nil:
		price = normalizer.ParseDecimalOrZero(config.StopLimitStopLimitGTD.LimitPrice)
		quantity = normalizer.ParseDecimalOrZero(config.StopLimitStopLimitGTD.BaseSize)

	case config.TriggerBracketGTC != nil:
		price = normalizer.ParseDecimalOrZero(config.TriggerBracketGTC.LimitPrice)
		quantity = normalizer.ParseDecimalOrZero(config.TriggerBracketGTC.BaseSize)

	case config.TriggerBracketGTD != nil:
		price = normalizer.ParseDecimalOrZero(config.TriggerBracketGTD.LimitPrice)
		quantity = normalizer.ParseDecimalOrZero(config.TriggerBracketGTD.BaseSize)
	}

	return price, quantity
}

// determineOrderType determines the CQC order type from Coinbase order configuration and type field.
func determineOrderType(config CoinbaseOrderConfiguration, orderTypeStr string) venuesv1.OrderType {
	// First check the order_type field if present
	if orderTypeStr != "" {
		parsedType := normalizer.ParseOrderType(orderTypeStr)
		if parsedType != venuesv1.OrderType_ORDER_TYPE_UNSPECIFIED {
			return parsedType
		}
	}

	// Fallback to inferring from configuration
	switch {
	case config.MarketMarketIOC != nil:
		return venuesv1.OrderType_ORDER_TYPE_MARKET

	case config.SorLimitIOC != nil:
		return venuesv1.OrderType_ORDER_TYPE_LIMIT // SOR Limit IOC is a type of limit order

	case config.LimitLimitGTC != nil:
		return venuesv1.OrderType_ORDER_TYPE_LIMIT

	case config.LimitLimitGTD != nil:
		return venuesv1.OrderType_ORDER_TYPE_LIMIT

	case config.LimitLimitFOK != nil:
		return venuesv1.OrderType_ORDER_TYPE_LIMIT

	case config.StopLimitStopLimitGTC != nil:
		return venuesv1.OrderType_ORDER_TYPE_STOP_LIMIT

	case config.StopLimitStopLimitGTD != nil:
		return venuesv1.OrderType_ORDER_TYPE_STOP_LIMIT

	case config.TriggerBracketGTC != nil, config.TriggerBracketGTD != nil:
		return venuesv1.OrderType_ORDER_TYPE_STOP_LIMIT

	default:
		return venuesv1.OrderType_ORDER_TYPE_UNSPECIFIED
	}
}
