package auth_test

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Combine-Capital/cqvx/internal/auth"
)

// ExampleSigner demonstrates implementing a custom signer.
func ExampleSigner() {
	// Define a simple API key signer
	type APIKeySigner struct {
		apiKey string
		secret string
	}

	// Implement the Sign method
	sign := func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
		// In a real implementation, you would generate a proper signature
		signature := fmt.Sprintf("signature-for-%s-%s", req.Method, req.Path)

		return &auth.SignResult{
			Headers: map[string]string{
				"X-API-Key":   "my-api-key",
				"X-Signature": signature,
			},
		}, nil
	}

	// Create a mock signer for demonstration
	signer := &mockSigner{signFunc: sign}

	// Create HTTP client with auth middleware
	client := &http.Client{
		Transport: auth.Middleware(signer, nil),
	}

	// Make authenticated request
	req, _ := http.NewRequest("GET", "https://api.example.com/orders", nil)
	resp, _ := client.Do(req)

	fmt.Println("Status:", resp.StatusCode)
	// Output will vary based on actual endpoint
}

// ExampleMiddleware demonstrates using the Middleware function.
func ExampleMiddleware() {
	// Create a signer that adds a bearer token
	signer := &mockSigner{
		signFunc: func(ctx context.Context, req auth.SignRequest) (*auth.SignResult, error) {
			return &auth.SignResult{
				Headers: map[string]string{
					"Authorization": "Bearer my-token-123",
				},
			}, nil
		},
	}

	// Wrap the default transport with auth middleware
	transport := auth.Middleware(signer, http.DefaultTransport)

	// Create HTTP client with authenticated transport
	client := &http.Client{
		Transport: transport,
	}

	// All requests through this client will be authenticated
	req, _ := http.NewRequest("GET", "https://api.example.com/data", nil)
	client.Do(req)
	// Request will include: Authorization: Bearer my-token-123
}

// ExampleSignRequest demonstrates the structure of a sign request.
func ExampleSignRequest() {
	req := auth.SignRequest{
		Method:    "POST",
		Path:      "/api/v1/orders",
		Body:      []byte(`{"symbol":"BTC-USD","side":"buy","quantity":"1.0"}`),
		Timestamp: "1634567890",
		Headers:   http.Header{"Content-Type": []string{"application/json"}},
	}

	// A signer would use these fields to generate a signature
	fmt.Println("Method:", req.Method)
	fmt.Println("Path:", req.Path)
	fmt.Println("Body length:", len(req.Body))
	// Output:
	// Method: POST
	// Path: /api/v1/orders
	// Body length: 50
}

// ExampleSignResult demonstrates the structure of a sign result.
func ExampleSignResult() {
	result := auth.SignResult{
		Headers: map[string]string{
			"X-API-Key":   "key123",
			"X-Signature": "abc123def456",
			"X-Timestamp": "1634567890",
		},
		QueryParams: map[string]string{
			"api_key": "key123",
		},
	}

	// The middleware will add these to the request
	fmt.Println("Headers count:", len(result.Headers))
	fmt.Println("Query params count:", len(result.QueryParams))
	// Output:
	// Headers count: 3
	// Query params count: 1
}

// ExampleHMACSigner demonstrates using HMAC-SHA256 authentication for Coinbase Exchange.
func ExampleHMACSigner() {
	// Create HMAC signer with Coinbase credentials
	config := auth.HMACConfig{
		APIKey:     "your-api-key",
		Secret:     "dGVzdC1zZWNyZXQ=", // base64-encoded secret
		Passphrase: "your-passphrase",
	}

	signer, err := auth.NewHMACSigner(config)
	if err != nil {
		fmt.Printf("Failed to create HMAC signer: %v\n", err)
		return
	}

	// Sign a request
	req := auth.SignRequest{
		Method: "GET",
		Path:   "/api/v3/brokerage/accounts",
		Body:   []byte(""),
	}

	result, err := signer.Sign(context.Background(), req)
	if err != nil {
		fmt.Printf("Failed to sign request: %v\n", err)
		return
	}

	// The result contains authentication headers
	fmt.Println("CB-ACCESS-KEY:", result.Headers["CB-ACCESS-KEY"])
	fmt.Println("CB-ACCESS-PASSPHRASE:", result.Headers["CB-ACCESS-PASSPHRASE"])
	fmt.Println("Has CB-ACCESS-SIGN:", len(result.Headers["CB-ACCESS-SIGN"]) > 0)
	fmt.Println("Has CB-ACCESS-TIMESTAMP:", len(result.Headers["CB-ACCESS-TIMESTAMP"]) > 0)
	// Output:
	// CB-ACCESS-KEY: your-api-key
	// CB-ACCESS-PASSPHRASE: your-passphrase
	// Has CB-ACCESS-SIGN: true
	// Has CB-ACCESS-TIMESTAMP: true
}

// ExampleJWTSigner demonstrates using JWT authentication for Coinbase Prime.
func ExampleJWTSigner() {
	// Example EC private key (in production, load from secure storage)
	privateKey := `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIIGlRHkQJvpBd8Ir+L5P5RLqQrZ2IQzD4vX6F6l7R6svoAoGCCqGSM49
AwEHoUQDQgAExampleKeyDataHere
-----END EC PRIVATE KEY-----`

	// Create JWT signer with Coinbase Prime credentials
	config := auth.JWTConfig{
		KeyName:    "organizations/test-org/apiKeys/test-key",
		PrivateKey: privateKey,
		ExpiresIn:  120, // 2 minutes
	}

	signer, err := auth.NewJWTSigner(config)
	if err != nil {
		// In this example, the key is invalid, so this will error
		fmt.Println("Note: Example uses invalid key for demonstration")
		return
	}

	// Sign a request
	req := auth.SignRequest{
		Method: "GET",
		Path:   "/api/v3/brokerage/accounts",
	}

	result, err := signer.Sign(context.Background(), req)
	if err != nil {
		fmt.Printf("Failed to sign request: %v\n", err)
		return
	}

	// The result contains Authorization header with Bearer token
	authHeader := result.Headers["Authorization"]
	fmt.Println("Has Authorization header:", len(authHeader) > 0)
	fmt.Println("Is Bearer token:", authHeader[:7] == "Bearer ")
	// Output (with valid key):
	// Has Authorization header: true
	// Is Bearer token: true
}
