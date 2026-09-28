// nolint:revive,godoclint
package common

import (
	"errors"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/internal/datautils"
)

func TestRawFilterAsString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		filter  RawFilter
		wantErr bool
	}{
		{name: "expected type with a string", filter: RawFilter{Type: "soql", Filter: "IsDeleted = true"}},
		{name: "other type", filter: RawFilter{Type: "sql", Filter: "IsDeleted = true"}, wantErr: true},
		{name: "missing type", filter: RawFilter{Filter: "IsDeleted = true"}, wantErr: true},
		{name: "non-string filter", filter: RawFilter{Type: "soql", Filter: 42}, wantErr: true},
		{name: "type set without a filter", filter: RawFilter{Type: "soql"}, wantErr: true},
		{name: "blank filter", filter: RawFilter{Type: "soql", Filter: "  "}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := tt.filter.AsString("soql")
			if tt.wantErr != (err != nil) {
				t.Fatalf("AsString() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && !errors.Is(err, ErrInvalidRawFilter) {
				t.Fatalf("error %v does not wrap ErrInvalidRawFilter", err)
			}

			if err == nil && got != tt.filter.Filter {
				t.Fatalf("AsString() = %q, want %q", got, tt.filter.Filter)
			}
		})
	}
}

func TestRawFilterStringValidators(t *testing.T) {
	t.Parallel()

	errNoOr := errors.New("OR is not allowed")
	noOr := func(expression string) error {
		if strings.Contains(expression, " OR ") {
			return errNoOr
		}

		return nil
	}

	if err := (RawFilter{Type: "soql", Filter: "A = 1"}).ValidateString("soql", noOr); err != nil {
		t.Fatalf("ValidateString() error = %v, want nil", err)
	}

	err := RawFilter{Type: "soql", Filter: "A = 1 OR B = 2"}.ValidateString("soql", noOr)
	if !errors.Is(err, ErrInvalidRawFilter) || !errors.Is(err, errNoOr) {
		t.Fatalf("ValidateString() error = %v, want ErrInvalidRawFilter wrapping the validator's error", err)
	}
}

func TestSearchParamsNeedExactlyOneFilterKind(t *testing.T) {
	t.Parallel()

	fieldFilter := SearchFilter{}.FilterBy("Name", FilterOperatorEQ, "Paris")
	rawFilter := &RawFilter{Type: "soql", Filter: "Name = 'Paris'"}

	tests := []struct {
		name    string
		params  SearchParams
		wantErr error
	}{
		{name: "field filters", params: SearchParams{Filter: fieldFilter}},
		{name: "raw filter", params: SearchParams{RawFilter: rawFilter}},
		{name: "neither", params: SearchParams{}, wantErr: ErrMissingSearchFilters},
		{name: "both", params: SearchParams{Filter: fieldFilter, RawFilter: rawFilter}, wantErr: ErrSearchFiltersCombined},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.params.ObjectName = "Account"
			tt.params.Fields = datautils.NewSet("Name")

			err := tt.params.ValidateParams(true)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateParams() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
