package wrike

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseTasksPage1 := testutils.DataFromFile(t, "tasks-page1.json")
	responseTasksPage2 := testutils.DataFromFile(t, "tasks-page2.json")
	responseTasksDescription := testutils.DataFromFile(t, "tasks-with-description.json")
	responseTasksEmpty := testutils.DataFromFile(t, "tasks-empty.json")
	responseContacts := testutils.DataFromFile(t, "contacts.json")
	responseTimesheets := testutils.DataFromFile(t, "timesheets.json")
	responseGroupsPage1 := testutils.DataFromFile(t, "groups-page1.json")
	responseUnauthorized := testutils.DataFromFile(t, "error-unauthorized.json")

	since := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2026, 10, 7, 12, 30, 0, 0, time.UTC)

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectTasks},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "account", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: objectTasks, Fields: connectors.Fields("id", "title")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v4/tasks"),
					mockcond.QueryParam(paramPageSize, "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseTasksEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "First page of tasks carries the next page token",
			Input: common.ReadParams{
				ObjectName: objectTasks,
				Fields:     connectors.Fields("id", "title", "status"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v4/tasks"),
					mockcond.QueryParam(paramPageSize, "2"),
					mockcond.QueryParamsMissing(paramNextPageToken, paramUpdatedDate, paramFields),
				},
				Then: mockserver.Response(http.StatusOK, responseTasksPage1),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":     "IEAG5DBKKQ3U5ZXA",
						"title":  "Write the connector",
						"status": "Active",
					},
					Raw: map[string]any{
						"id":         "IEAG5DBKKQ3U5ZXA",
						"importance": "High",
					},
					Id: "IEAG5DBKKQ3U5ZXA",
				}},
				NextPage: testconn.URLTestServer + "/api/v4/tasks?nextPageToken=TOKEN-PAGE-2&pageSize=2",
				Done:     false,
			},
		},
		{
			Name: "Last page of tasks via NextPage has no token",
			Input: common.ReadParams{
				ObjectName: objectTasks,
				Fields:     connectors.Fields("id", "title"),
				NextPage:   testconn.URLTestServer + "/api/v4/tasks?nextPageToken=TOKEN-PAGE-2&pageSize=2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v4/tasks"),
					mockcond.QueryParam(paramNextPageToken, "TOKEN-PAGE-2"),
					mockcond.QueryParam(paramPageSize, "2"),
				},
				Then: mockserver.Response(http.StatusOK, responseTasksPage2),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "IEAG5DBKKQ3U5ZXC", "title": "Ship the connector"},
					Raw:    map[string]any{"id": "IEAG5DBKKQ3U5ZXC", "status": "Completed"},
					Id:     "IEAG5DBKKQ3U5ZXC",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Since and Until become the updatedDate range",
			Input: common.ReadParams{
				ObjectName: objectTasks,
				Fields:     connectors.Fields("id"),
				Since:      since,
				Until:      until,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v4/tasks"),
					mockcond.QueryParam(paramUpdatedDate, `{"start":"2026-10-01T00:00:00Z","end":"2026-10-07T12:30:00Z"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseTasksPage2),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "IEAG5DBKKQ3U5ZXC"},
					Raw:    map[string]any{"id": "IEAG5DBKKQ3U5ZXC"},
					Id:     "IEAG5DBKKQ3U5ZXC",
				}},
				Done: true,
			},
		},
		{
			Name: "Since alone becomes an open-ended range",
			Input: common.ReadParams{
				ObjectName: objectTasks,
				Fields:     connectors.Fields("id"),
				Since:      since,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.QueryParam(paramUpdatedDate, `{"start":"2026-10-01T00:00:00Z"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseTasksEmpty),
			}.Server(),
			Expected: &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
		},
		{
			Name: "Requested optional fields are sent in the fields parameter",
			Input: common.ReadParams{
				ObjectName: objectTasks,
				Fields:     connectors.Fields("id", "title", "description", "attachmentCount"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v4/tasks"),
					mockcond.QueryParam(paramFields, `["attachmentCount","description"]`),
				},
				Then: mockserver.Response(http.StatusOK, responseTasksDescription),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":              "IEAG5DBKKQ3U5ZXA",
						"title":           "Write the connector",
						"description":     "<p>Static metadata first</p>",
						"attachmentcount": float64(2),
					},
					Raw: map[string]any{"id": "IEAG5DBKKQ3U5ZXA", "description": "<p>Static metadata first</p>"},
					Id:  "IEAG5DBKKQ3U5ZXA",
				}},
				Done: true,
			},
		},
		{
			Name: "Contacts are a single response without paging or filters",
			Input: common.ReadParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id", "firstName", "type"),
				Since:      since,
				PageSize:   50,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v4/contacts"),
					mockcond.QueryParamsMissing(paramPageSize, paramUpdatedDate, paramFields),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "KUAAAAAA", "firstname": "Joseph", "type": "Person"},
						Raw:    map[string]any{"id": "KUAAAAAA", "me": true},
						Id:     "KUAAAAAA",
					},
					{
						Fields: map[string]any{"id": "KUAAAAAB", "firstname": "Engineering", "type": "Group"},
						Raw:    map[string]any{"id": "KUAAAAAB"},
						Id:     "KUAAAAAB",
					},
				},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Timesheets use timesheetId as the record id",
			Input: common.ReadParams{ObjectName: objectTimesheets, Fields: connectors.Fields("timesheetId", "timeframe")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.And{mockcond.MethodGET(), mockcond.Path("/api/v4/timesheets")},
				Then:  mockserver.Response(http.StatusOK, responseTimesheets),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"timesheetid": "TS-1", "timeframe": "Week"},
					Raw:    map[string]any{"timesheetId": "TS-1", "userId": "KUAAAAAA"},
					Id:     "TS-1",
				}},
				Done: true,
			},
		},
		{
			Name: "Groups page with pageToken",
			Input: common.ReadParams{
				ObjectName: objectGroups,
				Fields:     connectors.Fields("id", "title"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v4/groups"),
					mockcond.QueryParam(paramPageSize, "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseGroupsPage1),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "KX77AAAA", "title": "Engineering"},
					Raw:    map[string]any{"id": "KX77AAAA", "myTeam": true},
					Id:     "KX77AAAA",
				}},
				NextPage: testconn.URLTestServer + "/api/v4/groups?pageSize=1&pageToken=GROUP-TOKEN-2",
				Done:     false,
			},
		},
		{
			Name:  "Invalid token surfaces as access token error",
			Input: common.ReadParams{ObjectName: objectTasks, Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, responseUnauthorized),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
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
