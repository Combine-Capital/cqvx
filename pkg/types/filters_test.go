package types_test

import (
	"testing"
	"time"

	"github.com/Combine-Capital/cqvx/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimeRange_Validate(t *testing.T) {
	tests := []struct {
		name      string
		timeRange types.TimeRange
		wantErr   bool
	}{
		{
			name:      "zero time range is valid",
			timeRange: types.TimeRange{},
			wantErr:   false,
		},
		{
			name: "valid time range",
			timeRange: types.TimeRange{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "invalid time range - end before start",
			timeRange: types.TimeRange{
				Start: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: true,
		},
		{
			name: "only start time is valid",
			timeRange: types.TimeRange{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "only end time is valid",
			timeRange: types.TimeRange{
				End: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
		{
			name: "same start and end is valid",
			timeRange: types.TimeRange{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.timeRange.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, types.ErrInvalidTimeRange)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTimeRange_IsZero(t *testing.T) {
	tests := []struct {
		name      string
		timeRange types.TimeRange
		want      bool
	}{
		{
			name:      "zero time range",
			timeRange: types.TimeRange{},
			want:      true,
		},
		{
			name: "has start only",
			timeRange: types.TimeRange{
				Start: time.Now(),
			},
			want: false,
		},
		{
			name: "has end only",
			timeRange: types.TimeRange{
				End: time.Now(),
			},
			want: false,
		},
		{
			name: "has both",
			timeRange: types.TimeRange{
				Start: time.Now(),
				End:   time.Now(),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.timeRange.IsZero())
		})
	}
}

func TestTimeRange_Duration(t *testing.T) {
	tests := []struct {
		name      string
		timeRange types.TimeRange
		want      time.Duration
	}{
		{
			name:      "zero time range returns zero duration",
			timeRange: types.TimeRange{},
			want:      0,
		},
		{
			name: "only start returns zero",
			timeRange: types.TimeRange{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: 0,
		},
		{
			name: "only end returns zero",
			timeRange: types.TimeRange{
				End: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			want: 0,
		},
		{
			name: "24 hour duration",
			timeRange: types.TimeRange{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
			},
			want: 24 * time.Hour,
		},
		{
			name: "1 hour duration",
			timeRange: types.TimeRange{
				Start: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2024, 1, 1, 1, 0, 0, 0, time.UTC),
			},
			want: 1 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.timeRange.Duration())
		})
	}
}

func TestTimeRange_Contains(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		timeRange types.TimeRange
		testTime  time.Time
		want      bool
	}{
		{
			name:      "zero time range contains everything",
			timeRange: types.TimeRange{},
			testTime:  baseTime,
			want:      true,
		},
		{
			name: "time after start (no end)",
			timeRange: types.TimeRange{
				Start: baseTime.Add(-1 * time.Hour),
			},
			testTime: baseTime,
			want:     true,
		},
		{
			name: "time before start (no end)",
			timeRange: types.TimeRange{
				Start: baseTime.Add(1 * time.Hour),
			},
			testTime: baseTime,
			want:     false,
		},
		{
			name: "time before end (no start)",
			timeRange: types.TimeRange{
				End: baseTime.Add(1 * time.Hour),
			},
			testTime: baseTime,
			want:     true,
		},
		{
			name: "time after end (no start)",
			timeRange: types.TimeRange{
				End: baseTime.Add(-1 * time.Hour),
			},
			testTime: baseTime,
			want:     false,
		},
		{
			name: "time within range",
			timeRange: types.TimeRange{
				Start: baseTime.Add(-1 * time.Hour),
				End:   baseTime.Add(1 * time.Hour),
			},
			testTime: baseTime,
			want:     true,
		},
		{
			name: "time before range",
			timeRange: types.TimeRange{
				Start: baseTime.Add(1 * time.Hour),
				End:   baseTime.Add(2 * time.Hour),
			},
			testTime: baseTime,
			want:     false,
		},
		{
			name: "time after range",
			timeRange: types.TimeRange{
				Start: baseTime.Add(-2 * time.Hour),
				End:   baseTime.Add(-1 * time.Hour),
			},
			testTime: baseTime,
			want:     false,
		},
		{
			name: "time equals start (inclusive)",
			timeRange: types.TimeRange{
				Start: baseTime,
				End:   baseTime.Add(1 * time.Hour),
			},
			testTime: baseTime,
			want:     true,
		},
		{
			name: "time equals end (exclusive)",
			timeRange: types.TimeRange{
				Start: baseTime.Add(-1 * time.Hour),
				End:   baseTime,
			},
			testTime: baseTime,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.timeRange.Contains(tt.testTime))
		})
	}
}

func TestSymbolFilter_IsEmpty(t *testing.T) {
	tests := []struct {
		name   string
		filter types.SymbolFilter
		want   bool
	}{
		{
			name:   "empty filter",
			filter: types.SymbolFilter{},
			want:   true,
		},
		{
			name: "has symbols",
			filter: types.SymbolFilter{
				Symbols: []string{"BTC-USD"},
			},
			want: false,
		},
		{
			name: "has base",
			filter: types.SymbolFilter{
				Base: "BTC",
			},
			want: false,
		},
		{
			name: "has quote",
			filter: types.SymbolFilter{
				Quote: "USD",
			},
			want: false,
		},
		{
			name: "has all fields",
			filter: types.SymbolFilter{
				Symbols: []string{"BTC-USD"},
				Base:    "BTC",
				Quote:   "USD",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filter.IsEmpty())
		})
	}
}

func TestSymbolFilter_Matches(t *testing.T) {
	tests := []struct {
		name   string
		filter types.SymbolFilter
		symbol string
		want   bool
	}{
		{
			name:   "empty filter matches everything",
			filter: types.SymbolFilter{},
			symbol: "BTC-USD",
			want:   true,
		},
		{
			name: "symbol in list matches",
			filter: types.SymbolFilter{
				Symbols: []string{"BTC-USD", "ETH-USD"},
			},
			symbol: "BTC-USD",
			want:   true,
		},
		{
			name: "symbol not in list doesn't match",
			filter: types.SymbolFilter{
				Symbols: []string{"BTC-USD", "ETH-USD"},
			},
			symbol: "XRP-USD",
			want:   false,
		},
		{
			name: "single symbol matches",
			filter: types.SymbolFilter{
				Symbols: []string{"BTC-USD"},
			},
			symbol: "BTC-USD",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filter.Matches(tt.symbol))
		})
	}
}

