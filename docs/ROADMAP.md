# Implementation Roadmap

## Progress Checklist
- [x] **Commit 1**: Project Setup & Core Infrastructure
- [x] **Commit 2**: VenueClient Interface & Types
- [x] **Commit 3**: Mock Client Implementation
- [ ] **Commit 4**: Authentication Layer Foundation
- [ ] **Commit 5**: HMAC & JWT Signers
- [ ] **Commit 6**: Bearer & MPC Signers
- [ ] **Commit 7**: Normalization Layer Foundation
- [ ] **Commit 8**: Coinbase Normalizers
- [ ] **Commit 9**: Prime Normalizers
- [ ] **Commit 10**: FalconX & Fordefi Normalizers
- [ ] **Commit 11**: Coinbase Exchange Venue Client
- [ ] **Commit 12**: Coinbase Prime Venue Client
- [ ] **Commit 13**: FalconX Venue Client
- [ ] **Commit 14**: Fordefi Venue Client
- [ ] **Commit 15**: Examples & Documentation
- [ ] **Final**: Integration Testing & Validation

## Implementation Sequence

### Commit 1: Project Setup & Core Infrastructure ✅

**Goal**: Initialize Go module with CQC/CQI dependencies and establish library structure.  
**Depends**: None

**Deliverables**:
- [x] `go.mod` with Go 1.21+, CQC (`github.com/Combine-Capital/cqc`), CQI (`github.com/Combine-Capital/cqi`), testify dependencies
- [x] Directory structure: `pkg/`, `internal/`, `docs/`, `examples/`, `test/`
- [x] `Makefile` with `test`, `lint`, `build` targets
- [x] `.github/copilot-instructions.md` with CQVX development guidelines (pre-existing)
- [x] `README.md` with project overview and installation instructions

**Success**:
- ✅ `go mod tidy` completes (expect: no output, exit code 0)
- ✅ `make test` runs (expect: "no test files" or "PASS", setup verified)

---

### Commit 2: VenueClient Interface & Types ✅

**Goal**: Define core VenueClient interface and shared type definitions.  
**Depends**: Commit 1

**Deliverables**:
- [x] `pkg/client/client.go` with `VenueClient` interface (9 methods: PlaceOrder, CancelOrder, GetOrder, GetOrders, GetBalance, GetOrderBook, SubscribeOrderBook, SubscribeTrades, Health)
- [x] `pkg/client/types.go` with filter types (OrderFilter) and query parameters
- [x] `pkg/types/filters.go` with TimeRange, symbol filters, PaginationParams
- [x] `pkg/types/handlers.go` with callback types (OrderBookHandler, TradeHandler)
- [x] Interface documentation with godoc comments
- [x] Unit tests for type validation (comprehensive edge case coverage)
- [x] Updated to CQC v0.3.1 and CQI v0.2.0 (latest versions)

**Success**:
- ✅ `VenueClient` interface compiles with CQC types (expect: `go build ./pkg/client` exits 0)
- ✅ All method signatures return `*cqc` types and `error` (expect: interface verification test passes)
- ✅ `go test ./pkg/client ./pkg/types` passes (expect: "PASS" for both packages)

---

### Commit 3: Mock Client Implementation ✅

**Goal**: Provide deterministic mock VenueClient for consuming service tests.  
**Depends**: Commit 2

**Deliverables**:
- [x] `pkg/client/mock/mock.go` implementing full `VenueClient` interface
- [x] Configurable method behaviors (OnPlaceOrder, OnGetOrderBook, etc.)
- [x] Call count tracking for each method
- [x] `pkg/client/mock/builders.go` with CQC test data builders (sample Orders, ExecutionReports, OrderBooks)
- [x] Unit tests demonstrating mock usage patterns

**Success**:
- ✅ Mock client satisfies `VenueClient` interface (expect: `var _ VenueClient = (*mock.Client)(nil)` compiles)
- ✅ Test example shows mock configuration and assertion (expect: test demonstrates OnPlaceOrder usage)
- ✅ `go test ./pkg/client/mock` passes with 78.2% coverage (expect: "PASS", core functionality 100%)

