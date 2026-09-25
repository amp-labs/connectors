package bamboohr

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseEmployeesDirectory := testutils.DataFromFile(t, "read/employees-directory.json")
	responseMetaUsers := testutils.DataFromFile(t, "read/meta-users.json")
	responseTimeOffRequestsPage1 := testutils.DataFromFile(t, "read/time-off-requests-page1.json")
	responseTimeOffRequestsPage2 := testutils.DataFromFile(t, "read/time-off-requests-page2.json")
	responseWhosOut := testutils.DataFromFile(t, "read/whos-out.json")
	responseEmployeesPage1 := testutils.DataFromFile(t, "read/employees-page1.json")
	responseEmptyData := testutils.DataFromFile(t, "read/empty-data.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectEmployeesDirectory},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "unknown", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: objectTimeOffRequests, Fields: connectors.Fields("id", "status")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v1/time-off/requests"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("pageSize", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseEmptyData),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read employees directory",
			Input: common.ReadParams{ObjectName: objectEmployeesDirectory, Fields: connectors.Fields("id", "firstName", "lastName")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v1/employees/directory"),
				},
				Then: mockserver.Response(http.StatusOK, responseEmployeesDirectory),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":        "123",
						"firstname": "Ava",
						"lastname":  "Nguyen",
					},
					Raw: map[string]any{
						"id":        "123",
						"firstName": "Ava",
						"lastName":  "Nguyen",
						"jobTitle":  "Engineering Manager",
					},
					Id: "123",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read meta users from keyed object response",
			Input: common.ReadParams{ObjectName: objectMetaUsers, Fields: connectors.Fields("id", "email", "status")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v1/meta/users"),
				},
				Then: mockserver.Response(http.StatusOK, responseMetaUsers),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":     float64(1001),
						"email":  "ava@example.com",
						"status": "enabled",
					},
					Raw: map[string]any{
						"id":         float64(1001),
						"employeeId": float64(123),
						"email":      "ava@example.com",
					},
					Id: "1001",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read time off requests first page",
			Input: common.ReadParams{
				ObjectName: objectTimeOffRequests,
				Fields:     connectors.Fields("id", "employeeId", "status"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v1/time-off/requests"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("pageSize", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseTimeOffRequestsPage1),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":         float64(501),
						"employeeid": float64(123),
						"status":     "approved",
					},
					Raw: map[string]any{
						"id":        float64(501),
						"startDate": "2026-04-01",
					},
					Id: "501",
				}},
				NextPage: "https://example.bamboohr.com/api/v1/time-off/requests?page=2&pageSize=1",
				Done:     false,
			},
		},
		{
			Name: "Read time off requests second page using NextPage",
			Input: common.ReadParams{
				ObjectName: objectTimeOffRequests,
				Fields:     connectors.Fields("id", "status"),
				NextPage:   testconn.URLTestServer + "/api/v1/time-off/requests?page=2&pageSize=1",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v1/time-off/requests"),
					mockcond.QueryParam("page", "2"),
					mockcond.QueryParam("pageSize", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseTimeOffRequestsPage2),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":     float64(502),
						"status": "requested",
					},
					Raw: map[string]any{
						"id":        float64(502),
						"startDate": "2026-05-10",
					},
					Id: "502",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read who is out",
			Input: common.ReadParams{ObjectName: objectWhosOut, Fields: connectors.Fields("id", "type", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v1/whos-out"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("pageSize", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseWhosOut),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "7001",
						"type": "timeOff",
						"name": "Ava Nguyen",
					},
					Raw: map[string]any{
						"id":         "7001",
						"employeeId": float64(123),
						"name":       "Ava Nguyen",
					},
					Id: "7001",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read employees cursor page",
			Input: common.ReadParams{
				ObjectName: objectEmployees,
				Fields:     connectors.Fields("employeeId", "firstName", "status"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v1/employees"),
					mockcond.QueryParam("page[limit]", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseEmployeesPage1),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"employeeid": "123",
						"firstname":  "Ava",
						"status":     "Active",
					},
					Raw: map[string]any{
						"employeeId": "123",
						"firstName":  "Ava",
					},
					Id: "123",
				}},
				NextPage: "https://example.bamboohr.com/api/v1/employees?page%5Blimit%5D=1&page%5Bafter%5D=next-cursor",
				Done:     false,
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
