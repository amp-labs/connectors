package mail

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

// testAccountID is the account id the test adapter is built with. Only
// account-scoped objects (e.g. messages) use it; others ignore it.
const testAccountID = "acc123"

func TestRead(t *testing.T) { // nolint:funlen
	t.Parallel()

	accountsResponse := testutils.DataFromFile(t, "accounts.json")
	notesResponse := testutils.DataFromFile(t, "notes.json")
	messagesResponse := testutils.DataFromFile(t, "messages.json")

	tests := []testroutines.Read{
		{
			Name:         "Object name and fields are required",
			Input:        common.ReadParams{ObjectName: "notes"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unsupported object",
			Input:        common.ReadParams{ObjectName: "folders", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:  "Read accounts object",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("accountId")},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/accounts"),
					Then: mockserver.Response(http.StatusOK, accountsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "2554907000000008002",
					Fields: map[string]any{"accountid": "2554907000000008002"},
					Raw: map[string]any{
						"accountId":           "2554907000000008002",
						"primaryEmailAddress": "john@zylker.com",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read notes with pagination",
			Input: common.ReadParams{
				ObjectName: "notes",
				Fields:     connectors.Fields("entityId"),
				PageSize:   1,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/api/notes/me"),
						mockcond.QueryParam("after", "1"),
						mockcond.QueryParam("limit", "1"),
					},
					Then: mockserver.Response(http.StatusOK, notesResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     "1781549700389120100",
					Fields: map[string]any{"entityid": "1781549700389120100"},
					Raw:    map[string]any{"entityId": "1781549700389120100"},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read tasks with pagination",
			Input: common.ReadParams{
				ObjectName: "tasks",
				Fields:     connectors.Fields("id"),
				PageSize:   2,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/api/tasks/me"),
						mockcond.QueryParam("from", "0"),
						mockcond.QueryParam("limit", "2"),
					},
					Then: mockserver.ResponseString(http.StatusOK, `{
						"status": {"code": 200, "description": "success"},
						"data": {
							"paging": {"nextPage": "tasks/me?from=2&limit=2"},
							"tasks": [{"id": "t1"}, {"id": "t2"}]
						}
					}`),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id:     "t1",
					Fields: map[string]any{"id": "t1"},
					Raw:    map[string]any{"id": "t1"},
				}},
				NextPage: testroutines.URLTestServer + "/api/tasks/me?from=2&limit=2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read messages with ",
			Input: common.ReadParams{
				ObjectName: "messages",
				Fields:     connectors.Fields("messageId", "subject"),
				PageSize:   1,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/api/accounts/" + testAccountID + "/messages/view"),
						mockcond.QueryParam("start", "1"),
						mockcond.QueryParam("limit", "1"),
					},
					Then: mockserver.Response(http.StatusOK, messagesResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"messageid": "1709876543210000001",
						"subject":   "Welcome to Zoho Mail",
					},
					Raw: map[string]any{
						"messageId":   "1709876543210000001",
						"subject":     "Welcome to Zoho Mail",
						"fromAddress": "alice@zylker.com",
					},
				}},
				NextPage: testroutines.URLTestServer + "/api/accounts/" + testAccountID + "/messages/view?start=2&limit=1",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tc := testroutines.TestCase[common.ReadParams, *common.ReadResult](tt)
			t.Cleanup(tc.Close)

			adapter := constructTestAdapter(t, tt.Server.URL, testAccountID)

			output, err := adapter.Read(t.Context(), tc.Input)
			tc.Validate(t, err, output)
		})
	}
}