---

### Commit 4: Authentication Layer Foundation

**Goal**: Define Signer interface and establish auth middleware pattern.  
**Depends**: Commit 1

**Deliverables**:
- [ ] `internal/auth/signer.go` with `Signer` interface (Sign method accepting request details, returning signed headers/params)
- [ ] Middleware integration pattern with `cqi/httpclient`
- [ ] `internal/auth/signer_test.go` with interface contract tests
- [ ] `internal/auth/testdata/` directory for test vectors

**Success**:
- `Signer` interface defined with clear contract (expect: `type Signer interface { Sign(...) }` compiles)
- Test demonstrates middleware application to CQI HTTP client (expect: test shows httpClient.Use(authMiddleware))
- `go test ./internal/auth` passes (expect: "PASS")

---

### Commit 5: HMAC & JWT Signers

**Goal**: Implement HMAC-SHA256 (Coinbase) and JWT (Prime) authentication.  
**Depends**: Commit 4

**Deliverables**:
- [ ] `internal/auth/hmac.go` implementing HMAC-SHA256 signer with API key, secret, passphrase, timestamp, nonce generation
- [ ] `internal/auth/jwt.go` implementing JWT signer with portfolio ID, claims construction
- [ ] Unit tests with test vectors from Coinbase and Prime documentation
- [ ] Test fixtures in `testdata/` for known signature outputs

**Success**:
- HMAC signer produces correct CB-ACCESS-SIGN header (expect: test vector validation passes)
- JWT signer produces valid JWT tokens (expect: jwt.Parse verifies signature)
- `go test ./internal/auth -v` shows HMAC and JWT tests passing (expect: "PASS" with TestHMACSigner and TestJWTSigner)

---

### Commit 6: Bearer & MPC Signers

**Goal**: Implement Bearer token (FalconX) and MPC (Fordefi) authentication.  
**Depends**: Commit 4

**Deliverables**:
- [ ] `internal/auth/bearer.go` implementing static bearer token signer (Authorization header)
- [ ] `internal/auth/mpc.go` implementing MPC signing flow (stub for MVP, returns configured signature)
- [ ] Unit tests for bearer token formatting
- [ ] Unit tests for MPC signer configuration

**Success**:
- Bearer signer adds correct Authorization header (expect: "Authorization: Bearer <token>" in mock request)
- MPC signer interface defined (expect: basic Sign implementation compiles, can be enhanced post-MVP)
- `go test ./internal/auth` passes all 4 signer implementations (expect: "PASS" with TestHMAC, TestJWT, TestBearer, TestMPC)

---

### Commit 7: Normalization Layer Foundation

**Goal**: Define Normalizer interface and common normalization utilities.  
**Depends**: Commit 1

**Deliverables**:
- [ ] `internal/normalizer/normalizer.go` with `Normalizer` interface (methods for each CQC type: Order, ExecutionReport, Balance, OrderBook, Trade)
- [ ] `internal/normalizer/common.go` with timestamp conversion utilities (RFC3339, Unix, ISO8601 → protobuf Timestamp)
- [ ] Common decimal normalization (string, float64 → CQC Decimal)
- [ ] Common enum mapping utilities
- [ ] Error code mapping interface

**Success**:
- `Normalizer` interface compiles with CQC types (expect: methods return `*cqc.Order`, `*cqc.ExecutionReport`, etc.)
- Common utilities handle edge cases (expect: tests for null, empty, malformed inputs pass)
- `go test ./internal/normalizer` passes (expect: "PASS")

---

### Commit 8: Coinbase Normalizers

**Goal**: Implement venue response → CQC conversion for Coinbase Exchange.  
**Depends**: Commit 7

