package mock

import (
	"time"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrderBuilder provides a fluent interface for building test Order instances.
type OrderBuilder struct {
	order *venuesv1.Order
}

// NewOrderBuilder creates a new OrderBuilder with sensible defaults.
func NewOrderBuilder() *OrderBuilder {
	return &OrderBuilder{
		order: &venuesv1.Order{
			OrderId:     stringPtr("test-order-1"),
			VenueSymbol: stringPtr("BTC-USD"),
			OrderType:   venuesv1.OrderType_ORDER_TYPE_LIMIT.Enum(),
			Side:        venuesv1.OrderSide_ORDER_SIDE_BUY.Enum(),
			Status:      venuesv1.OrderStatus_ORDER_STATUS_OPEN.Enum(),
			TimeInForce: venuesv1.TimeInForce_TIME_IN_FORCE_GTC.Enum(),
			Quantity:    float64Ptr(1.0),
			Price:       float64Ptr(50000.0),
			CreatedAt:   timestamppb.Now(),
		},
	}
}

// WithOrderID sets the order ID.
func (b *OrderBuilder) WithOrderID(id string) *OrderBuilder {
	b.order.OrderId = stringPtr(id)
	return b
}

// WithVenueOrderID sets the venue order ID.
func (b *OrderBuilder) WithVenueOrderID(id string) *OrderBuilder {
	b.order.VenueOrderId = stringPtr(id)
	return b
}

// WithClientOrderID sets the client order ID.
func (b *OrderBuilder) WithClientOrderID(id string) *OrderBuilder {
	b.order.ClientOrderId = stringPtr(id)
	return b
}

// WithSymbol sets the venue symbol.
func (b *OrderBuilder) WithSymbol(symbol string) *OrderBuilder {
	b.order.VenueSymbol = stringPtr(symbol)
	return b
}

// WithOrderType sets the order type.
func (b *OrderBuilder) WithOrderType(orderType venuesv1.OrderType) *OrderBuilder {
	b.order.OrderType = orderType.Enum()
	return b
}

// WithSide sets the order side.
func (b *OrderBuilder) WithSide(side venuesv1.OrderSide) *OrderBuilder {
	b.order.Side = side.Enum()
	return b
}

// WithStatus sets the order status.
func (b *OrderBuilder) WithStatus(status venuesv1.OrderStatus) *OrderBuilder {
	b.order.Status = status.Enum()
	return b
}

// WithQuantity sets the order quantity.
func (b *OrderBuilder) WithQuantity(quantity float64) *OrderBuilder {
	b.order.Quantity = float64Ptr(quantity)
	return b
}

// WithPrice sets the order price.
func (b *OrderBuilder) WithPrice(price float64) *OrderBuilder {
	b.order.Price = float64Ptr(price)
	return b
}

// WithFilledQuantity sets the filled quantity.
func (b *OrderBuilder) WithFilledQuantity(filled float64) *OrderBuilder {
	b.order.FilledQuantity = float64Ptr(filled)
	return b
}

// Build returns the constructed Order.
func (b *OrderBuilder) Build() *venuesv1.Order {
	return b.order
}

// ExecutionReportBuilder provides a fluent interface for building test ExecutionReport instances.
type ExecutionReportBuilder struct {
	report *venuesv1.ExecutionReport
}

// NewExecutionReportBuilder creates a new ExecutionReportBuilder with sensible defaults.
func NewExecutionReportBuilder() *ExecutionReportBuilder {
	return &ExecutionReportBuilder{
		report: &venuesv1.ExecutionReport{
			ExecutionId:   stringPtr("test-exec-1"),
			OrderId:       stringPtr("test-order-1"),
			VenueSymbol:   stringPtr("BTC-USD"),
			ExecutionType: venuesv1.ExecutionType_EXECUTION_TYPE_NEW.Enum(),
			OrderStatus:   stringPtr("NEW"),
			Side:          stringPtr("BUY"),
			OrderType:     stringPtr("LIMIT"),
			Timestamp:     timestamppb.Now(),
			Quantity:      float64Ptr(1.0),
			Price:         float64Ptr(50000.0),
		},
	}
}

// WithExecutionID sets the execution ID.
func (b *ExecutionReportBuilder) WithExecutionID(id string) *ExecutionReportBuilder {
	b.report.ExecutionId = stringPtr(id)
	return b
}

// WithOrderID sets the order ID.
func (b *ExecutionReportBuilder) WithOrderID(id string) *ExecutionReportBuilder {
	b.report.OrderId = stringPtr(id)
	return b
}

// WithVenueOrderID sets the venue order ID.
func (b *ExecutionReportBuilder) WithVenueOrderID(id string) *ExecutionReportBuilder {
	b.report.VenueOrderId = stringPtr(id)
	return b
}

// WithSymbol sets the venue symbol.
func (b *ExecutionReportBuilder) WithSymbol(symbol string) *ExecutionReportBuilder {
	b.report.VenueSymbol = stringPtr(symbol)
	return b
}

// WithExecutionType sets the execution type.
func (b *ExecutionReportBuilder) WithExecutionType(execType venuesv1.ExecutionType) *ExecutionReportBuilder {
	b.report.ExecutionType = execType.Enum()
	return b
}

// WithOrderStatus sets the order status.
func (b *ExecutionReportBuilder) WithOrderStatus(status string) *ExecutionReportBuilder {
	b.report.OrderStatus = stringPtr(status)
	return b
}

// WithQuantity sets the execution quantity.
func (b *ExecutionReportBuilder) WithQuantity(quantity float64) *ExecutionReportBuilder {
	b.report.Quantity = float64Ptr(quantity)
	return b
}

// WithPrice sets the execution price.
func (b *ExecutionReportBuilder) WithPrice(price float64) *ExecutionReportBuilder {
	b.report.Price = float64Ptr(price)
	return b
}

// Build returns the constructed ExecutionReport.
func (b *ExecutionReportBuilder) Build() *venuesv1.ExecutionReport {
	return b.report
}

// BalanceBuilder provides a fluent interface for building test Balance instances.
type BalanceBuilder struct {
	balance *venuesv1.Balance
}

// NewBalanceBuilder creates a new BalanceBuilder with sensible defaults.
func NewBalanceBuilder() *BalanceBuilder {
	return &BalanceBuilder{
		balance: &venuesv1.Balance{
			AccountId: stringPtr("test-account"),
			AssetId:   stringPtr("BTC"),
			Available: float64Ptr(1.5),
			Total:     float64Ptr(2.0),
			Locked:    float64Ptr(0.5),
			Timestamp: timestamppb.Now(),
		},
	}
}

// WithAccountID sets the account ID.
func (b *BalanceBuilder) WithAccountID(id string) *BalanceBuilder {
	b.balance.AccountId = stringPtr(id)
	return b
}

// WithAssetID sets the asset ID.
func (b *BalanceBuilder) WithAssetID(assetID string) *BalanceBuilder {
	b.balance.AssetId = stringPtr(assetID)
	return b
}

// WithAvailable sets the available balance.
func (b *BalanceBuilder) WithAvailable(available float64) *BalanceBuilder {
	b.balance.Available = float64Ptr(available)
	return b
}

// WithTotal sets the total balance.
func (b *BalanceBuilder) WithTotal(total float64) *BalanceBuilder {
	b.balance.Total = float64Ptr(total)
	return b
}

// WithLocked sets the locked balance.
func (b *BalanceBuilder) WithLocked(locked float64) *BalanceBuilder {
	b.balance.Locked = float64Ptr(locked)
	return b
}

// Build returns the constructed Balance.
func (b *BalanceBuilder) Build() *venuesv1.Balance {
	return b.balance
}

// OrderBookBuilder provides a fluent interface for building test OrderBook instances.
type OrderBookBuilder struct {
	orderBook *marketsv1.OrderBook
}

// NewOrderBookBuilder creates a new OrderBookBuilder with sensible defaults.
func NewOrderBookBuilder() *OrderBookBuilder {
	return &OrderBookBuilder{
		orderBook: &marketsv1.OrderBook{
			VenueSymbol: stringPtr("BTC-USD"),
			Timestamp:   timestamppb.Now(),
			Bids:        []*marketsv1.OrderBookLevel{},
			Asks:        []*marketsv1.OrderBookLevel{},
		},
	}
}

// WithSymbol sets the venue symbol.
func (b *OrderBookBuilder) WithSymbol(symbol string) *OrderBookBuilder {
	b.orderBook.VenueSymbol = stringPtr(symbol)
	return b
}

// WithBid adds a bid level.
func (b *OrderBookBuilder) WithBid(price, quantity float64) *OrderBookBuilder {
	b.orderBook.Bids = append(b.orderBook.Bids, &marketsv1.OrderBookLevel{
		Price:    float64Ptr(price),
		Quantity: float64Ptr(quantity),
	})
	return b
}

// WithAsk adds an ask level.
func (b *OrderBookBuilder) WithAsk(price, quantity float64) *OrderBookBuilder {
	b.orderBook.Asks = append(b.orderBook.Asks, &marketsv1.OrderBookLevel{
		Price:    float64Ptr(price),
		Quantity: float64Ptr(quantity),
	})
	return b
}

// WithTimestamp sets the timestamp.
func (b *OrderBookBuilder) WithTimestamp(t time.Time) *OrderBookBuilder {
	b.orderBook.Timestamp = timestamppb.New(t)
	return b
}

// Build returns the constructed OrderBook.
func (b *OrderBookBuilder) Build() *marketsv1.OrderBook {
	return b.orderBook
}

// TradeBuilder provides a fluent interface for building test Trade instances.
type TradeBuilder struct {
	trade *marketsv1.Trade
}

// NewTradeBuilder creates a new TradeBuilder with sensible defaults.
func NewTradeBuilder() *TradeBuilder {
	return &TradeBuilder{
		trade: &marketsv1.Trade{
			TradeId:     stringPtr("test-trade-1"),
			VenueSymbol: stringPtr("BTC-USD"),
			Price:       float64Ptr(50000.0),
			Quantity:    float64Ptr(0.1),
			Side:        marketsv1.TradeSide_TRADE_SIDE_BUY.Enum(),
			Timestamp:   timestamppb.Now(),
		},
	}
}

// WithTradeID sets the trade ID.
func (b *TradeBuilder) WithTradeID(id string) *TradeBuilder {
	b.trade.TradeId = stringPtr(id)
	return b
}

// WithSymbol sets the venue symbol.
func (b *TradeBuilder) WithSymbol(symbol string) *TradeBuilder {
	b.trade.VenueSymbol = stringPtr(symbol)
	return b
}

// WithPrice sets the trade price.
func (b *TradeBuilder) WithPrice(price float64) *TradeBuilder {
	b.trade.Price = float64Ptr(price)
	return b
}

// WithQuantity sets the trade quantity.
func (b *TradeBuilder) WithQuantity(quantity float64) *TradeBuilder {
	b.trade.Quantity = float64Ptr(quantity)
	return b
}

// WithSide sets the trade side.
func (b *TradeBuilder) WithSide(side marketsv1.TradeSide) *TradeBuilder {
	b.trade.Side = side.Enum()
	return b
}

// WithTimestamp sets the timestamp.
func (b *TradeBuilder) WithTimestamp(t time.Time) *TradeBuilder {
	b.trade.Timestamp = timestamppb.New(t)
	return b
}

// Build returns the constructed Trade.
func (b *TradeBuilder) Build() *marketsv1.Trade {
	return b.trade
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func int64Ptr(i int64) *int64 {
	return &i
}
