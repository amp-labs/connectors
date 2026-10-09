package jump

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

const (
	tasksCursor      = "g3QAAAACdwJpZG0AAAAXdHNrX2JkZTFjOGIzMTRlODQ1OTg5M2V3C2luc2VydGVkX2F0dAAAAA13C21pY3Jvc2Vjb25kaAJiAAJEgWEGdwZzZWNvbmRhCXcIY2FsZW5kYXJ3E0VsaXhpci5DYWxlbmRhci5JU093BW1vbnRoYQZ3Cl9fc3RydWN0X193D0VsaXhpci5EYXRlVGltZXcKdXRjX29mZnNldGEAdwpzdGRfb2Zmc2V0YQB3BHllYXJiAAAH6ncEaG91cmEXdwNkYXlhF3cJem9uZV9hYmJybQAAAANVVEN3Bm1pbnV0ZWESdwl0aW1lX3pvbmVtAAAAB0V0Yy9VVEM=" //nolint:lll
	tasksPage2Cursor = "g3QAAAACdwJpZG0AAAAXdHNrX2NmMmVjZjg0Yzc4MzQzNDlhZjF3C2luc2VydGVkX2F0dAAAAA13C21pY3Jvc2Vjb25kaAJiAAI5oGEGdwZzZWNvbmRhCXcIY2FsZW5kYXJ3E0VsaXhpci5DYWxlbmRhci5JU093BW1vbnRoYQZ3Cl9fc3RydWN0X193D0VsaXhpci5EYXRlVGltZXcKdXRjX29mZnNldGEAdwpzdGRfb2Zmc2V0YQB3BHllYXJiAAAH6ncEaG91cmEXdwNkYXlhF3cJem9uZV9hYmJybQAAAANVVEN3Bm1pbnV0ZWESdwl0aW1lX3pvbmVtAAAAB0V0Yy9VVEM=" //nolint:lll
	contactsToken    = `{"insertedAfter":"2026-06-27T23:02:18.773614Z","lastId":"con_db16c04d87d44df18c4"}`
)

