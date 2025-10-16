package auth_test

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Combine-Capital/cqvx/internal/auth"
)

// ExampleBearerSigner demonstrates how to use the Bearer token signer for FalconX.
func ExampleBearerSigner() {
	// Create a Bearer signer
	config := auth.BearerConfig{
		Token: "your-bearer-token-here",
	}
	signer, err := auth.NewBearerSigner(config)
	if err != nil {
		log.Fatalf("failed to create Bearer signer: %v", err)
	}

	// Create a sign request
	req := auth.SignRequest{
		Method: "GET",
		Path:   "/api/v1/orders",
		Body:   nil,
	}

	// Sign the request
	ctx := context.Background()
	result, err := signer.Sign(ctx, req)
	if err != nil {
		log.Fatalf("failed to sign request: %v", err)
	}

	// Use the result headers
	fmt.Printf("Authorization: %s\n", result.Headers["Authorization"])
	// Output: Authorization: Bearer your-bearer-token-here
}

// ExampleBearerSigner_withMiddleware demonstrates using Bearer auth with HTTP middleware.
func ExampleBearerSigner_withMiddleware() {
	// Create a Bearer signer
	config := auth.BearerConfig{
		Token: "your-bearer-token-here",
	}
	signer, err := auth.NewBearerSigner(config)
	if err != nil {
		log.Fatalf("failed to create Bearer signer: %v", err)
	}

	// Create an HTTP client with Bearer authentication middleware
	client := &http.Client{
		Transport: auth.Middleware(signer, http.DefaultTransport),
	}

	// All requests made with this client will automatically include Bearer auth
	resp, err := client.Get("https://api.example.com/orders")
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	fmt.Printf("Status: %d\n", resp.StatusCode)
}

// ExampleMPCSigner demonstrates how to use the MPC signer for Fordefi.
func ExampleMPCSigner() {
	// Define a custom signer function (or use DefaultMPCSignerFunc for testing)
	signerFunc := func(ctx context.Context, message []byte) (string, error) {
		// In production, this would integrate with an MPC service
		// For MVP, we use a simple hash
		return auth.DefaultMPCSignerFunc(ctx, message)
	}

	// Create an MPC signer
	config := auth.MPCConfig{
		APIKey:     "your-api-key",
		SignerFunc: signerFunc,
	}
	signer, err := auth.NewMPCSigner(config)
	if err != nil {
		log.Fatalf("failed to create MPC signer: %v", err)
	}

	// Create a sign request
	req := auth.SignRequest{
		Method: "POST",
		Path:   "/api/v1/transactions",
		Body:   []byte(`{"amount":"100","currency":"USD"}`),
	}

	// Sign the request
	ctx := context.Background()
	result, err := signer.Sign(ctx, req)
	if err != nil {
		log.Fatalf("failed to sign request: %v", err)
	}

	// Use the result headers
	fmt.Printf("X-API-KEY: %s\n", result.Headers["X-API-KEY"])
	fmt.Printf("X-SIGNATURE length: %d\n", len(result.Headers["X-SIGNATURE"]))
	fmt.Printf("X-TIMESTAMP present: %v\n", result.Headers["X-TIMESTAMP"] != "")
	// Output:
	// X-API-KEY: your-api-key
	// X-SIGNATURE length: 64
	// X-TIMESTAMP present: true
}

// ExampleMPCSigner_withCustomSigner demonstrates using MPC with a custom signing function.
func ExampleMPCSigner_withCustomSigner() {
	// Define a custom signer that integrates with your MPC provider
	signerFunc := func(ctx context.Context, message []byte) (string, error) {
		// Example: Call an external MPC service
		// mpcService := myMPCProvider.NewClient()
		// signature, err := mpcService.Sign(ctx, message)
		// return signature, err

		// For this example, we'll use the default
		return auth.DefaultMPCSignerFunc(ctx, message)
	}

	config := auth.MPCConfig{
		APIKey:     "your-api-key",
		SignerFunc: signerFunc,
	}
	signer, err := auth.NewMPCSigner(config)
	if err != nil {
		log.Fatalf("failed to create MPC signer: %v", err)
	}

	// Use with middleware
	client := &http.Client{
		Transport: auth.Middleware(signer, http.DefaultTransport),
	}

	// All requests will be signed via MPC
	_ = client // Use client for requests
}
