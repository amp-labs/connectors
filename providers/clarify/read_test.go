package clarify

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

const testWorkspace = "test-ws"

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseNotFound := testutils.DataFromFile(t, "error-workspace-not-found.json")
	responseCompanyFirstPage := testutils.DataFromFile(t, "company-first-page.json")
	responseCompanyEmpty := testutils.DataFromFile(t, "company-empty.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "company"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "comments", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "company", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrNotFound,
				testutils.StringError("Workspace not found"),
			},
		},
		{
			Name: "First page of companies is flattened and continues at the next offset",
			Input: common.ReadParams{
				ObjectName: "company",
				Fields:     connectors.Fields("name", "employee_range", "domains", "_updated_at"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/workspaces/test-ws/objects/company/resources"),
					mockcond.QueryParam("page[limit]", "2"),
					mockcond.QueryParam("page[offset]", "0"),
				},
				Then: mockserver.Response(http.StatusOK, responseCompanyFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id: "051cfd45-b10b-44e5-8e6d-8c4a25012d97",
					Fields: map[string]any{
						"name":           "Test Co 09",
						"employee_range": "11-50",
						"domains":        map[string]any{"items": []any{"testco09.example.com"}},
						"_updated_at":    "2026-10-22T06:41:04.124Z",
					},
					Raw: map[string]any{"type": "company"},
				}, {
					Id: "0adf6280-c5d7-4b64-9de2-db7bc81ba0e5",
					Fields: map[string]any{
						"name":           "Ampersand",
						"employee_range": "11-50",
						"domains":        map[string]any{"items": []any{"withampersand.com"}},
						"_updated_at":    "2026-10-22T19:20:32.797Z",
					},
					Raw: map[string]any{"type": "company"},
				}},
				// links.next is http://api.clarify.ai/...; the connector pages on its own base URL.
				NextPage: testconn.URLTestServer +
					"/v1/workspaces/test-ws/objects/company/resources?page%5Blimit%5D=2&page%5Boffset%5D=2",
				Done: false,
			},
		},
		{
			Name: "Page without links.next is the last one",
			Input: common.ReadParams{
				ObjectName: "company",
				Fields:     connectors.Fields("name"),
				NextPage: testconn.URLTestServer +
					"/v1/workspaces/test-ws/objects/company/resources?page%5Blimit%5D=2&page%5Boffset%5D=2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/workspaces/test-ws/objects/company/resources"),
					mockcond.QueryParam("page[offset]", "2"),
				},
				Then: mockserver.Response(http.StatusOK, responseCompanyEmpty),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: "", Done: true},
		},
		{
			Name: "Incremental read sends _updated_at bounds moved to the next millisecond and keeps them on the next page",
			Input: common.ReadParams{
				ObjectName: "company",
				Fields:     connectors.Fields("name"),
				PageSize:   2,
				Since:      time.Date(2026, 10, 1, 0, 0, 0, 123_000_000, time.UTC),
				Until:      time.Date(2026, 10, 23, 0, 0, 0, 456_000_000, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/workspaces/test-ws/objects/company/resources"),
					mockcond.QueryParam("filter[_updated_at][Is on or after]", "2026-10-01T00:00:00.124Z"),
					mockcond.QueryParam("filter[_updated_at][Is before]", "2026-10-23T00:00:00.457Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseCompanyFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 2,
				NextPage: testconn.URLTestServer + "/v1/workspaces/test-ws/objects/company/resources?" +
					"filter%5B_updated_at%5D%5BIs%20before%5D=2026-10-23T00%3A00%3A00.457Z&" +
					"filter%5B_updated_at%5D%5BIs%20on%20or%20after%5D=2026-10-01T00%3A00%3A00.124Z&" +
					"page%5Blimit%5D=2&page%5Boffset%5D=2",
				Done: false,
			},
		},
		{
			Name: "Custom objects are read from the same endpoint",
			Input: common.ReadParams{
				ObjectName: "c_sample_contract",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/workspaces/test-ws/objects/c_sample_contract/resources"),
					mockcond.QueryParam("page[limit]", "500"),
				},
				Then: mockserver.Response(http.StatusOK, responseCompanyEmpty),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: "", Done: true},
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           testWorkspace,
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseUrl(serverURL)

	return connector, nil
}
