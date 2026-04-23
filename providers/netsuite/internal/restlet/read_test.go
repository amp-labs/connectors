package restlet

import (
	"testing"
)

// TestParseSearchResults_IdCoercion is a regression test for the bug where
// large integer _id values were being rendered in scientific notation
// ("3.609578e+06") because the parser decoded JSON numbers into float64.
// The live RESTlet emits _id as a JSON number.
func TestParseSearchResults_IdCoercion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		body   string
		wantId string
	}{
		{
			name:   "large integer _id is not scientific notation",
			body:   `{"header":{"status":"SUCCESS","hasMore":false,"totalResults":1},"body":[{"_id":3609578,"_type":"salesorder"}]}`,
			wantId: "3609578",
		},
		{
			name:   "string _id passes through",
			body:   `{"header":{"status":"SUCCESS","hasMore":false,"totalResults":1},"body":[{"_id":"abc-123","_type":"customrecord_x"}]}`,
			wantId: "abc-123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := newJSONResponse(t, tt.body)

			result, err := parseSearchResults(resp)
			if err != nil {
				t.Fatalf("parseSearchResults() unexpected error: %v", err)
			}

			if len(result.Data) != 1 {
				t.Fatalf("got %d rows, want 1", len(result.Data))
			}

			if result.Data[0].Id != tt.wantId {
				t.Fatalf("row.Id = %q, want %q", result.Data[0].Id, tt.wantId)
			}
		})
	}
}
