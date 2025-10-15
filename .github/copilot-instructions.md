# Copilot Instructions for CQVX

## Context & Documentation

Always use Context7 for current docs on frameworks/libraries/APIs; invoke automatically without being asked.

## Development Standards

### Go Best Practices
- Constructor functions must accept CQI clients as parameters (`*cqi.HTTPClient`, `*cqi.WSClient`, `*cqi.Logger`); never instantiate infrastructure inside venue clients.
- Use Go 1.21+ generics for type-safe handler functions; avoid `interface{}` for streaming callbacks.

### CQI Integration Patterns
- Never implement HTTP/WebSocket logic; compose CQI primitives with venue-specific auth/normalization only.
- Apply auth signatures as CQI HTTP middleware using `httpclient.WithMiddleware()`; never sign requests manually in venue code.

### Protocol Buffers
- All venue responses must normalize to CQC types (`cqc/venues/v1`, `cqc/markets/v1`); never return venue-specific structs from VenueClient methods.
- Map venue error codes to CQI error types (`cqi/errors`) in normalizers; include venue error code in error context for debugging.

### Code Quality Standards
- Every venue client must implement all 9 VenueClient methods; return "unsupported operation" errors for venue-specific limitations (e.g., FalconX streaming).
- Mock client must match VenueClient interface exactly; use `testify/mock` for method call assertions in consuming service tests.

### Project Conventions
- Venue implementations live in `pkg/venues/{venue}/`; auth/normalization logic lives in `internal/` and is shared across venues.
- Each venue has 7 files: `client.go`, `config.go`, `trading.go`, `account.go`, `market.go`, `streaming.go`, `health.go`.

### Agentic AI Guidelines
- Never create "summary" documents; direct action is more valuable than summarization.
