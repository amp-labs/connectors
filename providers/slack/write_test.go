package slack

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Update is not supported",
			Input: common.WriteParams{
				ObjectName: "conversations",
				RecordId:   "C0ABCDEF123",
				RecordData: map[string]any{"name": "general"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			// calls.add returns a nested "call" object with an "id" field.
			Name: "Create call returns nested call ID",
			Input: common.WriteParams{
				ObjectName: "calls",
				RecordData: map[string]any{
					"external_unique_id": "ext-id-123",
					"join_url":           "https://example.com/join",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/calls.add"),
					mockcond.Body(`{"external_unique_id":"ext-id-123","join_url":"https://example.com/join"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"ok": true,
					"call": {
						"id": "R0ABCDEF123",
						"date_start": 1609459200,
						"external_unique_id": "ext-id-123"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "R0ABCDEF123",
				Data: map[string]any{
					"id":                 "R0ABCDEF123",
					"date_start":         float64(1609459200),
					"external_unique_id": "ext-id-123",
				},
			},
			ExpectedErrs: nil,
		},
		{
			// conversations.create returns a nested channel object with an id field.
			Name: "Create conversation returns nested channel ID",
			Input: common.WriteParams{
				ObjectName: "conversations",
				RecordData: map[string]any{"name": "general"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/conversations.create"),
					mockcond.Body(`{"name":"general"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"ok": true,
					"channel": {
						"id": "C0ABCDEF123",
						"name": "general",
						"is_channel": true
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "C0ABCDEF123",
				Data: map[string]any{
					"id":         "C0ABCDEF123",
					"name":       "general",
					"is_channel": true,
				},
			},
			ExpectedErrs: nil,
		},
		{
			// canvases.create returns canvas_id at root level (no wrapper object).
			Name: "Create canvas returns root-level canvas_id",
			Input: common.WriteParams{
				ObjectName: "canvases",
				RecordData: map[string]any{"title": "My Canvas"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/canvases.create"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"ok": true,
					"canvas_id": "F0ABCDEF123"
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "F0ABCDEF123",
				Data: map[string]any{
					"ok":        true,
					"canvas_id": "F0ABCDEF123",
				},
			},
			ExpectedErrs: nil,
		},
		{
			// bookmarks.add returns a nested bookmark object with an id field.
			Name: "Create bookmark returns nested bookmark ID",
			Input: common.WriteParams{
				ObjectName: "bookmarks",
				RecordData: map[string]any{
					"channel_id": "C0ABCDEF123",
					"title":      "My Bookmark",
					"type":       "link",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/bookmarks.add"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"ok": true,
					"bookmark": {
						"id": "Bk0ABCDEF123",
						"channel_id": "C0ABCDEF123",
						"title": "My Bookmark"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "Bk0ABCDEF123",
				Data: map[string]any{
					"id":         "Bk0ABCDEF123",
					"channel_id": "C0ABCDEF123",
					"title":      "My Bookmark",
				},
			},
			ExpectedErrs: nil,
		},
		{

			Name: "Update call",
			Input: common.WriteParams{
				ObjectName: "calls",
				RecordId:   "R0ABCDEF123",
				RecordData: map[string]any{"title": "Updated Call"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/calls.update"),
					mockcond.Body(`{"id":"R0ABCDEF123","title":"Updated Call"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"ok": true,
					"call": {
						"id": "R0ABCDEF123",
						"title": "Updated Call",
						"date_start": 1609459200
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "R0ABCDEF123",
				Data: map[string]any{
					"id":         "R0ABCDEF123",
					"title":      "Updated Call",
					"date_start": float64(1609459200),
				},
			},
			ExpectedErrs: nil,
		},
		{
			// bookmarks.edit routes to the ".edit" suffix and returns a nested bookmark object.
			Name: "Update bookmark",
			Input: common.WriteParams{
				ObjectName: "bookmarks",
				RecordId:   "Bk0ABCDEF123",
				RecordData: map[string]any{"title": "Updated Bookmark"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/bookmarks.edit"),
					mockcond.Body(`{"bookmark_id":"Bk0ABCDEF123","title":"Updated Bookmark"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"ok": true,
					"bookmark": {
						"id": "Bk0ABCDEF123",
						"channel_id": "C0ABCDEF123",
						"title": "Updated Bookmark"
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "Bk0ABCDEF123",
				Data: map[string]any{
					"id":         "Bk0ABCDEF123",
					"channel_id": "C0ABCDEF123",
					"title":      "Updated Bookmark",
				},
			},
			ExpectedErrs: nil,
		},
		{
			// usergroups.update routes to the ".update" suffix and returns a nested usergroup object.
			Name: "Update usergroup",
			Input: common.WriteParams{
				ObjectName: "usergroups",
				RecordId:   "S0ABCDEF123",
				RecordData: map[string]any{"name": "Updated Group"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/usergroups.update"),
					mockcond.Body(`{"name":"Updated Group","usergroup":"S0ABCDEF123"}`),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{
					"ok": true,
					"usergroup": {
						"id": "S0ABCDEF123",
						"name": "Updated Group",
						"date_update": 1609459200
					}
				}`)),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "S0ABCDEF123",
				Data: map[string]any{
					"id":          "S0ABCDEF123",
					"name":        "Updated Group",
					"date_update": float64(1609459200),
				},
			},
			ExpectedErrs: nil,
		},
		{
			// Slack error in write response (ok: false) is mapped to a sentinel error.
			Name: "Slack error response maps to sentinel error",
			Input: common.WriteParams{
				ObjectName: "conversations",
				RecordData: map[string]any{"name": "general"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/conversations.create"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{"ok": false, "error": "not_authed"}`)),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
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
