package procore

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

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	projectsResponse := testutils.DataFromFile(t, "projects.json")
	operationsResponse := testutils.DataFromFile(t, "operations.json")

	since := time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)
	until := time.Date(2024, 10, 1, 23, 59, 59, 0, time.UTC)

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "projects"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Read projects with default pagination",
			Input: common.ReadParams{
				ObjectName: "company/projects",
				Fields:     connectors.Fields("id", "name", "active"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/v1.0/companies/" + testCompanyID + "/projects"),
					mockcond.QueryParam("page", "1"),
					mockcond.QueryParam("per_page", "1000"),
					mockcond.Header(http.Header{"Procore-Company-Id": []string{testCompanyID}}),
				},
				Then: mockserver.Response(http.StatusOK, projectsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":     float64(2783405),
							"name":   "Test Project Alpha",
							"active": true,
						},
						Raw: map[string]any{
							"id":             float64(2783405),
							"project_number": "PN-001",
						},
					},
					{
						Fields: map[string]any{
							"id":     float64(2783406),
							"name":   "Test Project Bravo",
							"active": false,
						},
						Raw: map[string]any{
							"id":             float64(2783406),
							"project_number": "PN-002",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental read of projects sends filters[updated_at]",
			Input: common.ReadParams{
				ObjectName: "company/projects",
				Fields:     connectors.Fields("id", "name"),
				Since:      since,
				Until:      until,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/v1.0/companies/" + testCompanyID + "/projects"),
					mockcond.QueryParam("filters[updated_at]", "2024-10-01T00:00:00Z...2024-10-01T23:59:59Z"),
				},
				Then: mockserver.Response(http.StatusOK, projectsResponse),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 2, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Link header advances NextPage with the full URL",
			Input: common.ReadParams{
				ObjectName: "company/projects",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link", `<https://api.procore.com/rest/v1.0/companies/`+testCompanyID+`/projects?page=2&per_page=100>; rel="next"`),
					mockserver.Response(http.StatusOK, projectsResponse),
				),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: "https://api.procore.com/rest/v1.0/companies/" + testCompanyID + "/projects?page=2&per_page=100",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "NextPage URL is used verbatim for the follow-up request",
			Input: common.ReadParams{
				ObjectName: "company/projects",
				Fields:     connectors.Fields("id"),
				NextPage:   common.NextPageToken(testroutines.URLTestServer + "/rest/v1.0/companies/" + testCompanyID + "/projects?page=3&per_page=100"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/v1.0/companies/" + testCompanyID + "/projects"),
					mockcond.QueryParam("page", "3"),
					mockcond.QueryParam("per_page", "100"),
				},
				Then: mockserver.Response(http.StatusOK, projectsResponse),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 2, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental filter window is preserved across pages",
			Input: common.ReadParams{
				ObjectName: "company/projects",
				Fields:     connectors.Fields("id"),
				Since:      since,
				Until:      until,
				NextPage: common.NextPageToken(testroutines.URLTestServer +
					"/rest/v1.0/companies/" + testCompanyID + "/projects" +
					"?filters%5Bupdated_at%5D=2024-10-01T00%3A00%3A00Z...2024-10-01T23%3A59%3A59Z" +
					"&page=2&per_page=100"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/v1.0/companies/" + testCompanyID + "/projects"),
					mockcond.QueryParam("page", "2"),
					mockcond.QueryParam("per_page", "100"),
					mockcond.QueryParam("filters[updated_at]", "2024-10-01T00:00:00Z...2024-10-01T23:59:59Z"),
				},
				Then: mockserver.Response(http.StatusOK, projectsResponse),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 2, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "v2.0 endpoint unwraps data array",
			Input: common.ReadParams{
				ObjectName: "operations",
				Fields:     connectors.Fields("id", "status"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/rest/v2.0/companies/" + testCompanyID + "/async_operations"),
				Then:  mockserver.Response(http.StatusOK, operationsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{"id": "op_abc123", "status": "succeeded"},
						Raw:    map[string]any{"operation_type": "bulk_update"},
					},
					{
						Fields: map[string]any{"id": "op_def456", "status": "pending"},
						Raw:    map[string]any{"operation_type": "bulk_create"},
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		//nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
