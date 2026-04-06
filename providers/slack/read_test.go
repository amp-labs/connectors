package slack

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	conversationsResponse := testutils.DataFromFile(t, "conversations-list.json")
	conversationsFirstPageResponse := testutils.DataFromFile(t, "conversations-first-page.json")
	usersResponse := testutils.DataFromFile(t, "users-list.json")
	errorWithMessageResponse := testutils.DataFromFile(t, "error-with-message.json")
	errorWithoutMessageResponse := testutils.DataFromFile(t, "error-without-message.json")

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
				If:    mockcond.Path("/api/conversations.list"),
				Then:  mockserver.Response(http.StatusOK, conversationsFirstPageResponse),
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
				If:    mockcond.Path("/api/conversations.list"),
				Then:  mockserver.Response(http.StatusOK, errorWithMessageResponse),
			}.Server(),
			ExpectedErrs: []error{testutils.StringError("response indicated failure: not_authed")},
		},
		{
			Name:  "Slack error response without message returns generic failure error",
			Input: common.ReadParams{ObjectName: "conversations", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/conversations.list"),
				Then:  mockserver.Response(http.StatusOK, errorWithoutMessageResponse),
			}.Server(),
			ExpectedErrs: []error{testutils.StringError("response indicated failure")},
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