func TestPaginationParams_Validate(t *testing.T) {
	tests := []struct {
		name    string
		params  types.PaginationParams
		wantErr bool
	}{
		{
			name:    "empty params are valid",
			params:  types.PaginationParams{},
			wantErr: false,
		},
		{
			name: "valid params",
			params: types.PaginationParams{
				Limit:  100,
				Offset: 50,
			},
			wantErr: false,
		},
		{
			name: "negative limit is invalid",
			params: types.PaginationParams{
				Limit: -1,
			},
			wantErr: true,
		},
		{
			name: "negative offset is invalid",
			params: types.PaginationParams{
				Offset: -1,
			},
			wantErr: true,
		},
		{
			name: "zero values are valid",
			params: types.PaginationParams{
				Limit:  0,
				Offset: 0,
			},
			wantErr: false,
		},
		{
			name: "cursor is valid",
			params: types.PaginationParams{
				Cursor: "next_page_token",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPaginationParams_HasLimit(t *testing.T) {
	tests := []struct {
		name   string
		params types.PaginationParams
		want   bool
	}{
		{
			name:   "no limit",
			params: types.PaginationParams{},
			want:   false,
		},
		{
			name: "zero limit",
			params: types.PaginationParams{
				Limit: 0,
			},
			want: false,
		},
		{
			name: "positive limit",
			params: types.PaginationParams{
				Limit: 100,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.params.HasLimit())
		})
	}
}

func TestPaginationParams_HasOffset(t *testing.T) {
	tests := []struct {
		name   string
		params types.PaginationParams
		want   bool
	}{
		{
			name:   "no offset",
			params: types.PaginationParams{},
			want:   false,
		},
		{
			name: "zero offset",
			params: types.PaginationParams{
				Offset: 0,
			},
			want: false,
		},
		{
			name: "positive offset",
			params: types.PaginationParams{
				Offset: 50,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.params.HasOffset())
		})
	}
}

func TestPaginationParams_HasCursor(t *testing.T) {
	tests := []struct {
		name   string
		params types.PaginationParams
		want   bool
	}{
		{
			name:   "no cursor",
			params: types.PaginationParams{},
			want:   false,
		},
		{
			name: "empty cursor",
			params: types.PaginationParams{
				Cursor: "",
			},
			want: false,
		},
		{
			name: "has cursor",
			params: types.PaginationParams{
				Cursor: "next_page",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.params.HasCursor())
		})
	}
}
