package coinbase

import (
	"context"
	"encoding/json"
	"fmt"

	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/Combine-Capital/cqvx/internal/normalizer"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CoinbaseAccount represents a Coinbase account balance response.
// This matches the structure returned by the Coinbase Advanced Trade API.
//
// Reference: https://docs.cdp.coinbase.com/advanced-trade-api/
type CoinbaseAccount struct {
	UUID             string                 `json:"uuid"`
	Name             string                 `json:"name"`
	Currency         string                 `json:"currency"`
	AvailableBalance CoinbaseAccountBalance `json:"available_balance"`
	Default          bool                   `json:"default"`
	Active           bool                   `json:"active"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
	DeletedAt        string                 `json:"deleted_at"`
	Type             string                 `json:"type"`
	Ready            bool                   `json:"ready"`
	Hold             CoinbaseAccountBalance `json:"hold"`
}

// CoinbaseAccountBalance represents balance value and currency.
type CoinbaseAccountBalance struct {
	Value    string `json:"value"`
	Currency string `json:"currency"`
}

// CoinbaseAccountsResponse represents the response from list accounts endpoint.
type CoinbaseAccountsResponse struct {
	Accounts []CoinbaseAccount `json:"accounts"`
	HasNext  bool              `json:"has_next"`
	Cursor   string            `json:"cursor"`
	Size     int               `json:"size"`
}

// NormalizeBalance converts a Coinbase account JSON response to CQC VenueAsset protobuf(s).
//
// Coinbase returns balance per asset/currency. Since CQC Balance is represented as VenueAsset
// with balance information, we normalize each account to a VenueAsset.
//
// The function handles:
//   - Parsing JSON response (single account or accounts list)
//   - Converting available and held balances
//   - Parsing timestamps
//   - Mapping account status to trading enabled flags
//
// Returns a slice of VenueAsset protos (one per currency) or error if parsing fails.
func NormalizeBalance(ctx context.Context, raw []byte) ([]*venuesv1.VenueAsset, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty balance response")
	}

	// Try parsing as accounts list first
	var accountsResp CoinbaseAccountsResponse
	if err := json.Unmarshal(raw, &accountsResp); err == nil && len(accountsResp.Accounts) > 0 {
		// Successfully parsed as accounts list
		assets := make([]*venuesv1.VenueAsset, 0, len(accountsResp.Accounts))
		for _, account := range accountsResp.Accounts {
			if asset, err := normalizeAccount(account); err == nil {
				assets = append(assets, asset)
			}
		}
		return assets, nil
	}

	// Try parsing as single account
	var account CoinbaseAccount
	if err := json.Unmarshal(raw, &account); err != nil {
		return nil, fmt.Errorf("failed to parse coinbase account: %w", err)
	}

	asset, err := normalizeAccount(account)
	if err != nil {
		return nil, err
	}

	return []*venuesv1.VenueAsset{asset}, nil
}

// normalizeAccount converts a single Coinbase account to a CQC VenueAsset.
func normalizeAccount(account CoinbaseAccount) (*venuesv1.VenueAsset, error) {
	// Parse balances
	availableBalance := normalizer.ParseDecimalOrZero(account.AvailableBalance.Value)
	heldBalance := normalizer.ParseDecimalOrZero(account.Hold.Value)

	// Total balance is available + held
	_ = availableBalance + heldBalance // We'll store in metadata

	// Parse timestamps
	var createdAt *timestamppb.Timestamp
	if account.CreatedAt != "" {
		if ts, err := normalizer.ParseTimestamp(account.CreatedAt); err == nil {
			createdAt = ts
		}
	}

	// Determine trading/withdrawal status
	tradingEnabled := account.Active && account.Ready
	withdrawEnabled := account.Active && account.Ready

	// Build VenueAsset (representing this currency balance on Coinbase)
	venueId := "coinbase"
	asset := &venuesv1.VenueAsset{
		VenueId:          &venueId,
		VenueAssetSymbol: &account.Currency,
		TradingEnabled:   &tradingEnabled,
		WithdrawEnabled:  &withdrawEnabled,
		IsActive:         &account.Active,
		ListedAt:         createdAt,
	}

	// Add delisted timestamp if present
	if account.DeletedAt != "" {
		if ts, err := normalizer.ParseTimestamp(account.DeletedAt); err == nil {
			asset.DelistedAt = ts
		}
	}

	// Note: CQC VenueAsset doesn't have balance fields directly.
	// In a full implementation, you'd either:
	// 1. Use a separate Balance protobuf message if CQC has one
	// 2. Store balance in the Metadata field
	// 3. Use the account endpoint to fetch asset info separately from balances
	//
	// For this normalizer, we're focusing on the asset/currency information.
	// Balance would typically be queried via a separate Balance protobuf or stored elsewhere.

	return asset, nil
}

// NormalizeAccountBalance is a helper that extracts just balance numbers from an account response.
// This can be used when only balance values are needed, not full asset information.
func NormalizeAccountBalance(ctx context.Context, raw []byte) (map[string]float64, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("empty balance response")
	}

	balances := make(map[string]float64)

	// Try parsing as accounts list first
	var accountsResp CoinbaseAccountsResponse
	if err := json.Unmarshal(raw, &accountsResp); err == nil && len(accountsResp.Accounts) > 0 {
		for _, account := range accountsResp.Accounts {
			available := normalizer.ParseDecimalOrZero(account.AvailableBalance.Value)
			held := normalizer.ParseDecimalOrZero(account.Hold.Value)
			balances[account.Currency] = available + held
		}
		return balances, nil
	}

	// Try parsing as single account
	var account CoinbaseAccount
	if err := json.Unmarshal(raw, &account); err != nil {
		return nil, fmt.Errorf("failed to parse coinbase account: %w", err)
	}

	available := normalizer.ParseDecimalOrZero(account.AvailableBalance.Value)
	held := normalizer.ParseDecimalOrZero(account.Hold.Value)
	balances[account.Currency] = available + held

	return balances, nil
}
