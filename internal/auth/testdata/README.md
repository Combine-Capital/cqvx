# Test Data Directory

This directory contains test vectors and fixtures for authentication signer implementations.

## Structure

Test vectors will be organized by signer type:

- `hmac/` - HMAC-SHA256 test vectors from Coinbase Exchange API documentation
- `jwt/` - JWT test vectors from Coinbase Prime API documentation

## HMAC-SHA256 (Coinbase Exchange)

### Test Vector 1: Simple GET Request
- **Method**: GET
- **Path**: /orders
- **Body**: (empty)
- **Timestamp**: 1234567890
- **Secret**: "secret" (base64: `c2VjcmV0`)
- **Expected Signature**: `c0bz9rdYCiGfAsKzIyfvmtx6eU1fbWn3SVcwKIVqZM4=`

### Signature Computation
```
prehash = timestamp + method + path + body
        = "1234567890" + "GET" + "/orders" + ""
        = "1234567890GET/orders"

signature = base64(HMAC-SHA256(base64_decode(secret), prehash))
```

## JWT (Coinbase Prime)

### JWT Structure
- **Algorithm**: ES256 (ECDSA with P-256 curve)
- **Issuer**: "cdp"
- **Expiration**: 120 seconds from issuance
- **Key Format**: PEM-encoded EC private key

### Claims
- `iss`: "cdp"
- `nbf`: Current Unix timestamp
- `exp`: nbf + 120 seconds
- `sub`: API key name (e.g., "organizations/{org_id}/apiKeys/{key_id}")
- `uri`: "{METHOD} {HOST}{PATH}"

### Headers
- `alg`: "ES256"
- `typ`: "JWT"
- `kid`: API key name
- `nonce`: 32-character hex string (16 random bytes)

## Notes
- HMAC secret must be base64-decoded before signing
- JWT nonces must be unique for replay protection
- JWT tokens expire after 2 minutes by default
- All timestamps are Unix seconds (not milliseconds)
- `bearer/` - Bearer token test cases for FalconX
- `mpc/` - MPC signing test fixtures for Fordefi

## Test Vector Format

Test vectors should include:
- Input request details (method, path, body, timestamp)
- Expected signature output
- Credentials used (sanitized/example keys only)
- Source documentation reference

## Usage

Test vectors are used in unit tests to verify that signers produce correct
signatures according to venue API specifications. This ensures compatibility
with live venue APIs.

## Note

Never commit real API keys or secrets to this directory. Use only example
keys from official documentation.
