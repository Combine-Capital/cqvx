# MVP Technical Specification: CQVX - Crypto Quant Venue Exchange

**Project Type:** Go Client Library  
**Target Consumers:** CQ platform services (cqmd, cqex, cqpm, cqdefi)  
**Deployment Model:** Imported as Go module dependency

## Core Requirements (from Brief)

### MVP Scope
1. **Library Architecture**
   - Public API in `pkg/`, implementation in `internal/`
   - No `cmd/` directory (library only)
   - Dependency injection: accept `cqi.HTTPClient` and `cqi.WebSocketClient` in constructors
   - Provide `mock.Client` for deterministic testing

2. **Core Interfaces**
   - `VenueClient` interface with 9 methods (trading, account, market data, streaming, health)
   - All methods return CQC-typed data or CQI-typed errors
   - All streaming uses CQI WebSocket client

3. **CQI Infrastructure Integration**
   - `cqi/httpclient`: REST calls with retry, circuit breaker, rate limiting
   - `cqi/websocket`: Streaming with auto-reconnect, connection pooling
   - `cqi/logging`, `cqi/metrics`, `cqi/errors`: Observability and error handling
   - Per-venue rate limits via `cqi/httpclient.RateLimitConfig`

4. **MVP Venue Implementations** (4 venues)
   - Coinbase Advanced Trade: REST v3 + WebSocket, HMAC-SHA256, 10/15 req/sec, spot (implemented as `pkg/venues/coinbase/`)
   - Coinbase Prime: REST + JWT, institutional, SOR/TWAP/VWAP
   - FalconX: REST + Bearer, RFQ, polling only
   - Fordefi: REST + MPC, approval workflow, custody

5. **Internal Components**
   - `internal/auth/`: Per-venue `Signer` interface (HMAC-SHA256, JWT, Bearer, MPC for MVP venues)
   - `internal/normalizer/`: Venue JSON/XML → CQC protobuf, error code mapping

### Post-MVP (Explicitly Excluded from Initial Implementation)
- Coinbase Exchange (legacy v2 API, replaced by Advanced Trade v3)
- Binance, Kraken, OKX, Bybit CEX connectors
- Uniswap V3, Curve DEX connectors
- Aave lending protocol
- FIX Protocol adapter

## Technology Stack

### Language & Runtime
- **Go 1.21+**: Required by CQI, supports generics for type-safe clients
- **Go Modules**: Dependency management

### Required Dependencies
- **CQC** (`github.com/Combine-Capital/cqc`): Protocol Buffer types
  - `cqc/venues/v1`: Venue, Order, OrderStatus, Balance, ExecutionReport
  - `cqc/markets/v1`: Price, OrderBook, Trade, MarketData
- **CQI** (`github.com/Combine-Capital/cqi`): Infrastructure primitives
  - `cqi/httpclient`: HTTP client with retry, circuit breaker, rate limiting
  - `cqi/websocket`: WebSocket client with auto-reconnect
  - `cqi/logging`: Structured logging (zerolog)
  - `cqi/metrics`: Prometheus metrics
  - `cqi/errors`: Typed error handling
- **Standard Library**: `context`, `time`, `sync`, `encoding/json`

### Testing Dependencies
- **testify**: Assertions and mocking (`github.com/stretchr/testify`)
- **gomock** (optional): Mock generation for interfaces

