package pylon

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	createTasksResponse := testutils.DataFromFile(t, "create-tasks.json")

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
			Name: "Successfully creation of a task",
			Input: common.WriteParams{ObjectName: "tasks", RecordData: map[string]any{
				"title": "ampersand test dev",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createTasksResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "c265bbc5-fcbb-4cd1-a470-212c3caa8b63",
				Data: map[string]any{
					"id":                      "c265bbc5-fcbb-4cd1-a470-212c3caa8b63",
					"title":                   "ampersand test dev",
					"body_html":               "",
					"status":                  "not_started",
					"created_at":              "2025-08-29T16:58:59Z",
					"updated_at":              "2025-08-29T16:58:59Z",
					"customer_portal_visible": false,
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
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
