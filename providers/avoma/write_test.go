package avoma

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	smartCategoriesResponse := testutils.DataFromFile(t, "create_smart_categories.json")
	updateSmartCategoriesResponse := testutils.DataFromFile(t, "update_smart_categories.json")
	callsResponse := testutils.DataFromFile(t, "create_calls.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the smart categories",
			Input: common.WriteParams{ObjectName: "smart_categories", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, smartCategoriesResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "5b66d318-627e-4336-9eec-bd79212eb1db",
				Errors:   nil,
				Data: map[string]any{
					"keywords": []any{
						"demo",
					},
					"name": "demo",
					"prompts": []any{
						"smart",
					},
					"settings": map[string]any{
						"aug_notes_enabled":       true,
						"keyword_notes_enabled":   true,
						"prompt_extract_length":   "short",
						"prompt_extract_strategy": "after",
					},
					"uuid": "5b66d318-627e-4336-9eec-bd79212eb1db",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "updating the smart categories",
			Input: common.WriteParams{
				ObjectName: "smart_categories",
				RecordData: "dummy",
				RecordId:   "5b66d318-627e-4336-9eec-bd79212eb1db",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, updateSmartCategoriesResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
				Errors:  nil,
				Data: map[string]any{
					"keywords": []any{
						"demo",
					},
					"prompts": []any{
						"smart",
					},
					"settings": map[string]any{
						"aug_notes_enabled":       true,
						"keyword_notes_enabled":   true,
						"prompt_extract_length":   "short",
						"prompt_extract_strategy": "after",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Creating calls",
			Input: common.WriteParams{ObjectName: "calls", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, callsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "Xkeltda345",
				Errors:   nil,
				Data: map[string]any{
					"external_id":   "Xkeltda345",
					"user_email":    "sample@gmail.com",
					"state":         "created",
					"frm":           "+11234567890",
					"to":            "+12234567890",
					"start_at":      "2025-06-12T20:00:00Z",
					"recording_url": "https://example.com/recording.mp3",
					"direction":     "Outbound",
					"meeting": map[string]any{
						"id":              float64(36422168),
						"created":         "2025-06-12T13:19:26.807495Z",
						"modified":        "2025-06-12T13:19:27.288047Z",
						"uuid":            "ef52f35e-14ef-47ea-bca8-dc94b9ce7eda",
						"subject":         "Call with string (+12234567890) on June 12, 2025",
						"organizer_email": "sample@gmail.com",
						"external_id":     "ringcentral_string",
						"ical_uid":        "",
						"state":           "ended",
						"start_at":        "2025-06-12T20:00:00Z",
					},
					"source": "ringcentral",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
