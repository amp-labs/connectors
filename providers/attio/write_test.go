// nolint
package attio

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

	listResponse := testutils.DataFromFile(t, "write_lists.json")
	notesresponse := testutils.DataFromFile(t, "write_notes.json")
	tasksResponse := testutils.DataFromFile(t, "write_tasks.json")
	companiesRecordResponse := testutils.DataFromFile(t, "write_companies_record.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Create lists as POST",
			Input: common.WriteParams{ObjectName: "lists", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, listResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "e09a041c-0555-4bb2-8f6e-997bfc9b54e8",
				Errors:   nil,
				Data: map[string]any{
					"api_slug": "sales_investing",
					"id": map[string]any{
						"list_id":      "e09a041c-0555-4bb2-8f6e-997bfc9b54e8",
						"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
					},
					"name": "Sales",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update lists as PATCH",
			Input: common.WriteParams{ObjectName: "lists", RecordId: "e09a041c-0555-4bb2-8f6e-997bfc9b54e8", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, listResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "e09a041c-0555-4bb2-8f6e-997bfc9b54e8",
				Errors:   nil,
				Data: map[string]any{
					"api_slug": "sales_investing",
					"id": map[string]any{
						"list_id":      "e09a041c-0555-4bb2-8f6e-997bfc9b54e8",
						"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
					},
					"name": "Sales",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create notes as POST",
			Input: common.WriteParams{ObjectName: "notes"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, notesresponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "126e58a5-5e3f-4644-89ff-6474e97fcecd",
				Errors:   nil,
				Data: map[string]any{
					"content_plaintext": "summary",
					"id": map[string]any{
						"note_id":      "126e58a5-5e3f-4644-89ff-6474e97fcecd",
						"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
					},
					"title": "Call summary",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create tasks as POST",
			Input: common.WriteParams{ObjectName: "tasks", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, tasksResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "b38142c7-00f6-4d92-813e-7b0f689a5873",
				Errors:   nil,
				Data: map[string]any{
					"content_plaintext": "view summary",
					"id": map[string]any{
						"task_id":      "b38142c7-00f6-4d92-813e-7b0f689a5873",
						"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update tasks as PATCH",
			Input: common.WriteParams{ObjectName: "tasks", RecordId: "bf012982-06a9-47f7-9e87-07dc4945d502", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, tasksResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "b38142c7-00f6-4d92-813e-7b0f689a5873",
				Errors:   nil,
				Data: map[string]any{
					"content_plaintext": "view summary",
					"id": map[string]any{
						"task_id":      "b38142c7-00f6-4d92-813e-7b0f689a5873",
						"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create record for standard object as POST",
			Input: common.WriteParams{ObjectName: "companies", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, companiesRecordResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "238747c9-6dce-4475-8928-fe66ddfe8526",
				Errors:   nil,
				Data: map[string]any{
					"id": map[string]any{
						"workspace_id": "63d34516-b287-4c27-9d28-fe2adbebcd50",
						"object_id":    "1fa986a6-952e-4e92-ba01-acca61a7b616",
						"record_id":    "238747c9-6dce-4475-8928-fe66ddfe8526",
					},
					"values": map[string]any{
						"name": []any{
							map[string]any{
								"active_from":  "2025-03-21T10:27:09.906000000Z",
								"active_until": nil,
								"created_by_actor": map[string]any{
									"type": "api-token",
									"id":   "0d91718e-9193-47e7-8b6c-6feeae40dcc1",
								},
								"value":          "pay",
								"attribute_type": "text",
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update record for standard object as PUT",
			Input: common.WriteParams{ObjectName: "companies", RecordId: "238747c9-6dce-4475-8928-fe66ddfe8526", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, companiesRecordResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "238747c9-6dce-4475-8928-fe66ddfe8526",
				Errors:   nil,
				Data: map[string]any{
					"id": map[string]any{
						"workspace_id": "63d34516-b287-4c27-9d28-fe2adbebcd50",
						"object_id":    "1fa986a6-952e-4e92-ba01-acca61a7b616",
						"record_id":    "238747c9-6dce-4475-8928-fe66ddfe8526",
					},
					"values": map[string]any{
						"name": []any{
							map[string]any{
								"active_from":  "2025-03-21T10:27:09.906000000Z",
								"active_until": nil,
								"created_by_actor": map[string]any{
									"type": "api-token",
									"id":   "0d91718e-9193-47e7-8b6c-6feeae40dcc1",
								},
								"value":          "pay",
								"attribute_type": "text",
							},
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
