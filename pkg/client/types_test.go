package client_test

import (
	"testing"
	"time"

	venuesv1 "github.com/Combine-Capital/cqc/gen/go/cqc/venues/v1"
	"github.com/Combine-Capital/cqvx/pkg/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrderFilter_Validate(t *testing.T) {
	tests := []struct {
		name    string
		filter  client.OrderFilter
		wantErr bool
		errType error
	}{
		{
			name:    "valid empty filter",
			filter:  client.OrderFilter{},
			wantErr: false,
		},
		{
			name: "valid filter with all fields",
			filter: client.OrderFilter{
				Symbols:   []string{"BTC-USD", "ETH-USD"},
				Statuses:  []venuesv1.OrderStatus{venuesv1.OrderStatus_ORDER_STATUS_OPEN},
				StartTime: time.Now().Add(-24 * time.Hour),
				EndTime:   time.Now(),
				Limit:     100,
				Offset:    0,
			},
			wantErr: false,
		},
		{
			name: "invalid negative limit",
			filter: client.OrderFilter{
				Limit: -1,
			},
			wantErr: true,
			errType: client.ErrInvalidLimit,
		},
		{
			name: "invalid negative offset",
			filter: client.OrderFilter{
				Offset: -1,
			},
			wantErr: true,
			errType: client.ErrInvalidOffset,
		},
		{
			name: "invalid time range - end before start",
			filter: client.OrderFilter{
				StartTime: time.Now(),
				EndTime:   time.Now().Add(-1 * time.Hour),
			},
			wantErr: true,
			errType: client.ErrInvalidTimeRange,
		},
		{
			name: "valid time range",
			filter: client.OrderFilter{
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid with only start time",
			filter: client.OrderFilter{
				StartTime: time.Now().Add(-1 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "valid with only end time",
			filter: client.OrderFilter{
				EndTime: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.wantErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestOrderFilter_HasTimeRange(t *testing.T) {
	tests := []struct {
		name   string
		filter client.OrderFilter
		want   bool
	}{
		{
			name:   "no time range",
			filter: client.OrderFilter{},
			want:   false,
		},
		{
			name: "has start time only",
			filter: client.OrderFilter{
				StartTime: time.Now(),
			},
			want: true,
		},
		{
			name: "has end time only",
			filter: client.OrderFilter{
				EndTime: time.Now(),
			},
			want: true,
		},
		{
			name: "has both times",
			filter: client.OrderFilter{
				StartTime: time.Now().Add(-1 * time.Hour),
				EndTime:   time.Now(),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filter.HasTimeRange())
		})
	}
}

func TestOrderFilter_HasSymbolFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter client.OrderFilter
		want   bool
	}{
		{
			name:   "no symbols",
			filter: client.OrderFilter{},
			want:   false,
		},
		{
			name: "empty symbols slice",
			filter: client.OrderFilter{
				Symbols: []string{},
			},
			want: false,
		},
		{
			name: "has symbols",
			filter: client.OrderFilter{
				Symbols: []string{"BTC-USD"},
			},
			want: true,
		},
		{
			name: "has multiple symbols",
			filter: client.OrderFilter{
				Symbols: []string{"BTC-USD", "ETH-USD"},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filter.HasSymbolFilter())
		})
	}
}

func TestOrderFilter_HasStatusFilter(t *testing.T) {
	tests := []struct {
		name   string
		filter client.OrderFilter
		want   bool
	}{
		{
			name:   "no statuses",
			filter: client.OrderFilter{},
			want:   false,
		},
		{
			name: "empty statuses slice",
			filter: client.OrderFilter{
				Statuses: []venuesv1.OrderStatus{},
			},
			want: false,
		},
		{
			name: "has single status",
			filter: client.OrderFilter{
				Statuses: []venuesv1.OrderStatus{venuesv1.OrderStatus_ORDER_STATUS_OPEN},
			},
			want: true,
		},
		{
			name: "has multiple statuses",
			filter: client.OrderFilter{
				Statuses: []venuesv1.OrderStatus{
					venuesv1.OrderStatus_ORDER_STATUS_OPEN,
					venuesv1.OrderStatus_ORDER_STATUS_FILLED,
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.filter.HasStatusFilter())
		})
	}
}

func TestOrderFilter_EdgeCases(t *testing.T) {
	t.Run("zero limit is valid", func(t *testing.T) {
		filter := client.OrderFilter{Limit: 0}
		assert.NoError(t, filter.Validate())
	})

	t.Run("zero offset is valid", func(t *testing.T) {
		filter := client.OrderFilter{Offset: 0}
		assert.NoError(t, filter.Validate())
	})

	t.Run("large values are valid", func(t *testing.T) {
		filter := client.OrderFilter{
			Limit:  10000,
			Offset: 1000000,
		}
		assert.NoError(t, filter.Validate())
	})

	t.Run("same start and end time is valid", func(t *testing.T) {
		now := time.Now()
		filter := client.OrderFilter{
			StartTime: now,
			EndTime:   now,
		}
		assert.NoError(t, filter.Validate())
	})
}
