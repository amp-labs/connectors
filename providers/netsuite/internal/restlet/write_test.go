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

// TestParseWriteResponse_Data ensures response body fields beyond recordId
// (e.g. linesAdded from returnLineDetails) flow through to WriteResult.Data so
// callers can access them.
func TestParseWriteResponse_Data(t *testing.T) {
	t.Parallel()

	body := `{"header":{"status":"SUCCESS"},"body":{"recordId":144727,"type":"salesOrder",` +
		`"linesAdded":{"item":[{"index":0,"lineId":"95859"},{"index":1,"lineId":"95860"}]}}}`

	resp := newJSONResponse(t, body)

	adapter := &Adapter{}

	result, err := adapter.parseWriteResponse(context.Background(), common.WriteParams{}, nil, resp)
	if err != nil {
		t.Fatalf("parseWriteResponse() unexpected error: %v", err)
	}

	if result.RecordId != "144727" {
		t.Fatalf("RecordId = %q, want %q", result.RecordId, "144727")
	}

	if result.Data == nil {
		t.Fatal("Data is nil, want map with linesAdded")
	}

	linesAdded, ok := result.Data["linesAdded"]
	if !ok {
		t.Fatalf("Data missing 'linesAdded'; got keys: %v", keys(result.Data))
	}

	m, ok := linesAdded.(map[string]any)
	if !ok {
		t.Fatalf("linesAdded = %T, want map[string]any", linesAdded)
	}

	items, ok := m["item"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("linesAdded.item = %v, want 2-element array", m["item"])
	}
}

// TestParseWriteResponse_Header ensures the RESTlet's response header reaches callers in
// WriteResult.Data, next to the body fields, so they can read metadata such as the version.
func TestParseWriteResponse_Header(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		body        string
		wantVersion any
		wantNoKey   bool
	}{
		{
			name:        "header with scriptVersion",
			body:        `{"header":{"status":"SUCCESS","scriptVersion":"1.2.1"},"body":{"recordId":"42","type":"invoice"}}`,
			wantVersion: "1.2.1",
		},
		{
			name:        "older header with version",
			body:        `{"header":{"status":"SUCCESS","version":"2.3.8"},"body":{"recordId":"42"}}`,
			wantVersion: nil,
		},
		{
			name:        "body without recordId, as for ping",
			body:        `{"header":{"status":"SUCCESS","scriptVersion":"1.2.1"},"body":{"message":"pong"}}`,
			wantVersion: "1.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := newJSONResponse(t, tt.body)

			result, err := (&Adapter{}).parseWriteResponse(context.Background(), common.WriteParams{}, nil, resp)
			if err != nil {
				t.Fatalf("parseWriteResponse() unexpected error: %v", err)
			}

			header, ok := result.Data[responseHeaderKey].(map[string]any)
			if !ok {
				t.Fatalf("Data[%q] = %T, want the response header", responseHeaderKey, result.Data[responseHeaderKey])
			}

			if header["status"] != "SUCCESS" {
				t.Errorf("header.status = %v, want SUCCESS", header["status"])
			}

			if got := header["scriptVersion"]; got != tt.wantVersion {
				t.Errorf("header.scriptVersion = %v, want %v", got, tt.wantVersion)
			}
		})
	}
}

func keys(m map[string]any) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}

	return ks
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
