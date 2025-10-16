# CQVX - Crypto Quant Venue Exchange

**Go client library for unified multi-venue crypto trading**

Part of the **Crypto Quant Platform** - Professional-grade crypto trading infrastructure.

[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## Overview

CQVX is a Go client library that provides a unified interface for interacting with multiple cryptocurrency trading venues. It abstracts venue-specific implementations behind a common `VenueClient` interface, enabling platform services to integrate with various exchanges, prime brokers, and liquidity providers through a single, consistent API.

### Key Features

- **Unified Interface**: Single `VenueClient` interface for all trading venues
- **Type-Safe**: All data types use CQC protocol buffers (no `interface{}`)
- **Production-Ready Infrastructure**: Built on CQI primitives (retry, circuit breaker, rate limiting, WebSocket auto-reconnect)
- **Mock Support**: Deterministic mock client for testing without live connections
- **Four MVP Venues**: Coinbase Exchange, Coinbase Prime, FalconX, Fordefi

### Supported Venues (MVP)

| Venue                 | Type          | Authentication | Streaming    | Order Types                    |
| --------------------- | ------------- | -------------- | ------------ | ------------------------------ |
| **Coinbase Exchange** | Retail spot   | HMAC-SHA256    | Yes          | Market, Limit                  |
| **Coinbase Prime**    | Institutional | JWT            | Yes          | Market, Limit, SOR, TWAP, VWAP |
| **FalconX**           | OTC RFQ       | Bearer Token   | No (polling) | RFQ                            |
| **Fordefi**           | Custody/MPC   | MPC Signing    | No           | Approval workflow              |

## Installation

```bash
go get github.com/Combine-Capital/cqvx
```

### Prerequisites

- Go 1.21 or higher
- CQC protocol buffer types (`github.com/Combine-Capital/cqc`)
- CQI infrastructure primitives (`github.com/Combine-Capital/cqi`)

## Quick Start

```go
package main

import (
    "context"
    "log"
    
    "github.com/Combine-Capital/cqvx/pkg/client"
    "github.com/Combine-Capital/cqvx/pkg/venues/coinbase"
    "github.com/Combine-Capital/cqc/venues/v1"
    "github.com/Combine-Capital/cqi/httpclient"
    "github.com/Combine-Capital/cqi/logging"
)

func main() {
    // Initialize CQI infrastructure
    logger := logging.New("cqvx-example")
    httpClient := httpclient.New(httpclient.Config{
        RateLimit: 10, // requests per second
    })
    
    // Create venue client
    config := coinbase.Config{
        APIKey:     "your-api-key",
        Secret:     "your-secret",
        Passphrase: "your-passphrase",
        BaseURL:    "https://api.coinbase.com",
    }
    
    client := coinbase.NewClient(config, httpClient, nil, logger)
    
    // Place an order
    order := &venues.Order{
        Symbol:   "BTC-USD",
        Side:     venues.Side_SIDE_BUY,
        Type:     venues.OrderType_ORDER_TYPE_LIMIT,
        Quantity: "0.01",
        Price:    "50000.00",
    }
    
    ctx := context.Background()
    report, err := client.PlaceOrder(ctx, order)
    if err != nil {
        log.Fatalf("Failed to place order: %v", err)
    }
    
    log.Printf("Order placed: %s", report.OrderId)
}
```

## Architecture

CQVX follows a clean architecture pattern with clear separation between public API and internal implementation:

```
┌─────────────────────────────────────────────────────────────────┐
│                    Consuming Service (cqmd/cqex/cqpm)           │
│                              ↓ imports                          │
├─────────────────────────────────────────────────────────────────┤
│  pkg/client/          VenueClient Interface (9 methods)         │
│  pkg/venues/          Venue Implementations                     │
│    ├── coinbase/      Coinbase Exchange Client                  │
│    ├── prime/         Coinbase Prime Client                     │
│    ├── falconx/       FalconX RFQ Client                        │
│    └── fordefi/       Fordefi MPC Client                        │
├─────────────────────────────────────────────────────────────────┤
│  internal/auth/       Authentication (HMAC, JWT, Bearer, MPC)   │
│  internal/normalizer/ Response normalization (venue → CQC)      │
├─────────────────────────────────────────────────────────────────┤
│  CQI Infrastructure   HTTP, WebSocket, Logging, Metrics         │
└─────────────────────────────────────────────────────────────────┘
```

### VenueClient Interface

All venue clients implement the following 9 methods:

**Trading Operations**:
- `PlaceOrder(ctx, *Order) (*ExecutionReport, error)`
- `CancelOrder(ctx, orderID) (*OrderStatus, error)`
- `GetOrder(ctx, orderID) (*Order, error)`
- `GetOrders(ctx, filter) ([]*Order, error)`

**Account Operations**:
- `GetBalance(ctx) (*Balance, error)`

**Market Data Operations**:
- `GetOrderBook(ctx, symbol) (*OrderBook, error)`

**Streaming Operations**:
- `SubscribeOrderBook(ctx, symbol, handler) error`
- `SubscribeTrades(ctx, symbol, handler) error`

**Health Check**:
- `Health(ctx) error`

## Repository Structure

```
cqvx/
├── cmd/              # Application entrypoints (none for library)
├── pkg/              # Public API (importable by consumers)
│   ├── client/       # VenueClient interface and types
│   │   └── mock/     # Mock client for testing
│   ├── venues/       # Venue implementations
│   │   ├── coinbase/ # Coinbase Exchange
│   │   ├── prime/    # Coinbase Prime
│   │   ├── falconx/  # FalconX
│   │   └── fordefi/  # Fordefi
│   └── types/        # Common types and filters
├── internal/         # Private implementation
│   ├── auth/         # Authentication signers
│   └── normalizer/   # Response normalization
├── examples/         # Usage examples
│   ├── simple/       # Basic order placement
│   └── streaming/    # WebSocket streaming
├── test/             # Integration tests
├── docs/             # Documentation
│   ├── BRIEF.md      # Project requirements
│   ├── SPEC.md       # Technical specification
│   └── ROADMAP.md    # Implementation roadmap
└── Makefile          # Build automation
```

## Development

### Build & Test

```bash
# Download dependencies
make tidy

# Run all tests
make test

# Run tests with coverage
make test-coverage

# Lint code
make lint

# Format code
make fmt

# Build all packages
make build

# Run all checks (format, vet, test)
make check
```

### Testing with Mock Client

CQVX provides a fully-featured mock client for testing consuming services without requiring live venue connections:

```go
package myservice_test

import (
    "context"
    "testing"
    
    "github.com/Combine-Capital/cqvx/pkg/client/mock"
    venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
    "github.com/stretchr/testify/assert"
)

func TestMyService_PlaceOrder(t *testing.T) {
    // Create mock client
    m := &mock.Client{}
    
    // Configure expected behavior
    m.OnPlaceOrder = func(ctx context.Context, order *venuesv1.Order) (*venuesv1.ExecutionReport, error) {
        return mock.NewExecutionReportBuilder().
            WithOrderID("mock-order-123").
            WithSymbol(*order.VenueSymbol).
            Build(), nil
    }
    
    // Use mock in your service
    service := NewMyService(m)
    result, err := service.PlaceTrade(context.Background(), "BTC-USD", 1.0)
    
    assert.NoError(t, err)
    assert.Equal(t, 1, m.PlaceOrderCallCount())
    
    // Verify call arguments
    _, capturedOrder := m.PlaceOrderCall(0)
    assert.Equal(t, "BTC-USD", *capturedOrder.VenueSymbol)
}
```

**Mock Features:**
- Full `VenueClient` interface implementation
- Configurable method behaviors (`OnPlaceOrder`, `OnGetOrderBook`, etc.)
- Call tracking with argument capture
- Test data builders for all CQC types
- Thread-safe for concurrent testing
- Default behaviors for all methods

## Contributing

This is an internal Combine Capital library. For development guidelines, see [Copilot Instructions](.github/copilot-instructions.md).

### Key Conventions

- **Constructor Pattern**: Accept CQI clients as parameters; never instantiate infrastructure inside venue clients
- **Type Safety**: All returns must use CQC protocol buffer types
- **Error Handling**: Map venue errors to CQI error types with context
- **No Direct HTTP**: Use CQI primitives; apply auth as middleware
- **Normalization**: All venue responses must normalize to CQC types in `internal/normalizer`

## Related Projects

- [CQ Hub](https://github.com/Combine-Capital/cqhub) - Platform Documentation
- [CQC](https://github.com/Combine-Capital/cqc) - Protocol Buffer Contracts
- [CQI](https://github.com/Combine-Capital/cqi) - Infrastructure Primitives

## Documentation

- [Project Brief](docs/BRIEF.md) - Requirements and success criteria
- [Technical Specification](docs/SPEC.md) - Detailed architecture and design
- [Implementation Roadmap](docs/ROADMAP.md) - Development plan and progress

## License

Copyright © 2025 Combine Capital. All rights reserved.

Internal use only. Not for public distribution.
