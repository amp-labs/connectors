package teamwork

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

	responseNotFoundError := testutils.DataFromFile(t, "read/object-not-found.json")
	responseCompaniesFirstPage := testutils.DataFromFile(t, "read/companies/1-first-page.json")
	responseCompaniesLastPage := testutils.DataFromFile(t, "read/companies/2-last-page.json")
	responseNotebooksFirstPage := testutils.DataFromFile(t, "read/notebooks/1-first-page.json")
	responseNotebooksSecondPage := testutils.DataFromFile(t, "read/notebooks/2-second-page.json")

	tests := []testroutines.Read{
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "companies", Fields: connectors.Fields("name")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFoundError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Not Found"),
			},
		},
		{
			Name: "Read companies first page incrementally",
			Input: common.ReadParams{
				ObjectName: "companies",
				Fields:     connectors.Fields("name", "city"),
				Since: time.Date(2024, 9, 19, 4, 30, 45, 600,
					time.FixedZone("UTC-8", -8*60*60)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/projects/api/v3/companies.json"),
					mockcond.QueryParam("updatedAfter", "2024-09-19T12:30:45Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseCompaniesFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Ampersand",
						"city": "San Francisco",
					},
					Raw: map[string]any{
						"id": float64(1387405),
					},
				}, {
					Fields: map[string]any{
						"name": "Nike",
						"city": "Chicago",
					},
					Raw: map[string]any{
						"id": float64(1412778),
					},
				}},
				NextPage: testroutines.URLTestServer + "/projects/api/v3/companies.json?" +
					"page=2&pageSize=500&" +
					"updatedAfter=2024-09-19T12:30:45Z",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read companies second page which is without the next cursor",
			Input: common.ReadParams{
				ObjectName: "companies",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/projects/api/v3/companies.json"),
				Then:  mockserver.Response(http.StatusOK, responseCompaniesLastPage),
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
			Name: "Read notebooks first page",
			Input: common.ReadParams{
				ObjectName: "notebooks",
				Fields:     connectors.Fields("name"),
				PageSize:   1,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/projects/api/v3/notebooks.json"),
					mockcond.QueryParam("pageSize", "1"),
				},
				Then: mockserver.Response(http.StatusOK, responseNotebooksFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"name": "Second notebook"},
					Raw:    map[string]any{"id": float64(305625)},
				}},
				NextPage: testroutines.URLTestServer + "/projects/api/v3/notebooks.json?page=2&pageSize=1",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read notebooks last page",
			Input: common.ReadParams{
				ObjectName: "notebooks",
				Fields:     connectors.Fields("name"),
				NextPage:   testroutines.URLTestServer + "/projects/api/v3/notebooks.json?page=2&pageSize=1",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/projects/api/v3/notebooks.json"),
					mockcond.QueryParam("pageSize", "1"),
					mockcond.QueryParam("page", "2"),
				},
				Then: mockserver.Response(http.StatusOK, responseNotebooksSecondPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"name": "First notebook"},
					Raw:    map[string]any{"id": float64(305624)},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