func TestRead(t *testing.T) {
	t.Parallel()

	tasksResponse := testutils.DataFromFile(t, "read/tasks.json")
	tasksPage2Response := testutils.DataFromFile(t, "read/tasks-page2.json")
	meetingsResponse := testutils.DataFromFile(t, "read/meetings.json")
	contactsResponse := testutils.DataFromFile(t, "read/contacts.json")
	contactsPage2Response := testutils.DataFromFile(t, "read/contacts-page2.json")
	complexityErrorResponse := testutils.DataFromFile(t, "read/error-complexity.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Successfully read tasks with assignee",
			Input: common.ReadParams{
				ObjectName: "tasks",
				Fields:     connectors.Fields("id", "title", "Assignee"),
				PageSize:   2,
			},
			Server:     graphqlServer(tasksResponse, bodyContains(`first: 2`), bodyContains(`assignee {`)),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    "tsk_9e17843922b94782832",
							"title": "TEST",
						},
						Raw: map[string]any{
							"id":    "tsk_9e17843922b94782832",
							"title": "TEST",
							"assignee": map[string]any{
								"email":  "ada@example.com",
								"name":   "Amperasnd Developer",
								"source": "USER",
							},
						},
						Id: "tsk_9e17843922b94782832",
					},
					{
						Fields: map[string]any{
							"id":    "tsk_bde1c8b314e8459893e",
							"title": "Send follow-up email",
						},
						Raw: map[string]any{
							"id":    "tsk_bde1c8b314e8459893e",
							"title": "Send follow-up email",
						},
						Id: "tsk_bde1c8b314e8459893e",
					},
				},
				NextPage: tasksCursor,
				Done:     false,
			},
		},
		{
			Name: "Read tasks second page using NextPage cursor",
			Input: common.ReadParams{
				ObjectName: "tasks",
				Fields:     connectors.Fields("id", "title"),
				PageSize:   2,
				NextPage:   tasksCursor,
			},
			Server:     graphqlServer(tasksPage2Response, bodyContains(`after: \"`+tasksCursor+`\"`)),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: tasksPage2Cursor,
				Done:     false,
			},
		},
		{
			Name: "Incremental read sends updatedAfter and updatedBefore",
			Input: common.ReadParams{
				ObjectName: "tasks",
				Fields:     connectors.Fields("id"),
				Since:      time.Date(2026, time.June, 24, 0, 0, 0, 148609000, time.UTC),
				Until:      time.Date(2026, time.July, 1, 0, 0, 0, 96167000, time.UTC),
			},
			Server: graphqlServer(tasksPage2Response,
				bodyContains(`updatedAfter: \"2026-06-24T00:00:00.148609Z\"`),
				bodyContains(`updatedBefore: \"2026-07-01T00:00:00.096167Z\"`),
			),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: tasksPage2Cursor,
				Done:     false,
			},
		},
		{
			Name: "Meetings page size stays within the complexity limit",
			Input: common.ReadParams{
				ObjectName: "meetings",
				Fields:     connectors.Fields("id", "status", "source", "startedAt"),
			},
			Server:     graphqlServer(meetingsResponse, bodyContains(`first: 27`)),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":        "mtg_2df34b22211d4f248f8",
							"status":    "FAILED",
							"source":    "WEB_RECORDER",
							"startedat": "2026-06-27T23:06:44.246638Z",
						},
						Raw: map[string]any{
							"id":        "mtg_2df34b22211d4f248f8",
							"status":    "FAILED",
							"source":    "WEB_RECORDER",
							"startedAt": "2026-06-27T23:06:44.246638Z",
						},
						Id: "mtg_2df34b22211d4f248f8",
					},
					{
						Fields: map[string]any{
							"id":        "mtg_fbff3264957a48b4888",
							"status":    "COMPLETED",
							"source":    nil,
							"startedat": "2026-06-23T23:18:09.022872Z",
						},
						Raw: map[string]any{
							"id":        "mtg_fbff3264957a48b4888",
							"status":    "COMPLETED",
							"source":    nil,
							"startedAt": "2026-06-23T23:18:09.022872Z",
						},
						Id: "mtg_fbff3264957a48b4888",
					},
				},
				Done: true,
			},
		},
		{
			Name:   "Query complexity error is returned as bad request",
			Input:  common.ReadParams{ObjectName: "meetings", Fields: connectors.Fields("id")},
			Server: graphqlServer(complexityErrorResponse),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Field meetings is too complex: complexity is 3700 and maximum is 1000"),
			},
		},
		{
			Name:  "Invalid token is returned as access token error",
			Input: common.ReadParams{ObjectName: "tasks", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized,
					[]byte(`{"error":{"code":"TOKEN_INVALID","message":"Invalid or expired API token"}}`)),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
				testutils.StringError("Invalid or expired API token"),
			},
		},
		{
			Name: "Contacts first page produces an insertedAfter page token",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
				PageSize:   1,
			},
			Server:     graphqlServer(contactsResponse, bodyContains(`first: 1`)),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "con_db16c04d87d44df18c4",
						"name": "Ada Lovelace",
					},
					Raw: map[string]any{
						"id":         "con_db16c04d87d44df18c4",
						"insertedAt": "2026-06-27T23:02:18.773614Z",
					},
					Id: "con_db16c04d87d44df18c4",
				}},
				NextPage: contactsToken,
				Done:     false,
			},
		},
		{
			Name: "Contacts next page uses insertedAfter and drops the boundary record",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
				PageSize:   1,
				NextPage:   contactsToken,
			},
			Server: graphqlServer(contactsPage2Response,
				bodyContains(`first: 2`),
				bodyContains(`insertedAfter: \"2026-06-27T23:02:18.773614Z\"`),
			),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "con_e0640d697bad4a2fb20",
						"name": "john doe",
					},
					Raw: map[string]any{
						"id": "con_e0640d697bad4a2fb20",
					},
					Id: "con_e0640d697bad4a2fb20",
				}},
				Done: true,
			},
		},
		{
			Name: "Contacts next page stays within the complexity limit",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "contactInfo", "integrationReferences"),
				NextPage:   contactsToken,
			},
			Server:     graphqlServer(contactsPage2Response, bodyContains(`first: 58`)),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Done: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// Every list query is a POST to the single GraphQL endpoint.
func graphqlServer(response []byte, conditions ...mockcond.Condition) *httptest.Server {
	return mockserver.Conditional{
		Setup: mockserver.ContentJSON(),
		If:    append(mockcond.And{mockcond.MethodPOST()}, conditions...),
		Then:  mockserver.Response(http.StatusOK, response),
	}.Server()
}

// bodyContains matches requests whose raw body contains the substring.
// GraphQL queries arrive JSON-encoded, so quotes appear escaped (\").
func bodyContains(substring string) mockcond.Check {
	return func(w http.ResponseWriter, r *http.Request) bool {
		return strings.Contains(readBody(r), substring)
	}
}

func readBody(r *http.Request) string {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}

	_ = r.Body.Close()
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return string(body)
}
