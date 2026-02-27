// nolint
package attio

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
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
	responseCompanies := testutils.DataFromFile(t, "companies_read.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of all lists",
			Input: common.ReadParams{ObjectName: "lists", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/lists"),
				Then:  mockserver.Response(http.StatusOK, responseLists),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": map[string]any{
								"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
								"list_id":      "7ddc974a-2ab2-4a96-a83e-853eacb0329f",
							},
						},
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
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/workspace_members"),
				Then:  mockserver.Response(http.StatusOK, responseWorkspace),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": map[string]any{
							"workspace_id":        "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"workspace_member_id": "67af46e4-a450-4fee-a1d1-39729b3af771",
						},
					},
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
			Input: common.ReadParams{ObjectName: "notes", Fields: connectors.Fields(""), NextPage: "test?limit=50"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseNotes),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": map[string]any{
								"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
								"note_id":      "32dc76ee-d094-40e1-b176-0f8e1b772f0a",
							},
						},
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
				NextPage: "test?limit=50&offset=50",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all tasks",
			Input: common.ReadParams{ObjectName: "tasks", Fields: connectors.Fields(""), NextPage: "test?limit=500"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseTasks),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": map[string]any{
								"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
								"task_id":      "4a585693-fa14-4ead-9e19-cc9251df59be",
							},
						},
						Raw: map[string]any{
							"id": map[string]any{
								"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
								"task_id":      "4a585693-fa14-4ead-9e19-cc9251df59be",
							},
							"content_plaintext": "Follow up on current software solutions",
						},
					},
				},
				NextPage: "test?limit=500&offset=500",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all companies",
			Input: common.ReadParams{ObjectName: "companies", Fields: connectors.Fields("name"), Since: time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC), NextPage: "test?limit=500"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseCompanies),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"name": []any{
								map[string]any{
									"active_from":  "2025-03-12T07:55:38.981000000Z",
									"active_until": nil,
									"created_by_actor": map[string]any{
										"type": "system",
										"id":   nil,
									},
									"value":          "Attio",
									"attribute_type": "text",
								},
							},
						},
						Raw: map[string]any{
							"id": map[string]any{
								"workspace_id": "63d34516-b287-4c27-9d28-fe2adbebcd50",
								"object_id":    "1fa986a6-952e-4e92-ba01-acca61a7b616",
								"record_id":    "2db97cee-6c6b-4486-ae52-db8e4b6f44e9",
							},
							"created_at": "2025-03-12T07:55:38.327000000Z",
							"values": map[string]any{
								"name": []any{
									map[string]any{
										"active_from":  "2025-03-12T07:55:38.981000000Z",
										"active_until": nil,
										"created_by_actor": map[string]any{
											"type": "system",
											"id":   nil,
										},
										"value":          "Attio",
										"attribute_type": "text",
									},
								},
							},
						},
					},
				},
				NextPage: "500",
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