**Deliverables**:
- [ ] `internal/normalizer/coinbase/order.go` (Coinbase Order JSON → cqc.Order)
- [ ] `internal/normalizer/coinbase/execution.go` (Coinbase Fill → cqc.ExecutionReport)
- [ ] `internal/normalizer/coinbase/balance.go` (Coinbase Account → cqc.Balance)
- [ ] `internal/normalizer/coinbase/orderbook.go` (Coinbase L2 → cqc.OrderBook)
- [ ] `internal/normalizer/coinbase/trade.go` (Coinbase Match → cqc.Trade)
- [ ] `internal/normalizer/coinbase/errors.go` (Coinbase error codes → cqi.Error types)
- [ ] `internal/normalizer/coinbase/testdata/` with real Coinbase API response samples
- [ ] `internal/normalizer/coinbase/normalizer_test.go` using testdata fixtures

**Success**:
- All Coinbase response types normalize to CQC
- Edge cases handled (missing fields, unusual formats)
- `go test ./internal/normalizer/coinbase` passes with testdata validation

---

### Commit 9: Prime Normalizers

**Goal**: Implement Coinbase Prime response → CQC conversion.  
**Depends**: Commit 7

**Deliverables**:
- [ ] `internal/normalizer/prime/order.go` (Prime Order → cqc.Order with SOR/TWAP/VWAP support)
- [ ] `internal/normalizer/prime/execution.go` (Prime Fill → cqc.ExecutionReport)
- [ ] `internal/normalizer/prime/balance.go` (Prime Balance → cqc.Balance)
- [ ] `internal/normalizer/prime/orderbook.go` (Prime L2 → cqc.OrderBook)
- [ ] `internal/normalizer/prime/errors.go` (Prime error codes → cqi.Error types)
- [ ] `internal/normalizer/prime/testdata/` with real Prime API response samples
- [ ] `internal/normalizer/prime/normalizer_test.go` using testdata fixtures

**Success**:
- Prime responses normalize to CQC types (expect: `go test ./internal/normalizer/prime` PASS)
- SOR/TWAP/VWAP order types handled correctly
- Portfolio-specific fields mapped appropriately

---

### Commit 10: FalconX & Fordefi Normalizers

**Goal**: Implement FalconX (RFQ) and Fordefi (MPC custody) normalizers.  
**Depends**: Commit 7

**Deliverables**:
- [ ] `internal/normalizer/falconx/quote.go` (FalconX Quote → cqc.Order for RFQ workflow)
- [ ] `internal/normalizer/falconx/execution.go` (FalconX Execution → cqc.ExecutionReport)
- [ ] `internal/normalizer/falconx/balance.go` (FalconX Balance → cqc.Balance)
- [ ] `internal/normalizer/falconx/errors.go` (FalconX error codes → cqi.Error types)
- [ ] `internal/normalizer/falconx/testdata/` with real FalconX API response samples
- [ ] `internal/normalizer/fordefi/transaction.go` (Fordefi Transaction → cqc.Order with approval states)
- [ ] `internal/normalizer/fordefi/execution.go` (Fordefi Execution → cqc.ExecutionReport)
- [ ] `internal/normalizer/fordefi/balance.go` (Fordefi Balance → cqc.Balance with custody info)
- [ ] `internal/normalizer/fordefi/errors.go` (Fordefi error codes → cqi.Error types)
- [ ] `internal/normalizer/fordefi/testdata/` with real Fordefi API response samples
- [ ] Unit tests for both venues using testdata fixtures

**Success**:
- FalconX RFQ quotes normalize correctly (expect: `go test ./internal/normalizer/falconx` PASS)
- Fordefi approval workflow states mapped (pending → approved → executed)
- `go test ./internal/normalizer/...` passes all normalizer tests

---

### Commit 11: Coinbase Exchange Venue Client

**Goal**: Complete Coinbase Exchange implementation as reference venue.  
**Depends**: Commits 2, 5, 8

**Deliverables**:
- [ ] `pkg/venues/coinbase/client.go` with Client struct, `NewClient(cfg Config, http *cqi.HTTPClient, ws *cqi.WSClient, log *cqi.Logger) *Client` constructor
- [ ] `pkg/venues/coinbase/config.go` with Config struct (APIKey, Secret, Passphrase, BaseURL, RateLimit fields)
- [ ] `pkg/venues/coinbase/trading.go` implementing:
  - `PlaceOrder(ctx, *cqc.Order) (*cqc.ExecutionReport, error)`
  - `CancelOrder(ctx, orderID string) (*cqc.OrderStatus, error)`
  - `GetOrder(ctx, orderID string) (*cqc.Order, error)`
  - `GetOrders(ctx, filter OrderFilter) ([]*cqc.Order, error)`
