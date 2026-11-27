package pipedrive

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/pipedrive/internal/crm"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// Search action is dispatched through this interface.
var _ connectors.SearchConnector = &Connector{}

func equalsFilter(pairs ...any) common.SearchFilter {
	filters := make([]common.FieldFilter, 0, len(pairs)/2) // nolint:mnd

	for i := 0; i+1 < len(pairs); i += 2 {
		filters = append(filters, common.FieldFilter{
			FieldName: pairs[i].(string), // nolint:forcetypeassert
			Operator:  common.FilterOperatorEQ,
			Value:     pairs[i+1],
		})
	}

	return common.SearchFilter{FieldFilters: filters}
}

func TestSearchCRM(t *testing.T) { // nolint:funlen,maintidx
	t.Parallel()

	responsePersonsSearch := testutils.DataFromFile(t, "search/persons-search.json")
	responsePersonsSearchEmpty := testutils.DataFromFile(t, "search/persons-search-empty.json")
	responsePersons := testutils.DataFromFile(t, "search/persons.json")
	responseOrganizations := testutils.DataFromFile(t, "search/organizations.json")
	responseEmpty := testutils.DataFromFile(t, "zero-records.json")

	janeDoe := common.ReadResultRow{
		Fields: map[string]any{"id": float64(7), "name": "Jane Doe", "org_id": float64(3)},
		Raw:    map[string]any{"first_name": "Jane", "update_time": "2024-10-16T09:33:33Z"},
		Id:     "7",
	}

	tests := []testconn.TestCaseSearch{
		{
			Name:         "Search object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "At least one field is requested",
			Input: common.SearchParams{
				ObjectName: "persons",
				Filter:     equalsFilter("id", 7),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "A filter is required",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingSearchFilters},
		},
		{
			Name: "Object without lookups cannot be searched",
			Input: common.SearchParams{
				ObjectName: "deals",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("id", 7),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Unknown filter field is rejected",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("name", "Jane"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{crm.ErrUnsupportedSearchField},
		},
		{
			Name: "Organizations cannot be searched by email",
			Input: common.SearchParams{
				ObjectName: "organizations",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("email", "jane@pipedrive.com"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{crm.ErrUnsupportedSearchField},
		},
		{
			Name: "Only the equality operator is supported",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id"),
				Filter: common.SearchFilter{
					FieldFilters: []common.FieldFilter{{
						FieldName: "email",
						Operator:  common.FilterOperator("contains"),
						Value:     "pipedrive.com",
					}},
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{crm.ErrUnsupportedSearchOperator},
		},
		{
			Name: "Email and phone cannot be combined",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("email", "jane@pipedrive.com", "phone", "+372 555555555"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{crm.ErrUnsupportedSearchFilterCombination},
		},
		{
			Name: "Identifier cannot be combined with other fields",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("id", 7, "email", "jane@pipedrive.com"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{crm.ErrUnsupportedSearchFilterCombination},
		},
		{
			Name: "Empty filter value is rejected",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("email", "  "),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{crm.ErrInvalidSearchValue},
		},
		{
			Name: "Person by identifier uses the list endpoint",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id", "name", "org_id"),
				Filter:     equalsFilter("id", float64(7)),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/api/v2/persons"),
					mockcond.QueryParam("ids", "7"),
				},
				Then: mockserver.Response(http.StatusOK, responsePersons),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{janeDoe},
				Done: true,
			},
		},
		{
			Name: "Missing person produces no records",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("id", "999"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/persons"),
					mockcond.QueryParam("ids", "999"),
				},
				Then: mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Expected: &common.SearchResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
		},
		{
			Name: "Persons cannot be searched by organization",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("org_id", 3),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{crm.ErrUnsupportedSearchField},
		},
		{
			Name: "Person by email is searched with exact match and then fetched in full",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id", "name", "org_id"),
				Filter:     equalsFilter("email", "jane@pipedrive.com"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/api/v2/persons/search"),
						mockcond.QueryParam("term", "jane@pipedrive.com"),
						mockcond.QueryParam("fields", "email"),
						mockcond.QueryParam("exact_match", "true"),
						mockcond.QueryParam("limit", "100"),
						mockcond.QueryParamsMissing("organization_id"),
					},
					Then: mockserver.Response(http.StatusOK, responsePersonsSearch),
				}, {
					If: mockcond.And{
						mockcond.Path("/api/v2/persons"),
						mockcond.QueryParam("ids", "7"),
					},
					Then: mockserver.Response(http.StatusOK, responsePersons),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{janeDoe},
				NextPage: testconn.URLTestServer + "/api/v2/persons/search" +
					"?cursor=eyJzZWFyY2giOjF9&exact_match=true&fields=email&limit=100&term=jane%40pipedrive.com",
				Done: false,
			},
		},
		{
			Name: "Person by phone is searched with exact match",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id", "name", "org_id"),
				Filter:     equalsFilter("phone", "+372 555555555"),
				Limit:      10,
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/api/v2/persons/search"),
						mockcond.QueryParam("term", "+372 555555555"),
						mockcond.QueryParam("fields", "phone"),
						mockcond.QueryParam("exact_match", "true"),
						mockcond.QueryParam("limit", "10"),
					},
					Then: mockserver.Response(http.StatusOK, responsePersonsSearch),
				}, {
					If: mockcond.And{
						mockcond.Path("/api/v2/persons"),
						mockcond.QueryParam("ids", "7"),
					},
					Then: mockserver.Response(http.StatusOK, responsePersons),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{janeDoe},
				NextPage: testconn.URLTestServer + "/api/v2/persons/search" +
					"?cursor=eyJzZWFyY2giOjF9&exact_match=true&fields=phone&limit=10" +
					"&term=%2B372+555555555",
				Done: false,
			},
		},
		{
			// The list endpoint is not mocked, calling it would fail the test.
			Name: "No matches skip the list endpoint",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("email", "nobody@pipedrive.com"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/persons/search"),
				Then:  mockserver.Response(http.StatusOK, responsePersonsSearchEmpty),
			}.Server(),
			Expected: &common.SearchResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
		},
		{
			Name: "Next page of a search goes back to the search endpoint",
			Input: common.SearchParams{
				ObjectName: "persons",
				Fields:     connectors.Fields("id", "name", "org_id"),
				Filter:     equalsFilter("email", "jane@pipedrive.com"),
				NextPage: testconn.URLTestServer + "/api/v2/persons/search" +
					"?cursor=abc&exact_match=true&fields=email&limit=100&term=jane%40pipedrive.com",
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/api/v2/persons/search"),
						mockcond.QueryParam("cursor", "abc"),
						mockcond.QueryParam("term", "jane@pipedrive.com"),
					},
					Then: mockserver.Response(http.StatusOK, responsePersonsSearch),
				}, {
					If: mockcond.And{
						mockcond.Path("/api/v2/persons"),
						mockcond.QueryParam("ids", "7"),
					},
					Then: mockserver.Response(http.StatusOK, responsePersons),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{janeDoe},
				NextPage: testconn.URLTestServer + "/api/v2/persons/search" +
					"?cursor=eyJzZWFyY2giOjF9&exact_match=true&fields=email&limit=100&term=jane%40pipedrive.com",
				Done: false,
			},
		},
		{
			Name: "Organization by identifier uses the list endpoint",
			Input: common.SearchParams{
				ObjectName: "organizations",
				Fields:     connectors.Fields("id", "name"),
				Filter:     equalsFilter("id", "3"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/organizations"),
					mockcond.QueryParam("ids", "3"),
				},
				Then: mockserver.Response(http.StatusOK, responseOrganizations),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": float64(3), "name": "Organization name"},
					Raw:    map[string]any{"owner_id": float64(1)},
					Id:     "3",
				}},
				Done: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSearcher, error) {
				return constructTestConnector(tt.Server.URL, providers.ModulePipedriveCRM)
			})
		})
	}
}

func TestSearchLegacyIsNotSupported(t *testing.T) {
	t.Parallel()

	conn, err := constructTestConnector("http://localhost", providers.ModulePipedriveLegacy)
	if err != nil {
		t.Fatalf("failed to construct connector: %v", err)
	}

	_, err = conn.Search(t.Context(), &common.SearchParams{
		ObjectName: "persons",
		Fields:     connectors.Fields("id"),
		Filter:     equalsFilter("id", 7),
	})
	if err == nil {
		t.Fatalf("expected an error for the legacy module")
	}
}
