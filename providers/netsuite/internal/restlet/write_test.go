package restlet

import (
	"context"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
)

// TestParseWriteResponse_RecordId is a regression test for the bug where
// large integer recordIds (e.g. 3609578) were being rendered in scientific
// notation ("3.609578e+06") because writeResponseBody.RecordId was of type `any`
// and the default json.Unmarshal decoded JSON numbers into float64.
func TestParseWriteResponse_RecordId(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		body string
		want string
	}{
		{
			name: "large integer recordId is not scientific notation",
			body: `{"header":{"status":"SUCCESS"},"body":{"recordId":3609578,"type":"salesorder"}}`,
			want: "3609578",
		},
		{
			name: "small integer recordId",
			body: `{"header":{"status":"SUCCESS"},"body":{"recordId":42,"type":"customer"}}`,
			want: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := newJSONResponse(t, tt.body)

			adapter := &Adapter{}

			result, err := adapter.parseWriteResponse(context.Background(), common.WriteParams{}, nil, resp)
			if err != nil {
				t.Fatalf("parseWriteResponse() unexpected error: %v", err)
			}

			if result.RecordId != tt.want {
				t.Fatalf("RecordId = %q, want %q", result.RecordId, tt.want)
			}
		})
	}
}

func newJSONResponse(t *testing.T, body string) *common.JSONHTTPResponse {
	t.Helper()

	httpResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
	}

	resp, err := common.ParseJSONResponse(context.Background(), httpResp, []byte(body))
	if err != nil {
		t.Fatalf("ParseJSONResponse: %v", err)
	}

	return resp
}