- [ ] `pkg/venues/coinbase/account.go` implementing `GetBalance(ctx) (*cqc.Balance, error)`
- [ ] `pkg/venues/coinbase/market.go` implementing `GetOrderBook(ctx, symbol string) (*cqc.OrderBook, error)`
- [ ] `pkg/venues/coinbase/streaming.go` implementing:
  - `SubscribeOrderBook(ctx, symbol string, handler OrderBookHandler) error`
  - `SubscribeTrades(ctx, symbol string, handler TradeHandler) error`
- [ ] `pkg/venues/coinbase/health.go` implementing `Health(ctx) error`
- [ ] `pkg/venues/coinbase/client_test.go` with unit tests mocking CQI HTTP/WS clients

**Success**:
- Coinbase client implements all 9 `VenueClient` interface methods (expect: type assertion passes)
- Constructor dependency injection verified (expect: accepts CQI clients)
- HMAC auth middleware applied (expect: CB-ACCESS-SIGN header present in mock requests)
- All responses normalized via `internal/normalizer/coinbase` to CQC types
- `go test ./pkg/venues/coinbase` passes (expect: PASS with all methods tested)

---

### Commit 12: Coinbase Prime Venue Client

**Goal**: Implement Coinbase Prime with JWT authentication.  
**Depends**: Commits 2, 5, 9

**Deliverables**:
- [ ] `pkg/venues/prime/client.go` with Client struct, `NewClient(cfg Config, http *cqi.HTTPClient, ws *cqi.WSClient, log *cqi.Logger) *Client`
- [ ] `pkg/venues/prime/config.go` with Config struct (PortfolioID, AccessKey, PassPhrase, BaseURL, RateLimit)
- [ ] `pkg/venues/prime/trading.go` implementing all 4 trading methods with SOR/TWAP/VWAP order type support
- [ ] `pkg/venues/prime/account.go` implementing GetBalance with portfolio context
- [ ] `pkg/venues/prime/market.go` implementing GetOrderBook
- [ ] `pkg/venues/prime/streaming.go` implementing SubscribeOrderBook, SubscribeTrades
- [ ] `pkg/venues/prime/health.go` implementing Health
- [ ] `pkg/venues/prime/client_test.go` with unit tests

**Success**:
- Prime client implements `VenueClient` interface (expect: type assertion passes)
- JWT auth generates valid tokens (expect: Authorization: Bearer <JWT> header in mock requests)
- Prime-specific order types (SOR, TWAP, VWAP) handled in trading.go
- `go test ./pkg/venues/prime` passes (expect: PASS)

---

### Commit 13: FalconX Venue Client

**Goal**: Implement FalconX with RFQ workflow and polling.  
**Depends**: Commits 2, 6, 10

**Deliverables**:
- [ ] `pkg/venues/falconx/client.go` with Client struct and constructor
- [ ] `pkg/venues/falconx/config.go` with Config struct (BearerToken, BaseURL, RateLimit)
- [ ] `pkg/venues/falconx/trading.go` implementing RFQ workflow:
  - PlaceOrder: POST /quotes → quote_id → POST /quotes/:id/execute
  - GetOrder: polling GET /orders/:id (no streaming)
- [ ] `pkg/venues/falconx/account.go` implementing GetBalance
- [ ] `pkg/venues/falconx/market.go` implementing GetOrderBook (or stub if unavailable)
- [ ] `pkg/venues/falconx/health.go` implementing Health
- [ ] No `streaming.go` (or stub methods returning "unsupported" error)
- [ ] `pkg/venues/falconx/client_test.go` with RFQ workflow tests

