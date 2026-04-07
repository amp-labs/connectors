package slack

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	conversationsResponse := testutils.DataFromFile(t, "conversations-list.json")
	conversationsFirstPageResponse := testutils.DataFromFile(t, "conversations-first-page.json")
	usersResponse := testutils.DataFromFile(t, "users-list.json")
	errorWithMessageResponse := testutils.DataFromFile(t, "error-with-message.json")
	errorWithoutMessageResponse := testutils.DataFromFile(t, "error-without-message.json")
	listConnectInvitesResponse := testutils.DataFromFile(t, "conversations-list-connect-invites.json")
	requestSharedInviteResponse := testutils.DataFromFile(t, "conversations-request-shared-invite.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of conversations",
			Input: common.ReadParams{ObjectName: "conversations", Fields: connectors.Fields("id", "name", "is_private")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/conversations.list"),
					mockcond.QueryParam("limit", "200"),
				},
				Then: mockserver.Response(http.StatusOK, conversationsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         "C0ABCDEF123",
						"name":       "general",
						"is_private": false,
					},
					Raw: map[string]any{
						"is_channel":  true,
						"is_archived": false,
						"is_member":   true,
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of users",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id", "name", "is_bot")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/users.list"),
					mockcond.QueryParam("limit", "200"),
				},
				Then: mockserver.Response(http.StatusOK, usersResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":     "W012A3CDE",
						"name":   "spengler",
						"is_bot": false,
					},
					Raw: map[string]any{
						"team_id":   "T012AB3C4",
						"real_name": "Egon Spengler",
						"is_admin":  true,
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "First page response includes next cursor",
			Input: common.ReadParams{ObjectName: "conversations", Fields: connectors.Fields("id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/conversations.list"),
				},
				Then: mockserver.Response(http.StatusOK, conversationsFirstPageResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "dGVhbS1jaGFubmVsOkM=",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Slack error response with message returns failure error",
			Input: common.ReadParams{ObjectName: "conversations", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/conversations.list"),
				},
				Then: mockserver.Response(http.StatusOK, errorWithMessageResponse),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
		},
		{
			Name:  "Slack error response without message returns generic failure error",
			Input: common.ReadParams{ObjectName: "conversations", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/conversations.list"),
				},
				Then: mockserver.Response(http.StatusOK, errorWithoutMessageResponse),
			}.Server(),
			ExpectedErrs: []error{common.ErrBadProviderResponse},
		},
		{
			Name:  "Read conversations.listConnectInvites via POST",
			Input: common.ReadParams{ObjectName: "conversations.listConnectInvites", Fields: connectors.Fields("id", "channel_id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/conversations.listConnectInvites"),
				},
				Then: mockserver.Response(http.StatusOK, listConnectInvitesResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         "I0ABCDEF123",
						"channel_id": "C0ABCDEF123",
					},
					Raw: map[string]any{
						"inviting_team": map[string]any{
							"id":   "T012AB3C4",
							"name": "Acme Corp",
						},
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read conversations.requestSharedInvite via POST",
			Input: common.ReadParams{ObjectName: "conversations.requestSharedInvite", Fields: connectors.Fields("id", "invite_id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/conversations.requestSharedInvite.list"),
				},
				Then: mockserver.Response(http.StatusOK, requestSharedInviteResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":        "I0XYZ789",
						"invite_id": "INV123",
					},
					Raw: map[string]any{
						"channel": map[string]any{
							"id":   "C0ABCDEF123",
							"name": "general",
						},
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			// conversations.updated = 1449252889 (2015-12-04). Since is set before that,
			// so the record should pass through the client-side filter.
			Name: "Since filter includes record updated after threshold",
			Input: common.ReadParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("id", "name"),
				Since:      time.Unix(1449252800, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/conversations.list"),
				},
				Then: mockserver.Response(http.StatusOK, conversationsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "C0ABCDEF123",
						"name": "general",
					},
					Raw: map[string]any{
						"is_channel": true,
						"is_member":  true,
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			// conversations.updated = 1449252889 (2015-12-04). Since is set after that,
			// so the record is too old and should be filtered out.
			Name: "Since filter excludes record updated before threshold",
			Input: common.ReadParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("id", "name"),
				Since:      time.Unix(1449252900, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/conversations.list"),
				},
				Then: mockserver.Response(http.StatusOK, conversationsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			// users.updated = 1502138686 (2017-08-07). The Since/Until window contains
			// that timestamp, so the record should be returned.
			Name: "Since and Until range includes user record",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "name"),
				Since:      time.Unix(1502138600, 0), // just before updated
				Until:      time.Unix(1502138700, 0), // just after updated
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/users.list"),
				},
				Then: mockserver.Response(http.StatusOK, usersResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "W012A3CDE",
						"name": "spengler",
					},
					Raw: map[string]any{
						"team_id":   "T012AB3C4",
						"real_name": "Egon Spengler",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			// users.updated = 1502138686 (2017-08-07). The Until is set before that,
			// so the record is too new and should be filtered out.
			Name: "Until filter excludes record updated after threshold",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("id", "name"),
				Until:      time.Unix(1502138600, 0), // just before updated
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/users.list"),
				},
				Then: mockserver.Response(http.StatusOK, usersResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Following cursor passes it as query param",
			Input: common.ReadParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("id", "name"),
				NextPage:   "dGVhbS1jaGFubmVsOkM=",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/conversations.list"),
					mockcond.QueryParam("cursor", "dGVhbS1jaGFubmVsOkM="),
				},
				Then: mockserver.Response(http.StatusOK, conversationsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "",
				Done:     true,
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