### Justification for Minimal Stack
- No additional HTTP/WebSocket libraries (CQI provides)
- No additional observability libraries (CQI provides)
- No database or caching (library does not persist state)
- No custom rate limiting (CQI provides)

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Consuming Service (cqmd/cqex/cqpm)           │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Service Code (e.g., Market Data Aggregator)            │   │
│  └────────────┬─────────────────────────────────────────────┘   │
│               │ imports & instantiates                           │
└───────────────┼──────────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────────┐
│                         CQVX Library                             │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  pkg/client/                                             │   │
│  │  ┌────────────────────────────────────────────────────┐  │   │
│  │  │  VenueClient Interface                             │  │   │
│  │  │  - PlaceOrder() → *cqc.ExecutionReport             │  │   │
│  │  │  - GetOrderBook() → *cqc.OrderBook                 │  │   │
│  │  │  - SubscribeTrades() → streaming via cqi.WSClient  │  │   │
│  │  └────────────────────────────────────────────────────┘  │   │
│  └──────────────────┬───────────────────────────────────────┘   │
│                     │ implemented by                             │
│  ┌──────────────────┴───────────────────────────────────────┐   │
│  │  pkg/venues/                                             │   │
│  │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐     │   │
│  │  │  coinbase/   │ │  prime/      │ │  falconx/    │     │   │
│  │  │  Client      │ │  Client      │ │  Client      │ ... │   │
│  │  └──────┬───────┘ └──────┬───────┘ └──────┬───────┘     │   │
│  └─────────┼────────────────┼────────────────┼─────────────┘   │
│            │                │                │                   │
│            └────────────────┼────────────────┘                   │
│                             │ uses                               │
│  ┌──────────────────────────────────────────────────────────────┘   │
│  │  internal/                                                │   │
│  │  ┌─────────────────────┐  ┌──────────────────────────┐   │   │
│  │  │  auth/              │  │  normalizer/             │   │   │
│  │  │  - Signer interface │  │  - JSON → CQC protobuf   │   │   │
│  │  │  - HMAC-SHA256      │  │  - Error code mapping    │   │   │
│  │  │  - JWT              │  │  - Timestamp conversion  │   │   │
│  │  │  - Bearer           │  │  - Decimal normalization │   │   │
│  │  │  - MPC              │  │                          │   │   │
│  │  └─────────────────────┘  └──────────────────────────┘   │   │
│  └──────────────────────────────────────────────────────────┘   │
│                             │ depends on                         │
└─────────────────────────────┼────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                    CQI Infrastructure                            │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────┐  │
│  │  httpclient      │  │  websocket       │  │  logging     │  │
│  │  - Retry         │  │  - Auto-reconnect│  │  - Zerolog   │  │
│  │  - Circuit break │  │  - Pooling       │  │              │  │
│  │  - Rate limit    │  │  - Multiplexing  │  │  metrics     │  │
│  └──────────────────┘  └──────────────────┘  │  - Prometheus│  │
│                                               │              │  │
│  ┌──────────────────────────────────────┐    │  errors      │  │
│  │  CQC Protocol Buffers                │    │  - Typed     │  │
│  │  - venues/v1: Order, ExecutionReport │    │  - Wrapping  │  │
│  │  - markets/v1: OrderBook, Trade      │    └──────────────┘  │
│  └──────────────────────────────────────┘                       │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
                    ┌──────────────────┐
                    │  Trading Venues  │
                    │  (Coinbase, etc) │
                    └──────────────────┘
```

## Data Flow

### End-to-End Flow: Service Places Order via CQVX

```
1. Service instantiates venue client:
   client := coinbase.NewClient(config, cqiHTTPClient, cqiWSClient, logger)

2. Service calls VenueClient method:
   report, err := client.PlaceOrder(ctx, order)
   
3. CQVX Coinbase Client processes:
   a. internal/auth: Generate HMAC-SHA256 signature
   b. Apply signature as middleware to cqiHTTPClient
   c. cqiHTTPClient: Make HTTP POST to Coinbase API
      - Automatic retry on transient failures
      - Rate limiting (15 req/sec enforced)
      - Metrics emitted (request count, latency)
   d. Receive Coinbase JSON response
   e. internal/normalizer: Parse JSON → CQC ExecutionReport protobuf
   f. Return *cqc.ExecutionReport to service

4. Service receives normalized CQC data:
   - No knowledge of Coinbase-specific formats
   - Same interface for all venues
   - CQI-typed errors on failure
```

### WebSocket Streaming Flow: Subscribe to Order Book

```
1. Service subscribes to order book:
   err := client.SubscribeOrderBook(ctx, "BTC-USD", handler)

