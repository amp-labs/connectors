package gotoconn

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	createGroupsResponse := testutils.DataFromFile(t, "create-groups.json")
	createTeamsResponse := testutils.DataFromFile(t, "create-teams.json")
	createTemplatesResponse := testutils.DataFromFile(t, "create-templates.json")
	createWebhookResponse := testutils.DataFromFile(t, "create-webhook.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Read-only object rejects write",
			Input: common.WriteParams{
				ObjectName: "sessions",
				RecordData: map[string]any{"foo": "bar"},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "SCIM update using PATCH",
			Input: common.WriteParams{
				ObjectName: "users",
				RecordId:   "user-123",
				RecordData: map[string]any{"displayName": "Alice"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/identity/v1/Users/user-123"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{"id":"user-123","displayName":"Alice"}`)),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "user-123",
				Data: map[string]any{
					"id":          "user-123",
					"displayName": "Alice",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create group (SCIM)",
			Input: common.WriteParams{
				ObjectName: "groups",
				RecordData: map[string]any{
					"displayName": "group test",
					"members": []map[string]any{
						{"type": "group", "value": "test"},
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/identity/v1/Groups"),
				},
				Then: mockserver.Response(http.StatusOK, createGroupsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "dfoahofadso",
				Data: map[string]any{
					"id":          "dfoahofadso",
					"displayName": "group test",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create team (Corporate)",
			Input: common.WriteParams{
				ObjectName: "teams",
				RecordData: map[string]any{
					"teamName":      "Support EU",
					"parentKey":     12,
					"portalKey":     34,
					"subPortalKeys": []int{56, 78},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/G2AC/rest/v1/teams"),
				},
				Then: mockserver.Response(http.StatusOK, createTeamsResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success: true,
				Data: map[string]any{
					"teamKey": float64(0),
				},
				RecordId: "0",
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Create template (Admin)",
			Input: common.WriteParams{
				ObjectName: "templates",
				RecordData: map[string]any{
					"title":   "Welcome",
					"subject": "Hi there",
					"text":    "Thanks for joining!",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/admin/rest/v1/accounts/" + testAccountKey + "/templates"),
				},
				Then: mockserver.Response(http.StatusOK, createTemplatesResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "kdhoasdofaso",
				Data:     map[string]any{},
			},
			ExpectedErrs: nil,
		},
		{
			// webhooks is an array-body endpoint: we accept a single record
			// object and must wrap it into a one-element array in the request.
			Name: "Create webhook wraps single object into array body",
			Input: common.WriteParams{
				ObjectName: "webhooks",
				RecordData: map[string]any{
					"callbackUrl":  "https://example.com/hook",
					"eventName":    "registrant.joined",
					"eventVersion": "1.0.0",
					"product":      "g2w",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/G2W/rest/v2/webhooks"),
					// the single record object must be wrapped in a one-element array
					mockcond.Body(`[{"callbackUrl":"https://example.com/hook","eventName":"registrant.joined","eventVersion":"1.0.0","product":"g2w"}]`), //nolint:lll
				},
				Then: mockserver.Response(http.StatusOK, createWebhookResponse),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cdb516e0-85f8-4e30-bf19-ddb0ea9d40c2",
				Data: map[string]any{
					"webhookKey":  "cdb516e0-85f8-4e30-bf19-ddb0ea9d40c2",
					"callbackUrl": "https://example.com/hook",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		//nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
