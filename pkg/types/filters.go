// Package types provides common types and utilities used across the CQVX library.
package types

import (
	"errors"
	"time"
)

// TimeRange represents a time interval for filtering data.
// Both Start and End are optional. A zero value means no bound in that direction.
type TimeRange struct {
	// Start is the beginning of the time range (inclusive).
	// If zero, there is no lower bound.
	Start time.Time

	// End is the end of the time range (exclusive).
	// If zero, there is no upper bound.
	End time.Time
}

// ErrInvalidTimeRange is returned when a time range has an end before its start.
var ErrInvalidTimeRange = errors.New("time range end must be after start")

// Validate checks if the time range is valid.
// Returns an error if End is before Start (when both are non-zero).
func (tr *TimeRange) Validate() error {
	if !tr.Start.IsZero() && !tr.End.IsZero() && tr.Start.After(tr.End) {
		return ErrInvalidTimeRange
	}
	return nil
}

// IsZero returns true if both Start and End are zero (no time range specified).
func (tr *TimeRange) IsZero() bool {
	return tr.Start.IsZero() && tr.End.IsZero()
}

// HasStart returns true if a start time is specified.
func (tr *TimeRange) HasStart() bool {
	return !tr.Start.IsZero()
}

// HasEnd returns true if an end time is specified.
func (tr *TimeRange) HasEnd() bool {
	return !tr.End.IsZero()
}

// Duration returns the duration of the time range.
// Returns zero if either Start or End is not specified.
func (tr *TimeRange) Duration() time.Duration {
	if tr.Start.IsZero() || tr.End.IsZero() {
		return 0
	}
	return tr.End.Sub(tr.Start)
}

// Contains returns true if the given time falls within this range.
// If Start is zero, any time before End is considered in range.
// If End is zero, any time after Start is considered in range.
func (tr *TimeRange) Contains(t time.Time) bool {
	if !tr.HasStart() && !tr.HasEnd() {
		return true
	}
	if tr.HasStart() && t.Before(tr.Start) {
		return false
	}
	if tr.HasEnd() && !t.Before(tr.End) {
		return false
	}
	return true
}

// SymbolFilter provides filtering capabilities for trading symbols.
type SymbolFilter struct {
	// Symbols is a list of trading pair symbols to filter by (e.g., "BTC-USD", "ETH-USD").
	// If empty, all symbols are included.
	Symbols []string

	// Base filters by base currency (e.g., "BTC", "ETH").
	// If empty, no base currency filtering is applied.
	Base string

	// Quote filters by quote currency (e.g., "USD", "USDT").
	// If empty, no quote currency filtering is applied.
	Quote string
}

// IsEmpty returns true if no filters are specified.
func (sf *SymbolFilter) IsEmpty() bool {
	return len(sf.Symbols) == 0 && sf.Base == "" && sf.Quote == ""
}

// Matches returns true if the given symbol matches the filter criteria.
func (sf *SymbolFilter) Matches(symbol string) bool {
	if sf.IsEmpty() {
		return true
	}

	// Check explicit symbol list
	if len(sf.Symbols) > 0 {
		for _, s := range sf.Symbols {
			if s == symbol {
				return true
			}
		}
		return false
	}

	// For base/quote filtering, venue implementations would need to parse symbols
	// This is a simple check that can be extended by venue-specific implementations
	return true
}

// PaginationParams defines common pagination parameters.
type PaginationParams struct {
	// Limit is the maximum number of items to return.
	// Zero means use the venue's default limit.
	Limit int

	// Offset is the number of items to skip.
	// Zero means start from the beginning.
	Offset int

	// Cursor is an opaque pagination token for cursor-based pagination.
	// Some venues use cursor-based pagination instead of offset-based.
	Cursor string
}

// Validate checks if pagination parameters are valid.
func (pp *PaginationParams) Validate() error {
	if pp.Limit < 0 {
		return errors.New("limit must be non-negative")
	}
	if pp.Offset < 0 {
		return errors.New("offset must be non-negative")
	}
	return nil
}

// HasLimit returns true if a limit is specified.
func (pp *PaginationParams) HasLimit() bool {
	return pp.Limit > 0
}

// HasOffset returns true if an offset is specified.
func (pp *PaginationParams) HasOffset() bool {
	return pp.Offset > 0
}

// HasCursor returns true if a cursor is specified.
func (pp *PaginationParams) HasCursor() bool {
	return pp.Cursor != ""
}