2. CQVX Coinbase Client processes:
   a. internal/auth: Generate WebSocket auth token
   b. cqiWSClient: Establish WebSocket connection
      - Automatic reconnect on disconnect
      - Connection pooling
      - Metrics emitted (message rate, connection status)
   c. Send Coinbase subscribe message
   d. Receive Coinbase JSON messages
   e. internal/normalizer: Parse JSON → CQC OrderBook protobuf
   f. Call handler(orderBook *cqc.OrderBook)

3. Service handler receives normalized CQC data:
   - Real-time updates in CQC format
   - Automatic reconnection handled by CQI
   - No venue-specific message handling
```

## System Components

### Component: pkg/client (Core Interface)
**Purpose:** Define unified VenueClient interface that all venues implement

**Inputs:** None (interface definition)

**Outputs:** Go interface type

**Dependencies:** 
- `cqc/venues/v1` (Order, ExecutionReport, OrderStatus, Balance types)
- `cqc/markets/v1` (OrderBook, Trade, Price types)

**Key Responsibilities:**
- Define 9 methods covering trading, account, market data, streaming, health
- Ensure all methods use context.Context for cancellation
- Specify CQC types for all returns
- Document expected error types (CQI errors)

**Files:**
- `pkg/client/client.go`: Interface definition
- `pkg/client/types.go`: Filter and query parameter types

---

### Component: pkg/venues/{venue} (Venue Implementations)
**Purpose:** Implement VenueClient interface for specific trading venues

**Inputs:**
- Configuration (API keys, endpoints, rate limits)
- CQI HTTPClient (injected dependency)
- CQI WebSocketClient (injected dependency)
- Logger (injected dependency)

**Outputs:** 
- Venue-specific Client struct implementing VenueClient
- Constructor function: `NewClient(cfg Config, http *cqi.HTTPClient, ws *cqi.WSClient, logger *cqi.Logger) *Client`

**Dependencies:**
- `internal/auth`: Venue-specific signature generation
- `internal/normalizer`: Venue response → CQC conversion
- `cqi/httpclient`: HTTP requests
- `cqi/websocket`: Streaming connections
- `cqi/logging`, `cqi/metrics`, `cqi/errors`: Observability

**Key Responsibilities:**
- Implement all 9 VenueClient methods
- Configure CQI clients with venue-specific settings (rate limits, endpoints)
- Apply venue authentication via internal/auth
- Normalize all responses via internal/normalizer
- Handle venue-specific quirks (pagination, timestamp formats)
- Provide Health() check for connection status

**MVP Venues:**
- `pkg/venues/coinbase/`: Coinbase Advanced Trade (REST v3 + WebSocket)
- `pkg/venues/prime/`: Coinbase Prime (REST + JWT)
- `pkg/venues/falconx/`: FalconX (REST + RFQ)
- `pkg/venues/fordefi/`: Fordefi (REST + MPC)

**Files per venue:**
- `client.go`: Client struct and constructor
- `config.go`: Configuration struct
- `trading.go`: PlaceOrder, CancelOrder, GetOrder, GetOrders
- `account.go`: GetBalance
- `market.go`: GetOrderBook
- `streaming.go`: SubscribeOrderBook, SubscribeTrades
- `health.go`: Health check

---

### Component: pkg/client/mock (Mock Client)
**Purpose:** Provide deterministic mock for testing consuming services

**Inputs:** Test configuration (expected calls, canned responses)

**Outputs:** Mock implementation of VenueClient

**Dependencies:** 
- `cqc` types for canned responses
- `testify/mock` (optional) or manual mock

**Key Responsibilities:**
- Implement all 9 VenueClient methods
- Allow configuration of expected calls and responses
- Support assertion of method calls in tests
- No network communication (pure in-memory)

**Files:**
- `pkg/client/mock/mock.go`: Mock client implementation
- `pkg/client/mock/builders.go`: Helper functions for creating test data

---

### Component: internal/auth (Authentication)
**Purpose:** Generate venue-specific authentication signatures and tokens

**Inputs:**
- Venue credentials (API key, secret, passphrase, etc.)
- HTTP request details (method, path, body, timestamp)

**Outputs:**
- Signed request (headers or query parameters)
- Middleware function for CQI HTTPClient

**Dependencies:**
- Go standard library: `crypto/hmac`, `crypto/sha256`, `encoding/base64`
- JWT libraries for JWT-based auth

**Key Responsibilities:**
- Define `Signer` interface
- Implement venue-specific signers:
  - `HMACSignerSHA256`: Coinbase Advanced Trade (API key + secret + passphrase)
  - `JWTSigner`: Coinbase Prime (JWT-based signing)
  - `BearerTokenSigner`: FalconX (static bearer token)
  - `MPCSigner`: Fordefi (multi-party computation signing)
- Apply signatures as CQI HTTP middleware
- Handle nonce generation for replay protection
- Handle timestamp formatting per venue requirements

**Files:**
- `internal/auth/signer.go`: Signer interface
- `internal/auth/hmac.go`: HMAC-SHA256 implementation
- `internal/auth/jwt.go`: JWT implementation
- `internal/auth/bearer.go`: Bearer token implementation
- `internal/auth/mpc.go`: MPC signing implementation

---

### Component: internal/normalizer (Response Normalization)
**Purpose:** Convert venue-specific responses to CQC protobuf types

**Inputs:**
- Venue JSON/XML response bytes
- Venue error responses

**Outputs:**
- CQC protobuf messages
- CQI-typed errors

**Dependencies:**
- `encoding/json`: JSON parsing
- `cqc` protobuf types
- `cqi/errors`: Error type construction

**Key Responsibilities:**
- Parse venue-specific JSON schemas
- Map venue fields to CQC protobuf fields
- Handle venue-specific enums → CQC enums
- Convert timestamp formats (RFC3339, Unix, ISO8601) → protobuf Timestamp
- Normalize decimal formats (string, float64) → protobuf Decimal
- Map venue error codes to CQI error types (Permanent, Temporary, RateLimit, etc.)
- Handle missing fields with appropriate defaults

**Files:**
- `internal/normalizer/normalizer.go`: Normalizer interface
- `internal/normalizer/coinbase/`: Coinbase Advanced Trade normalizers
  - `order.go`: Coinbase Order → cqc.Order
  - `execution.go`: Coinbase Fill → cqc.ExecutionReport
  - `balance.go`: Coinbase Account → cqc.Balance
  - `orderbook.go`: Coinbase L2 → cqc.OrderBook
  - `errors.go`: Coinbase error codes → cqi.Error
- `internal/normalizer/prime/`: Prime-specific normalizers
- `internal/normalizer/falconx/`: FalconX-specific normalizers
- `internal/normalizer/fordefi/`: Fordefi-specific normalizers

---

### Component: pkg/types (Common Types)
**Purpose:** Provide shared types for configuration and parameters

**Inputs:** None (type definitions)

**Outputs:** Go struct types

**Dependencies:** None

**Key Responsibilities:**
- Define filter types for queries (OrderFilter, TimeRange)
- Define common configuration structs
- Define callback handler types for streaming

**Files:**
- `pkg/types/filters.go`: Query filter types
- `pkg/types/handlers.go`: Streaming handler types

## File Structure

```
github.com/Combine-Capital/cqvx/
├── go.mod                           # Module definition
├── go.sum                           # Dependency checksums
├── README.md                        # Library usage documentation
├── LICENSE                          # License file
├── Makefile                         # Build and test automation
│
├── docs/
│   ├── BRIEF.md                     # Project brief (requirements)
│   ├── SPEC.md                      # This technical specification
│   └── ROADMAP.md                   # Implementation roadmap
│
├── pkg/                             # Public API (importable)
│   ├── client/
│   │   ├── client.go                # VenueClient interface (9 methods)
│   │   ├── types.go                 # Common types (Config, Filter)
│   │   └── mock/
│   │       ├── mock.go              # Mock VenueClient implementation
│   │       └── builders.go          # Test data builders
│   │
│   ├── types/
│   │   ├── filters.go               # Query filters (OrderFilter, TimeRange)
│   │   └── handlers.go              # Streaming handlers (OrderBookHandler, TradeHandler)
│   │
│   └── venues/                      # Venue implementations
│       ├── coinbase/                # Coinbase Advanced Trade (v3 API)
│       │   ├── client.go            # Client struct + NewClient()
│       │   ├── config.go            # Config struct
│       │   ├── trading.go           # PlaceOrder, CancelOrder, GetOrder, GetOrders
│       │   ├── account.go           # GetBalance
│       │   ├── market.go            # GetOrderBook
│       │   ├── streaming.go         # SubscribeOrderBook, SubscribeTrades
│       │   ├── health.go            # Health
│       │   └── client_test.go       # Unit tests
│       │
│       ├── prime/                   # Coinbase Prime
│       │   ├── client.go
│       │   ├── config.go
│       │   ├── trading.go
│       │   ├── account.go
│       │   ├── market.go
│       │   ├── streaming.go
│       │   ├── health.go
│       │   └── client_test.go
│       │
│       ├── falconx/                 # FalconX
│       │   ├── client.go
│       │   ├── config.go
│       │   ├── trading.go           # RFQ-specific implementation
│       │   ├── account.go
│       │   ├── market.go
│       │   ├── health.go
│       │   └── client_test.go
│       │
│       └── fordefi/                 # Fordefi
│           ├── client.go
│           ├── config.go
│           ├── trading.go           # MPC + approval workflow
│           ├── account.go
│           ├── market.go
│           ├── health.go
│           └── client_test.go
│
├── internal/                        # Private implementation (not importable)
│   ├── auth/
│   │   ├── signer.go                # Signer interface
│   │   ├── hmac.go                  # HMAC-SHA256 signer
│   │   ├── jwt.go                   # JWT signer
│   │   ├── bearer.go                # Bearer token signer
│   │   ├── mpc.go                   # MPC signer
│   │   ├── signer_test.go           # Auth tests
│   │   └── testdata/                # Test fixtures (sample signatures)
│   │
│   └── normalizer/
│       ├── normalizer.go            # Normalizer interface
│       ├── common.go                # Common normalization functions
│       │
│       ├── coinbase/
│       │   ├── order.go             # Coinbase Order → cqc.Order
│       │   ├── execution.go         # Coinbase Fill → cqc.ExecutionReport
│       │   ├── balance.go           # Coinbase Account → cqc.Balance
│       │   ├── orderbook.go         # Coinbase L2 → cqc.OrderBook
│       │   ├── trade.go             # Coinbase Trade → cqc.Trade
│       │   ├── errors.go            # Error code mapping
│       │   ├── normalizer_test.go   # Tests
│       │   └── testdata/            # Sample JSON responses
│       │
│       ├── prime/
│       │   ├── order.go
│       │   ├── execution.go
│       │   ├── balance.go
│       │   ├── orderbook.go
│       │   ├── errors.go
│       │   ├── normalizer_test.go
│       │   └── testdata/
│       │
│       ├── falconx/
│       │   ├── quote.go             # FalconX Quote → cqc.Order
│       │   ├── execution.go
│       │   ├── balance.go
│       │   ├── errors.go
│       │   ├── normalizer_test.go
│       │   └── testdata/
│       │
│       └── fordefi/
│           ├── transaction.go       # Fordefi Transaction → cqc.Order
│           ├── execution.go
│           ├── balance.go
│           ├── errors.go
│           ├── normalizer_test.go
│           └── testdata/
│
├── examples/                        # Usage examples
│   ├── simple/
│   │   └── main.go                  # Basic usage example
│   └── streaming/
│       └── main.go                  # WebSocket streaming example
│
└── test/
    └── integration/                 # Integration tests (optional, requires live creds)
        ├── coinbase_test.go
        ├── prime_test.go
        ├── falconx_test.go
        └── fordefi_test.go
