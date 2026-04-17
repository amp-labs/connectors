package odoo

import (
	"testing"

	"github.com/spyzhov/ajson"
)

func TestSearchReadNextPageOffset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		jsonBody      string
		currentOffset int
		limit         int
		want          string
		wantErr       bool
	}{
		{
			name:          "empty array",
			jsonBody:      `[]`,
			currentOffset: 0,
			limit:         100,
			want:          "",
			wantErr:       false,
		},
		{
			name:          "fewer rows than limit is last page",
			jsonBody:      `[{"id": 1}]`,
			currentOffset: 0,
			limit:         100,
			want:          "",
			wantErr:       false,
		},
		{
			name:          "full page from offset zero",
			jsonBody:      `[{"id": 1}, {"id": 2}]`,
			currentOffset: 0,
			limit:         2,
			want:          "2",
			wantErr:       false,
		},
		{
			name:          "full page adds to non-zero offset",
			jsonBody:      `[{"id": 3}, {"id": 4}]`,
			currentOffset: 10,
			limit:         2,
			want:          "12",
			wantErr:       false,
		},
		{
			name:          "exactly limit rows but zero limit edge",
			jsonBody:      `[{"id": 1}]`,
			currentOffset: 0,
			limit:         0,
			want:          "",
			wantErr:       false,
		},
		{
			name:          "non-array body errors",
			jsonBody:      `{"error": "bad"}`,
			currentOffset: 0,
			limit:         10,
			want:          "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			node, err := ajson.Unmarshal([]byte(tt.jsonBody))
			if err != nil {
				t.Fatalf("unmarshal: %v", err)
			}

			got, err := searchReadNextPageOffset(tt.currentOffset, tt.limit, node)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}

			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}