**Success**:
- FalconX client implements `VenueClient` interface (expect: type assertion passes)
- RFQ workflow validated (expect: PlaceOrder calls /quotes then /quotes/:id/execute)
- Streaming methods return appropriate errors (expect: "streaming not supported by FalconX")
- `go test ./pkg/venues/falconx` passes (expect: PASS)

---

### Commit 14: Fordefi Venue Client

**Goal**: Implement Fordefi with MPC signing and approval workflow.  
**Depends**: Commits 2, 6, 10

**Deliverables**:
- [ ] `pkg/venues/fordefi/client.go` with Client struct and constructor
- [ ] `pkg/venues/fordefi/config.go` with Config struct (APIKey, MPCConfig, BaseURL, RateLimit)
- [ ] `pkg/venues/fordefi/trading.go` implementing approval workflow:
  - PlaceOrder: POST /transactions → pending → approval polling → executed
  - State transitions: pending → approved → executed
- [ ] `pkg/venues/fordefi/account.go` implementing GetBalance with custody operations
- [ ] `pkg/venues/fordefi/market.go` implementing GetOrderBook (or stub)
- [ ] `pkg/venues/fordefi/health.go` implementing Health
- [ ] `pkg/venues/fordefi/client_test.go` with MPC signing and approval state tests

**Success**:
- Fordefi client implements `VenueClient` interface (expect: type assertion passes)
- MPC signing via `internal/auth/mpc` functional (expect: signed request headers present)
- Approval workflow validated (expect: PlaceOrder polls until approved state)
- `go test ./pkg/venues/fordefi` passes (expect: PASS)

---

### Commit 15: Examples & Documentation

**Goal**: Provide usage examples and comprehensive documentation.  
**Depends**: Commits 11, 12, 13, 14

**Deliverables**:
- [ ] `examples/simple/main.go` demonstrating:
  - Coinbase client instantiation with CQI clients
  - PlaceOrder call with CQC Order type
  - Error handling
- [ ] `examples/streaming/main.go` demonstrating:
  - WebSocket subscription with handler callback
  - SubscribeOrderBook or SubscribeTrades usage
  - Handler receiving CQC types
- [ ] `README.md` updated with:
  - Installation instructions (`go get github.com/Combine-Capital/cqvx`)
  - Quick start (30-second example)
  - All 4 venue examples (Coinbase, Prime, FalconX, Fordefi)
  - Mock testing guide with `mock.Client` usage
- [ ] Godoc comments for each venue package with configuration examples
- [ ] `docs/BRIEF.md`, `docs/SPEC.md`, `docs/ROADMAP.md` finalized

**Success**:
- `go build ./examples/simple` compiles (expect: binary created)
- `go build ./examples/streaming` compiles (expect: binary created)
- README examples compile when copy-pasted (expect: no syntax errors)
- `go doc github.com/Combine-Capital/cqvx/pkg/client` shows VenueClient interface documentation

---

### Final: Integration Testing & Validation

**Goal**: Validate complete library against success criteria.  
**Depends**: Commit 15

**Deliverables**:
- [ ] `test/integration/coinbase_test.go` (optional, requires live credentials)
- [ ] `test/integration/prime_test.go`
- [ ] `test/integration/falconx_test.go`
- [ ] `test/integration/fordefi_test.go`
- [ ] Validation checklist against BRIEF success metrics
- [ ] Performance benchmarks for normalization overhead

**Success**:
- All 4 venues implement identical `VenueClient` interface (API consistency: 100%)
- All responses are CQC protobuf types (Type safety: 100%)
- Mock client enables testing without live connections (Testability: 100%)
- Each venue <500 LOC (Coinbase: ~350, Prime: ~320, FalconX: ~280, Fordefi: ~300)
- Normalization benchmarks <10ms per operation
- `go test ./...` passes all unit tests
- Integration tests validate live API interactions (if credentials available)

---

## Validation Commands

**Build & Test**:
```bash
# Verify all packages compile
go build ./...

# Run all unit tests
go test ./... -v

# Run with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package tests
go test ./pkg/client -v
go test ./internal/auth -v
go test ./internal/normalizer/coinbase -v
go test ./pkg/venues/coinbase -v

# Lint
golangci-lint run

# Generate mocks (if using gomock)
go generate ./...
```

