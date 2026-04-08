package supersend

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

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseTeams := testutils.DataFromFile(t, "read/teams.json")
	responseSendersFirstPage := testutils.DataFromFile(t, "read/senders-first-page.json")
	responseSendersLastPage := testutils.DataFromFile(t, "read/senders-last-page.json")
	responseConversations := testutils.DataFromFile(t, "read/conversations.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "teams"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Read teams without pagination",
			Input: common.ReadParams{
				ObjectName: "teams",
				Fields:     connectors.Fields("id", "name", "domain", "isDefault"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/teams"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseTeams),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":        "team-001",
							"name":      "Engineering Team",
							"domain":    "engineering.example.com",
							"isdefault": true,
						},
						Raw: map[string]any{
							"about": "Product engineering team",
							"OrgId": "org-001",
						},
					},
					{
						Fields: map[string]any{
							"id":        "team-002",
							"name":      "Sales Team",
							"domain":    "sales.example.com",
							"isdefault": false,
						},
						Raw: map[string]any{
							"about": "Sales and outreach team",
							"OrgId": "org-001",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read senders first page with pagination",
			Input: common.ReadParams{
				ObjectName: "senders",
				Fields:     connectors.Fields("id", "email", "warm", "max_per_day"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/senders"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseSendersFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":          "sender-001",
							"email":       "john@example.com",
							"warm":        true,
							"max_per_day": float64(100),
						},
						Raw: map[string]any{
							"send_as": "John Doe",
							"TeamId":  "team-001",
						},
					},
					{
						Fields: map[string]any{
							"id":          "sender-002",
							"email":       "jane@example.com",
							"warm":        false,
							"max_per_day": float64(50),
						},
						Raw: map[string]any{
							"send_as": "Jane Smith",
							"TeamId":  "team-001",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/v1/senders?limit=100&offset=100",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read senders last page using NextPage token",
			Input: common.ReadParams{
				ObjectName: "senders",
				Fields:     connectors.Fields("id", "email"),
				NextPage:   testroutines.URLTestServer + "/v1/senders?limit=100&offset=200",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/senders"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("offset", "200"),
				},
				Then: mockserver.Response(http.StatusOK, responseSendersLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    "sender-003",
							"email": "alex@example.com",
						},
						Raw: map[string]any{
							"send_as": "Alex Johnson",
							"TeamId":  "team-002",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read conversations (nested responseKey)",
			Input: common.ReadParams{
				ObjectName: "conversation/latest-by-profile",
				Fields:     connectors.Fields("id", "title", "is_unread", "platform_type"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/conversation/latest-by-profile"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseConversations),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":            "conv-001",
							"title":         "Meeting follow-up",
							"is_unread":     false,
							"platform_type": float64(1),
						},
						Raw: map[string]any{
							"inbox_mood": "positive",
							"contact_id": "contact-001",
						},
					},
					{
						Fields: map[string]any{
							"id":            "conv-002",
							"title":         "Product inquiry",
							"is_unread":     true,
							"platform_type": float64(2),
						},
						Raw: map[string]any{
							"inbox_mood": "neutral",
							"contact_id": "contact-002",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read with custom page size and offset pagination",
			Input: common.ReadParams{
				ObjectName: "senders",
				Fields:     connectors.Fields("id", "email"),
				PageSize:   50,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/senders"),
					mockcond.QueryParam("limit", "50"),
				},
				Then: mockserver.Response(http.StatusOK, responseSendersFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    "sender-001",
							"email": "john@example.com",
						},
						Raw: map[string]any{
							"send_as": "John Doe",
						},
					},
					{
						Fields: map[string]any{
							"id":    "sender-002",
							"email": "jane@example.com",
						},
						Raw: map[string]any{
							"send_as": "Jane Smith",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/v1/senders?limit=50&offset=50",
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

// constructTestConnector is defined in metadata_test.go
