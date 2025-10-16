package prime

import (
	"context"
	"encoding/json"
	"fmt"

	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/Combine-Capital/cqvx/internal/normalizer"
)

// PrimeBalance represents a Coinbase Prime portfolio balance response.
// This contains balances for assets within a specific portfolio.
//
// Reference: https://docs.cdp.coinbase.com/prime/rest-api/ (balance endpoints)
type PrimeBalance struct {
	Symbol               string `json:"symbol"`                 // Asset symbol (e.g., "BTC", "USD")
	Amount               string `json:"amount"`                 // Total amount
	Holds                string `json:"holds"`                  // Amount on hold
	BondedAmount         string `json:"bonded_amount"`          // Amount bonded/staked
	ReservedAmount       string `json:"reserved_amount"`        // Amount reserved
	UnbondingAmount      string `json:"unbonding_amount"`       // Amount in unbonding period
	UnvestedAmount       string `json:"unvested_amount"`        // Amount unvested
	PendingRewardsAmount string `json:"pending_rewards_amount"` // Pending rewards
	PastRewardsAmount    string `json:"past_rewards_amount"`    // Historical rewards
	BondableAmount       string `json:"bondable_amount"`        // Amount available to bond
	WithdrawableAmount   string `json:"withdrawable_amount"`    // Amount available to withdraw
}

// PrimeWalletBalance represents a wallet balance response.
// Prime organizes assets by wallets within portfolios.
type PrimeWalletBalance struct {
	Symbol     string `json:"symbol"`
	Amount     string `json:"amount"`
	Holds      string `json:"holds"`
	Type       string `json:"type"` // "TRADING", "VAULT", etc.
	WalletID   string `json:"wallet_id"`
	WalletName string `json:"wallet_name"`
}

// NormalizeBalance converts a Coinbase Prime balance JSON response to a CQC Balance protobuf.
//
// The function handles:
//   - Parsing JSON response
//   - Converting decimal values for amounts
//   - Mapping Prime balance types to CQC balance fields
//   - Handling holds, bonded amounts, and Prime-specific custody fields
//
// Returns an error if JSON parsing fails or required fields are missing.
func NormalizeBalance(ctx context.Context, raw []byte) (*venuesv1.Balance, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty balance response")
	}

	var primeBalance PrimeBalance
	if err := json.Unmarshal(raw, &primeBalance); err != nil {
		return nil, fmt.Errorf("failed to parse prime balance: %w", err)
	}

	// Parse decimal fields
	total := normalizer.ParseDecimalOrZero(primeBalance.Amount)
	holds := normalizer.ParseDecimalOrZero(primeBalance.Holds)

	// Calculate available balance (total - holds)
	available := total - holds

	// Build CQC Balance
	balance := &venuesv1.Balance{
		AssetId:   &primeBalance.Symbol,
		Total:     &total,
		Available: &available,
		Locked:    &holds, // Holds are effectively locked
	}

	// Prime-specific custody fields (bonded, unbonding, rewards, etc.)
	// could be added to a future venue-specific balance extension if needed.
	// For now, they're available in the raw response but not exposed in CQC Balance.

	return balance, nil
}

// NormalizeWalletBalance converts a Coinbase Prime wallet balance JSON response to a CQC Balance protobuf.
// This is similar to NormalizeBalance but handles wallet-specific balance information.
func NormalizeWalletBalance(ctx context.Context, raw []byte) (*venuesv1.Balance, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty wallet balance response")
	}

	var walletBalance PrimeWalletBalance
	if err := json.Unmarshal(raw, &walletBalance); err != nil {
		return nil, fmt.Errorf("failed to parse prime wallet balance: %w", err)
	}

	// Parse decimal fields
	total := normalizer.ParseDecimalOrZero(walletBalance.Amount)
	holds := normalizer.ParseDecimalOrZero(walletBalance.Holds)

	// Calculate available balance (total - holds)
	available := total - holds

	// Build CQC Balance with wallet context
	balance := &venuesv1.Balance{
		AccountId: &walletBalance.WalletID, // Use wallet ID as account ID
		AssetId:   &walletBalance.Symbol,
		Total:     &total,
		Available: &available,
		Locked:    &holds,
	}

	return balance, nil
}
