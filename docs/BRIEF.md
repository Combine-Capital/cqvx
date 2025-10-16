# Project Brief: CQVX - Crypto Quant Venue Exchange

## Vision
Go client library providing normalized access to cryptocurrency trading venues (CEX, DEX, OTC) via CQC types, using CQI HTTP/WebSocket infrastructure for all network communication.

## Core Principle: Library, Not Service
Imported as a Go package by services (`cqmd`, `cqex`, `cqpm`, `cqdefi`). Does not run standalone—consuming services instantiate and manage venue clients.

## Platform Dependencies

- **CQC** (`github.com/Combine-Capital/cqc`): Protocol Buffer types for venues/v1 (Venue, Order, OrderStatus, Balance, ExecutionReport) and markets/v1 (Price, OrderBook, Trade, MarketData). All venue responses normalized to these types.

- **CQI** (`github.com/Combine-Capital/cqi`): HTTP client (`cqi/httpclient`), WebSocket client (`cqi/websocket`), retry logic, rate limiting, logging, metrics, tracing, errors. CQVX does not implement network infrastructure—composes CQI primitives with venue-specific authentication.

## User Personas

### Service Developer (Library Consumer)
**Goal:** Integrate trading venues into CQ services without implementing venue-specific protocols.

**Needs:** Uniform `VenueClient` interface, CQC-typed responses, mock clients for testing, dependency injection of CQI clients.

**Success:** Import venue package, instantiate client with CQI infrastructure, call methods receiving normalized data, test with mocks.

### CQVX Maintainer (Library Developer)
**Goal:** Add/maintain venue connectors focusing only on venue-specific logic.

**Needs:** Leverage CQI HTTP/WebSocket infrastructure, implement venue authentication/normalization, provide mock implementations.

**Success:** Write <500 LOC per venue (auth + normalization), zero infrastructure duplication, automatic observability inheritance from CQI.

## Core Requirements

### Library Architecture
- [MVP] Public API in `pkg/`, implementation in `internal/`
- [MVP] No `cmd/` directory (library only, not a service)
- [MVP] Accept `cqi.HTTPClient` and `cqi.WebSocketClient` in constructors
- [MVP] Provide `mock.Client` for deterministic testing

### Core Interfaces
- [MVP] `VenueClient` interface in `pkg/client/`:
  - `PlaceOrder(ctx, *cqc.Order) (*cqc.ExecutionReport, error)`
  - `CancelOrder(ctx, orderID) (*cqc.OrderStatus, error)`
  - `GetOrder(ctx, orderID) (*cqc.Order, error)`
  - `GetOrders(ctx, filter) ([]*cqc.Order, error)`
  - `GetBalance(ctx) (*cqc.Balance, error)`
  - `GetOrderBook(ctx, symbol) (*cqc.OrderBook, error)`
  - `SubscribeOrderBook(ctx, symbol, handler) error`
  - `SubscribeTrades(ctx, symbol, handler) error`
  - `Health(ctx) error`
- [MVP] All returns: CQC-typed data, CQI-typed errors
- [MVP] All streaming: CQI WebSocket client

### CQI Infrastructure Integration
- [MVP] `cqi/httpclient`: All REST calls (retry, circuit breaker, rate limiting, observability)
- [MVP] `cqi/websocket`: All streaming (auto-reconnect, connection pooling, observability)
- [MVP] `cqi/logging`: Structured logs with venue context
- [MVP] `cqi/metrics`: Request count, latency, error rates
- [MVP] `cqi/errors`: Typed error handling
- [MVP] Per-venue rate limits via `cqi/httpclient.RateLimitConfig`

### Venue Implementations (MVP)
- [MVP] **Coinbase Advanced Trade**: REST v3 + WebSocket, HMAC-SHA256 auth, 10/15 req/sec limits, spot trading (implemented as `pkg/venues/coinbase/`)
- [MVP] **Coinbase Prime**: REST + JWT auth, institutional trading, SOR/TWAP/VWAP orders
- [MVP] **FalconX**: REST + Bearer token, RFQ workflow, polling only (no WebSocket)
- [MVP] **Fordefi**: REST + MPC signing, approval workflow, custody operations

### Venue-Specific Logic (Internal)
- [MVP] `internal/auth/`: Per-venue `Signer` interface (HMAC-SHA256, ED25519, JWT, OAuth2) applied as CQI HTTP middleware
- [MVP] `internal/normalizer/`: Venue JSON/XML → CQC protobuf, handle enums/timestamps/decimals, map error codes to CQI types

### Post-MVP Venue Implementations
- [Post-MVP] **Coinbase Exchange** (legacy v2 API, replaced by Advanced Trade v3)
- [Post-MVP] **Binance**, **Kraken**, **OKX**, **Bybit**: REST + WebSocket CEX connectors
- [Post-MVP] **Uniswap V3**, **Curve**: DEX via Ethereum JSON-RPC
- [Post-MVP] **Aave**: Lending protocol
- [Post-MVP] **FIX Protocol**: Generic FIX 4.2/4.4 adapter

## Success Metrics

### Library Consumers (Service Developers)
1. **API consistency**: 100% venue clients implement identical `VenueClient` interface
2. **Type safety**: 100% responses are CQC protobuf messages
3. **Testability**: 100% operations testable via `mock.Client` without live connections
4. **Integration time**: <30 min to integrate new venue (import, configure, call)

### Library Maintainers (CQVX Developers)
1. **Code reuse**: 0% HTTP/WebSocket infrastructure duplication across venues
2. **Observability**: 100% operations emit CQI logs/metrics/traces automatically
3. **Error handling**: 100% venue errors mapped to CQI error types
4. **Maintenance**: <500 LOC per new venue (auth + normalization only)

### End-User Services
1. **Reliability**: ≥99.9% request success (CQI retry + circuit breaker)
2. **Performance**: <10ms normalization overhead (venue JSON → CQC protobuf)
3. **WebSocket stability**: 1-60s exponential backoff auto-reconnect
4. **Rate compliance**: 0% HTTP 429 errors (CQI rate limiting)

## Clear Boundaries

### ✅ CQVX DOES
- Venue authentication (signature algorithms, nonce generation, credential formatting)
- Venue API definitions (endpoints, paths, parameters)
- Response parsing (venue JSON/XML → CQC protobuf)
- Error mapping (venue codes → CQI error types)
- Venue quirks (pagination, timestamp formats, decimal precision)
- Venue configuration (API keys, endpoints, portfolio IDs)
- Mock client implementations

### ❌ CQVX DOES NOT
- HTTP/WebSocket infrastructure (connection pooling, retry, auto-reconnect) → **CQI**
- Rate limiting, circuit breakers, backoff strategies → **CQI**
- Logging, metrics, distributed tracing → **CQI**
- Run as standalone service (no cmd/, no health endpoints) → **Library only**
- Data storage, state management, caching → **Consuming services**
- Cross-venue aggregation, VWAP calculation, trading decisions → **cqmd, cqex, cqstrat**
- Event publishing, service lifecycle → **Consuming services**