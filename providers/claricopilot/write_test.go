package claricopilot

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

	createCallsResponse := testutils.DataFromFile(t, "create_calls.json")
	updateContactsResponse := testutils.DataFromFile(t, "update_contact.json")

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
			Name: "Successfully creation of an call",
			Input: common.WriteParams{ObjectName: "calls", RecordData: map[string]any{
				"source_id": "12345",
				"title":     "Test Call",
				"type":      "RECORDING",
				"status":    "GOTO_MEETING",
				"call_time": "2025-06-05T10:00:00Z",
				"user_emails": []string{
					"integration.user+clari@withampersand.com",
				},
				"source_user_ids": []string{
					"e04483dd-fb82-460a-a14e-b6f3e6a2b7a4",
				},
				"audio_url": "http://file-examples.com/wp-content/storage/2017/11/file_example_MP3_700KB.mp3",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createCallsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
				Data: map[string]any{
					"call_id": "b970b48c-030b-4926-b737-056f35550347",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully update a contact",
			Input: common.WriteParams{
				ObjectName: "contacts",
				RecordId:   "123293450",
				RecordData: map[string]any{
					"first_name": "Johntest updated",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, updateContactsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "123293450",
				Data: map[string]any{
					"id":             "68567029e08c4627fd4f9788",
					"crm_id":         "123293450",
					"account_id":     "684825ed9fe31e20616196e8",
					"first_name":     "Johntest updated",
					"last_name":      "Run",
					"job_title":      "string",
					"phones":         nil,
					"phone_norms":    nil,
					"stage":          nil,
					"reason":         nil,
					"phone_suffixes": nil,
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
