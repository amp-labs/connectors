package pipedrive

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// nolint
func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	updateActivityResponse := testutils.DataFromFile(t, "update-activity.json")
	unsupportedResponse := testutils.DataFromFile(t, "not-found.json")
	createActivityResponse := testutils.DataFromFile(t, "create-activity.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Unsupported object",
			Input: common.WriteParams{ObjectName: "arsenal", RecordData: map[string]any{"test": "value"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusNotFound, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrRetryable,
				testutils.StringError(string(unsupportedResponse)), // nolint:err113
			},
		},
		{
			Name: "Successful creation of an activity",
			Input: common.WriteParams{
				ObjectName: "activities",
				RecordData: map[string]any{
					"due_date":           "2024-10-30",
					"location":           "Dar es salaam",
					"public_description": "Demo activity",
					"subject":            "I usually can't come up with words",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createActivityResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "9",
				Data: map[string]any{
					"add_time":            "2025-01-14 09:44:56",
					"assigned_to_user_id": float64(20580207),
					"busy_flag":           false,
					"company_id":          float64(13313052),
					"created_by_user_id":  float64(20580207),
					"done":                false,
					"due_date":            "2024-10-30",
					"id":                  float64(9),
					"location":            "Dar es salaam",
					"owner_name":          "Integration User",
					"private":             false,
					"public_description":  "Demo activity",
					"subject":             "I usually can't come up with words",
					"type":                "call",
					"type_name":           "Call",
					"update_time":         "2025-01-14 09:44:56",
					"user_id":             float64(20580207),
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful update an activity",
			Input: common.WriteParams{
				ObjectName: "activities",
				RecordId:   "1",
				RecordData: map[string]any{
					"done": "1",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, updateActivityResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1",
				Data: map[string]any{
					"active_flag":         true,
					"add_time":            "2024-10-15 09:33:57",
					"assigned_to_user_id": float64(20580207),
					"busy_flag":           false,
					"company_id":          float64(13313052),
					"created_by_user_id":  float64(20580207),
					"done":                true,
					"due_date":            "2024-10-15",
					"id":                  float64(1),
					"marked_as_done_time": "2024-10-16 12:22:47",
					"owner_name":          "Integration User",
					"private":             false,
					"reference_type":      "salesphone",
					"subject":             "Make call with sales people",
					"type":                "call",
					"type_name":           "Call",
					"update_time":         "2025-01-14 09:37:13",
					"update_user_id":      float64(20580207),
					"user_id":             float64(20580207),
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL, providers.ModulePipedriveLegacy)
			})
		})
	}
}
