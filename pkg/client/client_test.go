package client_test

import (
	"context"
	"testing"

	marketsv1 "github.com/Combine-Capital/cqc/gen/go/cqc/markets/v1"
	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/Combine-Capital/cqvx/pkg/client"
)

// mockVenueClient is a minimal mock implementation to verify interface compliance
type mockVenueClient struct{}

// Verify that mockVenueClient implements the VenueClient interface at compile time
var _ client.VenueClient = (*mockVenueClient)(nil)

func (m *mockVenueClient) PlaceOrder(ctx context.Context, order *venuesv1.Order) (*venuesv1.ExecutionReport, error) {
	return nil, nil
}

func (m *mockVenueClient) CancelOrder(ctx context.Context, orderID string) (*venuesv1.OrderStatus, error) {
	return nil, nil
}

func (m *mockVenueClient) GetOrder(ctx context.Context, orderID string) (*venuesv1.Order, error) {
	return nil, nil
}

func (m *mockVenueClient) GetOrders(ctx context.Context, filter client.OrderFilter) ([]*venuesv1.Order, error) {
	return nil, nil
}

func (m *mockVenueClient) GetBalance(ctx context.Context) (*venuesv1.Balance, error) {
	return nil, nil
}

func (m *mockVenueClient) GetOrderBook(ctx context.Context, symbol string) (*marketsv1.OrderBook, error) {
	return nil, nil
}

func (m *mockVenueClient) SubscribeOrderBook(ctx context.Context, symbol string, handler client.OrderBookHandler) error {
	return nil
}

func (m *mockVenueClient) SubscribeTrades(ctx context.Context, symbol string, handler client.TradeHandler) error {
	return nil
}

func (m *mockVenueClient) Health(ctx context.Context) error {
	return nil
}

// TestVenueClientInterface verifies that all interface methods are properly defined
func TestVenueClientInterface(t *testing.T) {
	var _ client.VenueClient = &mockVenueClient{}
	t.Log("VenueClient interface verification passed")
}

// TestVenueClientMethodSignatures tests that interface methods have correct signatures
func TestVenueClientMethodSignatures(t *testing.T) {
	ctx := context.Background()
	mock := &mockVenueClient{}

	// Test all method signatures compile
	_, _ = mock.PlaceOrder(ctx, nil)
	_, _ = mock.CancelOrder(ctx, "test-order-id")
	_, _ = mock.GetOrder(ctx, "test-order-id")
	_, _ = mock.GetOrders(ctx, client.OrderFilter{})
	_, _ = mock.GetBalance(ctx)
	_, _ = mock.GetOrderBook(ctx, "BTC-USD")
	_ = mock.SubscribeOrderBook(ctx, "BTC-USD", func(ob *marketsv1.OrderBook) error { return nil })
	_ = mock.SubscribeTrades(ctx, "BTC-USD", func(t *marketsv1.Trade) error { return nil })
	_ = mock.Health(ctx)

	t.Log("All VenueClient method signatures verified")
}