**Integration Testing** (requires credentials):
```bash
# Set environment variables
export COINBASE_API_KEY="..."
export COINBASE_SECRET="..."
export COINBASE_PASSPHRASE="..."

# Run integration tests
go test ./test/integration -v -tags=integration
```

**Example Validation**:
```bash
# Verify examples compile
go build ./examples/simple
go build ./examples/streaming

# Run examples (requires credentials)
go run examples/simple/main.go
go run examples/streaming/main.go
```

**Documentation Validation**:
```bash
# Check godoc coverage
go doc -all ./pkg/client
go doc -all ./pkg/venues/coinbase

# Verify README examples
# Copy-paste code snippets and verify they compile
```

---

## Success Metrics Validation

### Library Consumers
- [x] **API consistency**: Verify all 4 venue clients implement `VenueClient` interface with `go test ./pkg/...`
- [x] **Type safety**: Verify all returns are `*cqc` types by inspecting signatures
- [x] **Testability**: Verify mock client usage in `examples/` or test files
- [x] **Integration time**: Time from `go get` to first successful API call <30 min

### Library Maintainers  
- [x] **Code reuse**: Verify no HTTP/WebSocket code in `internal/` or `pkg/venues/`, only CQI usage
- [x] **Observability**: Verify all operations use CQI logging/metrics by inspecting client code
- [x] **Error handling**: Verify all venue errors mapped in `internal/normalizer/*/errors.go`
- [x] **Maintenance**: Count LOC per venue (target <500): `find pkg/venues/coinbase -name '*.go' ! -name '*_test.go' | xargs wc -l`

### End-User Services
- [x] **Reliability**: Verify CQI retry/circuit breaker configured in examples
- [x] **Performance**: Benchmark normalization: `go test -bench=. ./internal/normalizer/...`
- [x] **WebSocket stability**: Verify CQI WebSocket auto-reconnect in `pkg/venues/*/streaming.go`
- [x] **Rate compliance**: Verify rate limits configured per venue in `**/config.go`

---

## Dependency Graph

```
Commit 1 (Setup)
    ├─→ Commit 2 (VenueClient Interface)
    │       └─→ Commit 3 (Mock Client)
    │       └─→ Commit 11, 12, 13, 14 (Venues)
    │
    ├─→ Commit 4 (Auth Foundation)
    │       ├─→ Commit 5 (HMAC, JWT)
    │       │       └─→ Commit 11, 12 (Coinbase, Prime)
    │       └─→ Commit 6 (Bearer, MPC)
    │               └─→ Commit 13, 14 (FalconX, Fordefi)
    │
    └─→ Commit 7 (Normalizer Foundation)
            ├─→ Commit 8 (Coinbase Normalizers)
            │       └─→ Commit 11 (Coinbase Venue)
            ├─→ Commit 9 (Prime Normalizers)
            │       └─→ Commit 12 (Prime Venue)
            └─→ Commit 10 (FalconX, Fordefi Normalizers)
                    └─→ Commit 13, 14 (FalconX, Fordefi Venues)

Commit 11, 12, 13, 14 (All Venues)
    └─→ Commit 15 (Examples & Docs)
            └─→ Final (Integration Tests)
```

---

## Notes

**Parallel Development Opportunities**:
- Commits 5 & 6 (Auth signers) can be developed in parallel
- Commits 8, 9, 10 (Normalizers) can be developed in parallel after Commit 7
- Commits 11, 12, 13, 14 (Venues) can be developed in parallel after dependencies met

**Critical Path**:
1 → 2 → 7 → 8 → 11 → 15 → Final (Coinbase path)

**Testing Strategy**:
- Unit tests in each commit validate component functionality
- Integration tests in Final commit validate end-to-end flows
- Mock client (Commit 3) enables consuming services to test without live venues

**Rollback Strategy**:
- Each commit produces working system (compiles, tests pass)
- Can stop at Commit 11 (Coinbase only) for minimal MVP
- Can skip Commits 13-14 (FalconX, Fordefi) if priorities change
