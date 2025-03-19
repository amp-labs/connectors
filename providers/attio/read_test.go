// nolint
package attio

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseLists := testutils.DataFromFile(t, "lists.json")
	responseWorkspace := testutils.DataFromFile(t, "workspace_members.json")
	responseNotes := testutils.DataFromFile(t, "notes.json")
	responseTasks := testutils.DataFromFile(t, "tasks.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unknown objects are not supported",
			Input:        common.ReadParams{ObjectName: "attributes", Fields: connectors.Fields("")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Read list of all lists",
			Input: common.ReadParams{ObjectName: "lists", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseLists),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"list_id":      "7ddc974a-2ab2-4a96-a83e-853eacb0329f",
						},
						"api_slug": "sales_6",
						"name":     "Sales",
					},
				},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all workspace members",
			Input: common.ReadParams{ObjectName: "workspace_members", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseWorkspace),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id":        "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"workspace_member_id": "67af46e4-a450-4fee-a1d1-39729b3af771",
						},
						"first_name":    "Integration",
						"last_name":     "User",
						"email_address": "integration.user@withampersand.com",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all notes",
			Input: common.ReadParams{ObjectName: "notes", Fields: connectors.Fields(""), NextPage: "test?limit=10"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseNotes),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"note_id":      "32dc76ee-d094-40e1-b176-0f8e1b772f0a",
						},
						"title":             "value",
						"content_plaintext": "",
					},
				},
				},
				NextPage: "test?limit=10&offset=10",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all tasks",
			Input: common.ReadParams{ObjectName: "tasks", Fields: connectors.Fields(""), NextPage: "test?limit=10"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseTasks),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"task_id":      "4a585693-fa14-4ead-9e19-cc9251df59be",
						},
						"content_plaintext": "Follow up on current software solutions",
					},
				},
				},
				NextPage: "test?limit=10&offset=10",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
