package flatfile

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

	createEventsResponse := testutils.DataFromFile(t, "create-events.json")
	updateEnvironmentsResponse := testutils.DataFromFile(t, "update-environments.json")

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
			Name: "Successfully creation of an event",
			Input: common.WriteParams{ObjectName: "events", RecordData: map[string]any{
				"context": map[string]any{
					"accountId":     "us_acc_YOUR_ID",
					"environmentId": "us_env_YOUR_ID",
					"spaceId":       "us_sp_YOUR_ID",
					"workbookId":    "us_wb_YOUR_ID",
					"actorId":       "us_key_SOME_KEY",
				},
				"domain": "workbook",
				"payload": map[string]any{
					"recordsAdded": 100,
				},
				"topic": "workbook:updated",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createEventsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "us_wb_YOUR_ID_event_1234567890",
				Data: map[string]any{
					"id":     "us_wb_YOUR_ID_event_1234567890",
					"topic":  "workbook:updated",
					"domain": "workbook",
					"context": map[string]any{
						"accountId":     "us_acc_YOUR_ID",
						"environmentId": "us_env_YOUR_ID",
						"spaceId":       "us_sp_YOUR_ID",
						"workbookId":    "us_wb_YOUR_ID",
						"actorId":       "us_key_SOME_KEY",
					},
					"dataUrl":     "",
					"createdAt":   "2023-11-07T20:46:04.300Z",
					"callbackUrl": "",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully update of an environment",
			Input: common.WriteParams{ObjectName: "environments", RecordId: "4784189d-610b-4488-b3a5-5f324f752417", RecordData: map[string]any{ //nolint:lll
				"name": "updated environment",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, updateEnvironmentsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "4784189d-610b-4488-b3a5-5f324f752417",
				Data: map[string]any{
					"id":        "4784189d-610b-4488-b3a5-5f324f752417",
					"accountId": "us_acc_YOUR_ID",
					"name":      "updated environment",
					"isProd":    false,
					"guestAuthentication": []any{
						"magic_link",
					},
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
