// nolint
package attio

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseObjects := testutils.DataFromFile(t, "objects.json")
	responseLists := testutils.DataFromFile(t, "lists.json")
	responseWorkspace := testutils.DataFromFile(t, "workspace_members.json")
	responseNotes := testutils.DataFromFile(t, "notes.json")
	responseTasks := testutils.DataFromFile(t, "tasks.json")
	responseWebhooks := testutils.DataFromFile(t, "webhooks.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "objects"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown objects are not supported",
			Input:        common.ReadParams{ObjectName: "attributes", Fields: connectors.Fields("")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "An Empty response",
			Input: common.ReadParams{ObjectName: "webhooks", Fields: connectors.Fields("")},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				mockutils.WriteBody(w, `{"data":[]}`)
			})),
			Expected: &common.ReadResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all objects",
			Input: common.ReadParams{ObjectName: "objects", Fields: connectors.Fields("")},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseObjects)
			})),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"object_id":    "9aff8088-dd1f-4e98-ad76-b1e49b9a419a",
						},
						"api_slug":      "people",
						"singular_noun": "Person",
						"plural_noun":   "People",
						"created_at":    "2024-09-17T11:41:18.736000000Z",
					},
				}, {
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"object_id":    "f4df082c-b46e-43e4-a747-f7918b487f44",
						},
						"api_slug":      "companies",
						"singular_noun": "Company",
						"plural_noun":   "Companies",
						"created_at":    "2024-09-17T11:41:18.736000000Z",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all lists",
			Input: common.ReadParams{ObjectName: "lists", Fields: connectors.Fields("")},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseLists)
			})),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"list_id":      "7ddc974a-2ab2-4a96-a83e-853eacb0329f",
						},
						"api_slug":         "sales_6",
						"created_at":       "2024-09-25T13:10:33.302000000Z",
						"name":             "Sales",
						"workspace_access": nil,
						"workspace_member_access": []interface{}{
							map[string]any{
								"level":               "full-access",
								"workspace_member_id": "cc3821aa-f738-42c0-a739-7b6de755e5f1",
							},
						},
						"parent_object": []any{
							"companies",
						},
						"created_by_actor": map[string]any{
							"type": "workspace-member",
							"id":   "cc3821aa-f738-42c0-a739-7b6de755e5f1",
						},
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseWorkspace)
			})),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id":        "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"workspace_member_id": "67af46e4-a450-4fee-a1d1-39729b3af771",
						},
						"first_name":    "Integration",
						"last_name":     "User",
						"avatar_url":    nil,
						"email_address": "integration.user@withampersand.com",
						"access_level":  "admin",
						"created_at":    "2024-09-17T11:41:21.779000000Z",
					},
				}, {
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id":        "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"workspace_member_id": "cc3821aa-f738-42c0-a739-7b6de755e5f1",
						},
						"first_name":    "Sanjay Kanth",
						"last_name":     "A",
						"avatar_url":    "https://lh3.googleusercontent.com/a/ACg8ocL0Zwi9XArWL-GgYiUqPoDMKS1p1AQuQPPHVMQr0V023Ox_fYY=s96-c",
						"email_address": "sanjaykanth.a@mitrahsoft.com",
						"access_level":  "admin",
						"created_at":    "2024-09-17T12:17:11.682000000Z",
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseNotes)
			})),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"note_id":      "32dc76ee-d094-40e1-b176-0f8e1b772f0a",
						},
						"parent_object":     "companies",
						"parent_record_id":  "ec902ed9-aab7-4347-8e26-dca240ffba08",
						"title":             "value",
						"content_plaintext": "",
						"created_by_actor": map[string]any{
							"type": "workspace-member",
							"id":   "cc3821aa-f738-42c0-a739-7b6de755e5f1",
						},
						"created_at": "2024-09-24T08:31:09.211000000Z",
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseTasks)
			})),
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
						"is_completed":      false,
						"deadline_at":       "2023-01-01T15:00:00.000000000Z",
						"linked_records":    []any{},
						"assignees": []interface{}{
							map[string]any{
								"referenced_actor_type": "workspace-member",
								"referenced_actor_id":   "67af46e4-a450-4fee-a1d1-39729b3af771",
							},
						},
						"created_by_actor": map[string]any{
							"type": "api-token",
							"id":   "53b1e97a-08d6-4d2e-856d-5371bb6f4052",
						},
						"created_at": "2024-09-23T11:04:07.647000000Z",
					},
				},
				},
				NextPage: "test?limit=10&offset=10",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all webhooks",
			Input: common.ReadParams{ObjectName: "webhooks", Fields: connectors.Fields(""), NextPage: "test?limit=10"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseWebhooks)
			})),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{},
					Raw: map[string]any{
						"id": map[string]any{
							"workspace_id": "0d4d7fa2-d6e8-4a61-a7dc-e178405ff3c6",
							"webhook_id":   "02303005-d363-4f49-8fd2-19944003b1d2",
						},
						"target_url": "https://example.com/webhook",
						"status":     "active",
						"subscriptions": []interface{}{
							map[string]any{
								"event_type": "note.created",
								"filter": map[string]any{
									"$or": []interface{}{
										map[string]any{
											"field":    "id.list_id",
											"operator": "equals",
											"value":    "d34c5f6b-0410-4830-853d-8a6602252687",
										},
									},
								},
							},
						},
						"created_at": "2024-09-20T11:14:17.403000000Z",
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
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine.
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
