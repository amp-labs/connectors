package sendgrid

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

	responseLists := testutils.DataFromFile(t, "read/lists.json")
	responseListsPage1 := testutils.DataFromFile(t, "read/lists-page1.json")
	responseListsEmpty := testutils.DataFromFile(t, "read/lists-empty.json")
	responseBounces := testutils.DataFromFile(t, "read/bounces.json")
	responseEmptyArray := testutils.DataFromFile(t, "read/empty-array.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: objectLists},
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
			Name:  "Zero records response for lists",
			Input: common.ReadParams{ObjectName: objectLists, Fields: connectors.Fields("id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/marketing/lists"),
					mockcond.QueryParam("page_size", defaultPageSize),
				},
				Then: mockserver.Response(http.StatusOK, responseListsEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read lists",
			Input: common.ReadParams{ObjectName: objectLists, Fields: connectors.Fields("id", "name")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/marketing/lists"),
					mockcond.QueryParam("page_size", defaultPageSize),
				},
				Then: mockserver.Response(http.StatusOK, responseLists),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "list-abc-123",
						"name": "Newsletter Subscribers",
					},
					Raw: map[string]any{
						"id":            "list-abc-123",
						"name":          "Newsletter Subscribers",
						"contact_count": float64(42),
					},
					Id: "list-abc-123",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read lists with next page from _metadata.next",
			Input: common.ReadParams{ObjectName: objectLists, Fields: connectors.Fields("id"), PageSize: 1},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/marketing/lists"),
					mockcond.QueryParam("page_size", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseListsPage1),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": "list-abc-123",
					},
					Raw: map[string]any{
						"id":            "list-abc-123",
						"name":          "Newsletter Subscribers",
						"contact_count": float64(42),
					},
					Id: "list-abc-123",
				}},
				NextPage: "https://api.sendgrid.com/v3/marketing/lists?page_size=1&page_token=token-page-2",
				Done:     false,
			},
		},
		{
			Name:  "Read bounces root array",
			Input: common.ReadParams{ObjectName: objectBounces, Fields: connectors.Fields("email", "reason")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/suppression/bounces"),
					mockcond.QueryParam("limit", defaultPageSize),
					mockcond.QueryParam("offset", "0"),
				},
				Then: mockserver.Response(http.StatusOK, responseBounces),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"email":  "bounce@example.com",
						"reason": "550 5.1.1 The email account that you tried to reach does not exist.",
					},
					Raw: map[string]any{
						"created": float64(1443651141),
						"email":   "bounce@example.com",
						"reason":  "550 5.1.1 The email account that you tried to reach does not exist.",
						"status":  "5.1.1",
					},
					Id: "bounce@example.com",
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Read bounces with next page via offset",
			Input: common.ReadParams{ObjectName: objectBounces, Fields: connectors.Fields("email"), PageSize: 2},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/suppression/bounces"),
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParam("offset", "0"),
				},
				Then: mockserver.Response(http.StatusOK, responseBounces),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"email": "bounce@example.com",
					},
					Raw: map[string]any{
						"created": float64(1443651141),
						"email":   "bounce@example.com",
						"reason":  "550 5.1.1 The email account that you tried to reach does not exist.",
						"status":  "5.1.1",
					},
					Id: "bounce@example.com",
				}},
				NextPage: testconn.URLTestServer + "/v3/suppression/bounces?limit=2&offset=2",
				Done:     false,
			},
		},
		{
			Name:  "Zero records response for bounces",
			Input: common.ReadParams{ObjectName: objectBounces, Fields: connectors.Fields("email")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/suppression/bounces"),
				},
				Then: mockserver.Response(http.StatusOK, responseEmptyArray),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
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
