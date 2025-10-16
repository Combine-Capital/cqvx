# Prime API Test Data

This directory contains sample JSON responses from the Coinbase Prime API used for testing normalizers.

## Files

- `order_limit.json` - Sample LIMIT order response
- `order_twap.json` - Sample TWAP (Time-Weighted Average Price) algorithmic order response
- `fill.json` - Sample fill/execution report from a portfolio
- `balance.json` - Sample portfolio balance response showing custody fields
- `orderbook.json` - Sample Level 2 order book snapshot

## Purpose

These test fixtures are used by `normalizer_test.go` to verify:
- Correct parsing of Prime-specific JSON structures
- Proper mapping to CQC protobuf types
- Handling of institutional trading features (SOR, TWAP, VWAP, portfolio context)
- Edge cases and missing fields

## Source

The JSON structures are based on Coinbase Prime API documentation:
https://docs.cdp.coinbase.com/api-reference/prime-api/