```

## Integration Patterns

### MVP Usage Pattern: Service Integrates Venue Client

```go
// Step 1: Import venue package
import (
    "github.com/Combine-Capital/cqvx/pkg/venues/coinbase"
    "github.com/Combine-Capital/cqi/pkg/httpclient"
    "github.com/Combine-Capital/cqi/pkg/websocket"
    "github.com/Combine-Capital/cqi/pkg/logging"
    cqcVenues "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
)

// Step 2: Create CQI infrastructure clients (service-managed)
httpClient := httpclient.New(httpclient.Config{
    Timeout: 10 * time.Second,
    Retry: httpclient.RetryConfig{MaxAttempts: 3, Backoff: "exponential"},
})

wsClient := websocket.New(websocket.Config{
    ReconnectBackoff: websocket.ExponentialBackoff{
        Initial: 1 * time.Second,
        Max: 60 * time.Second,
    },
})

logger := logging.New("cqmd")

// Step 3: Configure venue client
cfg := coinbase.Config{
    APIKey:     os.Getenv("COINBASE_API_KEY"),
    Secret:     os.Getenv("COINBASE_SECRET"),
    Passphrase: os.Getenv("COINBASE_PASSPHRASE"),
    BaseURL:    "https://api.exchange.coinbase.com",
    RateLimit: httpclient.RateLimitConfig{
        RequestsPerSecond: 15,
    },
}

