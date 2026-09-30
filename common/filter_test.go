// nolint:revive,godoclint
package common

import (
	"errors"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/internal/datautils"
)

// testFilter is a typed raw filter of syntax "test": a non-blank string without " OR ".
type testFilter struct {
	Expression string
}

var errTestNoOr = errors.New("must not contain OR")

func (testFilter) RawFilterType() string { return "test" }

func (testFilter) FromFilter(filter any) (testFilter, error) {
	expression, err := StringFilterValue(filter)
	if err != nil {
		return testFilter{}, err
	}

	if strings.Contains(expression, " OR ") {
		return testFilter{}, errTestNoOr
	}

	return testFilter{Expression: expression}, nil
}

func TestRawFilterAs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		filter    RawFilter
		want      string
		wantCause error
	}{
		{name: "matching type with a string", filter: RawFilter{Type: "test", Filter: "A = 1"}, want: "A = 1"},
		{name: "other type", filter: RawFilter{Type: "soql", Filter: "A = 1"}},
		{name: "missing type", filter: RawFilter{Filter: "A = 1"}},
		{name: "type set without a filter", filter: RawFilter{Type: "test"}},
		{name: "non-string filter", filter: RawFilter{Type: "test", Filter: 42}, wantCause: errFilterNotString},
		{name: "blank filter", filter: RawFilter{Type: "test", Filter: "  "}, wantCause: errFilterEmpty},
		{
			name:      "rejected by the typed filter",
			filter:    RawFilter{Type: "test", Filter: "A = 1 OR B = 2"},
			wantCause: errTestNoOr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := RawFilterAs[testFilter](tt.filter)
			if tt.want != "" {
				if err != nil || got.Expression != tt.want {
					t.Fatalf("RawFilterAs() = %q, %v; want %q, nil", got.Expression, err, tt.want)
				}

				return
			}

			if !errors.Is(err, ErrInvalidRawFilter) {
				t.Fatalf("RawFilterAs() error = %v, want ErrInvalidRawFilter", err)
			}

			if tt.wantCause != nil && !errors.Is(err, tt.wantCause) {
				t.Fatalf("RawFilterAs() error = %v, want it to wrap %v", err, tt.wantCause)
			}
		})
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
