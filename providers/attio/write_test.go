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
	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.WriteParams{ObjectName: "attributes", RecordData: "dummy"},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
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
					"api_slug":   "sales_investing",
					"created_at": "2024-10-04T10:03:02.070000000Z",
					"created_by_actor": map[string]any{
						"id":   "00bcd19c-9a89-467f-bf61-4e1a7f8b8754",
						"type": "api-token",
					},
					"id": map[string]any{
						"list_id":      "e09a041c-0555-4bb2-8f6e-997bfc9b54e8",
						"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
					},
					"name": "Sales",
					"parent_object": []any{
						"companies",
					},
					"workspace_access": "full-access",
					"workspace_member_access": []any{
						map[string]any{
							"level":               "full-access",
							"workspace_member_id": "67af46e4-a450-4fee-a1d1-39729b3af771",
						},
					},
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
					"api_slug":   "sales_investing",
					"created_at": "2024-10-04T10:03:02.070000000Z",
					"created_by_actor": map[string]any{
						"id":   "00bcd19c-9a89-467f-bf61-4e1a7f8b8754",
						"type": "api-token",
					},
					"id": map[string]any{
						"list_id":      "e09a041c-0555-4bb2-8f6e-997bfc9b54e8",
						"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
					},
					"name": "Sales",
					"parent_object": []any{
						"companies",
					},
					"workspace_access": "full-access",
					"workspace_member_access": []any{
						map[string]any{
							"level":               "full-access",
							"workspace_member_id": "67af46e4-a450-4fee-a1d1-39729b3af771",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Create notes as POST",
			Input: common.WriteParams{ObjectName: "notes", RecordData: "dummy"},
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
					"created_at":        "2024-10-04T10:03:03.347000000Z",
					"created_by_actor": map[string]any{
						"id":   "00bcd19c-9a89-467f-bf61-4e1a7f8b8754",
						"type": "api-token",
					},
					"id": map[string]any{
						"note_id":      "126e58a5-5e3f-4644-89ff-6474e97fcecd",
						"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
					},
					"parent_object":    "companies",
					"parent_record_id": "ec902ed9-aab7-4347-8e26-dca240ffba08",
					"title":            "Call summary",
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
					"assignees": []any{
						map[string]any{
							"referenced_actor_id":   "67af46e4-a450-4fee-a1d1-39729b3af771",
							"referenced_actor_type": "workspace-member",
						},
					},
					"content_plaintext": "view summary",
					"created_at":        "2024-10-04T10:03:04.216000000Z",
					"created_by_actor": map[string]any{
						"id":   "00bcd19c-9a89-467f-bf61-4e1a7f8b8754",
						"type": "api-token",
					},
					"deadline_at": "2023-10-04T15:00:00.000000000Z",
					"id": map[string]any{
						"task_id":      "b38142c7-00f6-4d92-813e-7b0f689a5873",
						"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
					},
					"is_completed": false,
					"linked_records": []any{
						map[string]any{
							"target_object_id": "f4df082c-b46e-43e4-a747-f7918b487f44",
							"target_record_id": "ec902ed9-aab7-4347-8e26-dca240ffba08",
						},
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
					"assignees": []any{
						map[string]any{
							"referenced_actor_id":   "67af46e4-a450-4fee-a1d1-39729b3af771",
							"referenced_actor_type": "workspace-member",
						},
					},
					"content_plaintext": "view summary",
					"created_at":        "2024-10-04T10:03:04.216000000Z",
					"created_by_actor": map[string]any{
						"id":   "00bcd19c-9a89-467f-bf61-4e1a7f8b8754",
						"type": "api-token",
					},
					"deadline_at": "2023-10-04T15:00:00.000000000Z",
					"id": map[string]any{
						"task_id":      "b38142c7-00f6-4d92-813e-7b0f689a5873",
						"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
					},
					"is_completed": false,
					"linked_records": []any{
						map[string]any{
							"target_object_id": "f4df082c-b46e-43e4-a747-f7918b487f44",
							"target_record_id": "ec902ed9-aab7-4347-8e26-dca240ffba08",
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
