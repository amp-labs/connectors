package reply

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
)

// TestWriteCustomFieldTitlePayloadTypes demonstrates that JSON-equivalent
// payloads should all resolve a title-named custom field before being sent.
// This is a failing reproducer: only the map[string]any case currently passes.
func TestWriteCustomFieldTitlePayloadTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		payload any
	}{
		{
			name:    "map_string_any_control",
			payload: map[string]any{"Deal Stage": "won"},
		},
		{
			name:    "common_Record",
			payload: common.Record{"Deal Stage": "won"},
		},
		{
			name:    "map_string_string",
			payload: map[string]string{"Deal Stage": "won"},
		},
		{
			name: "struct_with_JSON_tag",
			payload: struct {
				DealStage string `json:"Deal Stage"`
			}{DealStage: "won"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// An in-memory transport supplies definitions without a live Reply
			// account, a listening test server, or any external network access.
			connector, err := NewConnector(common.ConnectorParams{
				AuthenticatedClient: &http.Client{
					Transport: customFieldTitleReproTransport(func(request *http.Request) (*http.Response, error) {
						if request.Method != http.MethodGet ||
							request.URL.Path != "/v3/custom-fields" ||
							request.URL.Query().Get("entityType") != "contact" {
							return nil, fmt.Errorf("unexpected request: %s %s", request.Method, request.URL)
						}

						return &http.Response{
							StatusCode: http.StatusOK,
							Header:     http.Header{"Content-Type": []string{"application/json"}},
							Body: io.NopCloser(strings.NewReader(
								`[{"id":150697,"title":"Deal Stage","fieldType":"text"}]`,
							)),
							Request: request,
						}, nil
					}),
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			request, err := connector.buildWriteRequest(context.Background(), common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "759458702",
				RecordData: tt.payload,
			})
			if err != nil {
				t.Fatal(err)
			}
			defer request.Body.Close()

			var got map[string]any
			if err := json.NewDecoder(request.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}

			want := map[string]any{
				"customFields": []any{
					map[string]any{"id": float64(150697), "value": "won"},
				},
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("custom field title was not resolved for %T: got %#v; want %#v",
					tt.payload, got, want)
			}
		})
	}
}

type customFieldTitleReproTransport func(*http.Request) (*http.Response, error)

func (transport customFieldTitleReproTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}