// Step 4: Instantiate venue client
client := coinbase.NewClient(cfg, httpClient, wsClient, logger)

// Step 5: Use unified VenueClient interface
ctx := context.Background()

// Trading
order := &cqcVenues.Order{
    Symbol: "BTC-USD",
    Side: cqcVenues.OrderSide_BUY,
    Type: cqcVenues.OrderType_LIMIT,
    Quantity: "0.1",
    Price: "50000.00",
}
report, err := client.PlaceOrder(ctx, order)
if err != nil {
    // CQI-typed error with venue context
    logger.Error("order failed", "error", err)
}

// Market Data
book, err := client.GetOrderBook(ctx, "BTC-USD")

// Streaming
handler := func(book *cqcMarkets.OrderBook) {
    logger.Info("orderbook update", "bids", len(book.Bids))
}
err = client.SubscribeOrderBook(ctx, "BTC-USD", handler)

// Health
if err := client.Health(ctx); err != nil {
    logger.Warn("venue unhealthy", "error", err)
}
```

### MVP Testing Pattern: Using Mock Client

```go
// test/service_test.go
import (
    "testing"
    "github.com/Combine-Capital/cqvx/pkg/client/mock"
    cqcVenues "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
)

func TestOrderExecution(t *testing.T) {
    // Create mock venue client
    mockClient := mock.NewClient()
    
    // Configure mock behavior
    mockClient.OnPlaceOrder(func(ctx context.Context, order *cqcVenues.Order) (*cqcVenues.ExecutionReport, error) {
        return &cqcVenues.ExecutionReport{
            OrderId: "mock-12345",
            Status: cqcVenues.OrderStatus_FILLED,
            FilledQuantity: order.Quantity,
            AveragePrice: order.Price,
        }, nil
    })
    
    // Test service logic with mock
    service := NewTradingService(mockClient)
    err := service.ExecuteTrade(ctx, "BTC-USD", "0.1")
    assert.NoError(t, err)
    
    // Assert mock was called correctly
    assert.Equal(t, 1, mockClient.PlaceOrderCallCount())
}
```

### Post-MVP Extensions (Without Building for Them Now)

**Additional Venues:**
- Same VenueClient interface, just add new `pkg/venues/{venue}/` directory
- Reuse `internal/auth/` and `internal/normalizer/` patterns
- <500 LOC per venue (auth + normalization)

**Advanced Order Types:**
- Extend `cqc.Order` protobuf with new fields
- Update normalizers to handle new fields
- No changes to CQVX architecture

**Multi-Venue Routing:**
- Consuming service (cqex) implements routing logic
- CQVX remains simple single-venue client library
- No changes to CQVX

**Connection Pooling:**
- CQI manages connection pooling (no CQVX changes)
- CQVX clients automatically benefit

**Advanced Streaming:**
- CQI manages WebSocket multiplexing (no CQVX changes)
- CQVX just provides venue-specific subscribe logic

## Implementation Phases

### Phase 1: Core Infrastructure (Week 1)
1. Setup `go.mod` with CQC and CQI dependencies
2. Define `VenueClient` interface in `pkg/client/client.go`
3. Define common types in `pkg/types/`
4. Implement mock client in `pkg/client/mock/`
5. Write integration tests for mock client

### Phase 2: Authentication Layer (Week 1)
1. Define `Signer` interface in `internal/auth/signer.go`
2. Implement HMAC-SHA256 signer (for Coinbase)
3. Implement JWT signer (for Prime)
4. Implement Bearer token signer (for FalconX)
5. Implement MPC signer (for Fordefi)
6. Write comprehensive auth unit tests

### Phase 3: Normalization Layer (Week 2)
1. Define `Normalizer` interface
2. Implement common normalization utilities (timestamps, decimals)
3. Implement Coinbase normalizers (order, execution, balance, orderbook, errors)
4. Implement Prime normalizers
5. Implement FalconX normalizers
6. Implement Fordefi normalizers
7. Write normalization unit tests with testdata

### Phase 4: Venue Implementations (Week 3-4)
1. **Coinbase Advanced Trade** (Week 3)
   - Implement Client struct and constructor
   - Implement trading methods (PlaceOrder, CancelOrder, GetOrder, GetOrders)
   - Implement account methods (GetBalance)
   - Implement market data methods (GetOrderBook)
   - Implement streaming methods (SubscribeOrderBook, SubscribeTrades)
   - Implement health check
   - Write unit tests
   
2. **Coinbase Prime** (Week 3)
   - Same structure as Coinbase Advanced Trade, JWT auth variant
   
3. **FalconX** (Week 4)
   - RFQ workflow implementation
   - Polling-based (no WebSocket)
   
4. **Fordefi** (Week 4)
   - MPC signing integration
   - Approval workflow handling

### Phase 5: Documentation & Examples (Week 5)
1. Write comprehensive README.md with usage examples
2. Create `examples/simple/` with basic usage
3. Create `examples/streaming/` with WebSocket usage
4. Document each venue's specific configuration
5. Integration tests (optional, requires live credentials)

## Testing Strategy

### Unit Tests (Required)
- **Auth tests**: Verify signature generation matches venue expectations
  - Use test vectors from venue documentation
  - Mock HTTP client to verify headers/params
- **Normalizer tests**: Verify JSON → CQC conversion
  - Use real venue response samples in testdata/
  - Test edge cases (missing fields, unusual formats)
- **Client tests**: Verify method implementations
  - Mock CQI HTTP/WS clients
  - Verify correct endpoints, methods, bodies
  - Verify error handling

### Integration Tests (Optional)
- Requires live venue credentials (not for CI)
- Test end-to-end flows with real venues
- Validate normalization with live data
- Used for manual verification, not automated testing

### Mock Testing (Required)
- Comprehensive mock client for consuming services
- Verify mock matches VenueClient interface exactly
- Test data builders for common scenarios

## Success Criteria

### Library Consumers
- ✅ Can import and instantiate any venue client in <30 minutes
- ✅ All operations return CQC types (100% type-safe)
- ✅ Can test service logic with mock.Client (no live connections)
- ✅ Same code works across all 4 MVP venues

### Library Maintainers
- ✅ Adding new venue requires <500 LOC
- ✅ Zero HTTP/WebSocket infrastructure duplication
- ✅ All operations emit CQI logs/metrics automatically
- ✅ All venue errors mapped to CQI types

### Runtime
- ✅ ≥99.9% request success rate (CQI retry + circuit breaker)
- ✅ <10ms normalization overhead
- ✅ 1-60s exponential backoff WebSocket reconnect
- ✅ 0% HTTP 429 errors (CQI rate limiting)

## Non-Goals (Explicitly Out of Scope)

- ❌ Running as standalone service
- ❌ Implementing HTTP/WebSocket infrastructure
- ❌ Cross-venue aggregation or routing logic
- ❌ Data persistence or caching
- ❌ Event publishing to message bus
- ❌ Service discovery or registration
- ❌ Admin UI or monitoring dashboards
- ❌ Building for Post-MVP venues (Binance, Kraken, etc.)
