package clickup

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorInvalidWorkspace := testutils.DataFromFile(t, "read/invalid-workspace.json")
	errorInvalidUrl := testutils.DataFromFile(t, "read/invalid-url.txt")
	responseTasks := testutils.DataFromFile(t, "read/tasks.json")
	responseChannels := testutils.DataFromFile(t, "read/channels.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "tasks"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Text media type error response is understood",
			Input: common.ReadParams{ObjectName: "tasks", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorInvalidUrl),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				interpreter.ErrUnknownResponseFormat,
				testutils.StringError("404 page not found"),
			},
		},
		{
			Name:  "JSON media type error response is understood",
			Input: common.ReadParams{ObjectName: "tasks", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorInvalidWorkspace),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("OAUTH_192: Workspace not authorized"),
			},
		},
		{
			Name: "List tasks with offset pagination",
			Input: common.ReadParams{
				ObjectName: "tasks",
				Fields: connectors.Fields(
					"name",
					// Custom fields:
					"Age",
					"Documentation Link",
				),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v2/team/test-workspace-123/task"),
					mockcond.QueryParam("page", "0"),
				},
				Then: mockserver.Response(http.StatusOK, responseTasks),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name":               "Brand new task from Dashboard",
						"age":                "21",
						"documentation link": nil,
					},
					Raw: map[string]any{"description": "Some important description text should go here."},
					Id:  "869ff7zem",
				}, {
					Fields: map[string]any{
						"name":               "Server Migration to AWS Central",
						"age":                nil,
						"documentation link": "https://developer.clickup.com/reference/getwebhooks",
					},
					Raw: map[string]any{"description": "To be done soon."},
					Id:  "869ff5pu0",
				}},
				NextPage: testconn.URLTestServer + "/api/v2/team/test-workspace-123/task?page=1",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "List channels with cursor pagination",
			Input: common.ReadParams{
				ObjectName: "channels",
				Fields:     connectors.Fields("type", "latest_comment_at"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v3/workspaces/test-workspace-123/chat/channels"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseChannels),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"type":              "ASK_AI",
						"latest_comment_at": float64(1791576886312),
					},
					Raw: map[string]any{"creator": "308680077"},
					Id:  "2kxv6e49-92",
				}},
				NextPage: testconn.URLTestServer +
					"/api/v3/workspaces/test-workspace-123/chat/channels" +
					"?cursor=eyJzdGFydFRpbWUiOiIxNzkxNTc2ODg2MzEyIiwic3RhcnRJZCI6IjJreHY2ZTQ5LTkyIn0" +
					"&limit=100",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of tasks with since until query params",
			Input: common.ReadParams{ObjectName: "tasks", Fields: connectors.Fields("name"),
				Since: time.Unix(1700000000, 0), // 2023-11-14
				Until: time.Unix(1700086400, 0), // 2023-11-15
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v2/team/test-workspace-123/task"),
					mockcond.QueryParam("page", "0"),
					mockcond.QueryParam("date_updated_gt", "1700000000000"),
					mockcond.QueryParam("date_updated_lt", "1700086400000"),
				},
				Then: mockserver.Response(http.StatusOK, responseTasks),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 2,
				NextPage: testconn.URLTestServer + "/api/v2/team/test-workspace-123/task" +
					"?page=1" +
					"&date_updated_gt=1700000000000" +
					"&date_updated_lt=1700086400000",
				Done: false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}

func constructTestConnector(server *httptest.Server) (*Connector, error) {
	conn, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: server.Client(),
		Workspace:           "test-workspace-123",
	})
	if err != nil {
		return nil, err
	}

	conn.SetUnitTestMockServerBaseUrl(server.URL)

	return conn, nil
}
