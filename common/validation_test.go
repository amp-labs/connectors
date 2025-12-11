// nolint:revive,godoclint
package common

import (
	"errors"
	"testing"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
)

func TestReadParamsValidateParams(t *testing.T) { // nolint:funlen
	t.Parallel()

	now := time.Now()

	tests := []struct {
		name               string
		params             ReadParams
		withRequiredFields bool
		wantErr            error
	}{
		{
			name:               "Missing object name",
			params:             ReadParams{},
			withRequiredFields: false,
			wantErr:            ErrMissingObjects,
		},
		{
			name:               "Missing fields",
			params:             ReadParams{ObjectName: "test"},
			withRequiredFields: true,
			wantErr:            ErrMissingFields,
		},
		{
			name:               "Valid params without required fields",
			params:             ReadParams{ObjectName: "test"},
			withRequiredFields: false,
			wantErr:            nil,
		},
		{
			name: "Valid params with required fields",
			params: ReadParams{
				ObjectName: "test",
				Fields:     datautils.NewSet("id"),
				Since:      now.Add(-1 * time.Hour),
				Until:      now,
			},
			withRequiredFields: true,
			wantErr:            nil,
		},
		{
			name: "Bad order, since after until",
			params: ReadParams{
				ObjectName: "test",
				Fields:     datautils.NewSet("id"),
				Since:      now.Add(+1 * time.Hour),
				Until:      now,
			},
			withRequiredFields: true,
			wantErr:            ErrSinceUntilChronOrder,
		},
		{
			name: "Valid params since only",
			params: ReadParams{
				ObjectName: "test",
				Fields:     datautils.NewSet("id"),
				Since:      now,
			},
			withRequiredFields: true,
			wantErr:            nil,
		},
		{
			name: "Valid params until only",
			params: ReadParams{
				ObjectName: "test",
				Fields:     datautils.NewSet("id"),
				Since:      now,
			},
			withRequiredFields: true,
			wantErr:            nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.params.ValidateParams(tt.withRequiredFields)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
