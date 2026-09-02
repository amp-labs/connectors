package monaco

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

// Monaco must satisfy the SearchConnector interface, not merely have a method
// with the right shape -- that interface is how the action is dispatched.
var _ connectors.SearchConnector = &Connector{}

func equalsFilter(field string, value any) common.SearchFilter {
	return common.SearchFilter{
		FieldFilters: []common.FieldFilter{{
			FieldName: field,
			Operator:  common.FilterOperatorEQ,
			Value:     value,
		}},
	}
}

func TestSearch(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseContacts := testutils.DataFromFile(t, "read/contacts.json")
	responseContactsEmpty := testutils.DataFromFile(t, "read/contacts-empty.json")

	tests := []testconn.TestCaseSearch{
		{
			Name:         "Search object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "At least one field is requested",
			Input: common.SearchParams{
				ObjectName: objectContacts,
				Filter:     equalsFilter("email", "jane@acme.com"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "A filter is required",
			Input: common.SearchParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingSearchFilters},
		},
		{
			Name: "Audiences cannot be searched, its list request has no filters",
			Input: common.SearchParams{
				ObjectName: objectAudiences,
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("name", "Q3"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Tags cannot be searched, it is a plain GET collection",
			Input: common.SearchParams{
				ObjectName: objectTags,
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("name", "Hot"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Equality filter becomes an equals condition",
			Input: common.SearchParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id", "email"),
				Filter:     equalsFilter("email", "jane@acme.com"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					// Same endpoint Read uses; only `filters` differs.
					mockcond.Path("/v1/contacts/list"),
					mockcond.Body(`{
						"page": 1,
						"page_size": 100,
						"filters": [
							{"field":"email","condition":"equals","value":"jane@acme.com"}
						]
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContacts),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "con_abc123", "email": "jane@acme.com"},
					Raw:    map[string]any{"first_name": "Jane"},
					Id:     "con_abc123",
				}, {
					Fields: map[string]any{"id": "con_xyz789", "email": "john@globex.com"},
					Raw:    map[string]any{"first_name": "John"},
					Id:     "con_xyz789",
				}},
				NextPage: "2",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Multiple filters are sent as a flat array, AND-joined by Monaco",
			Input: common.SearchParams{
				ObjectName: objectOpportunities,
				Fields:     connectors.Fields("id"),
				Filter: common.SearchFilter{
					FieldFilters: []common.FieldFilter{{
						FieldName: "name",
						Operator:  common.FilterOperatorEQ,
						Value:     "Acme renewal",
					}, {
						FieldName: "stage",
						Operator:  common.FilterOperatorEQ,
						Value:     "negotiation",
					}},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v1/opportunities/list"),
					mockcond.Body(`{
						"page": 1,
						"page_size": 100,
						"filters": [
							{"field":"name","condition":"equals","value":"Acme renewal"},
							{"field":"stage","condition":"equals","value":"negotiation"}
						]
					}`),
				},
				Then: mockserver.Response(http.StatusOK, responseContactsEmpty),
			}.Server(),
			Expected:     &common.SearchResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Limit maps to page_size and is clamped at 500",
			Input: common.SearchParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("email", "jane@acme.com"),
				Limit:      5000,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(`{
					"page": 1,
					"page_size": 500,
					"filters": [
						{"field":"email","condition":"equals","value":"jane@acme.com"}
					]
				}`),
				Then: mockserver.Response(http.StatusOK, responseContactsEmpty),
			}.Server(),
			Expected:     &common.SearchResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "NextPage carries the page number, as in Read",
			Input: common.SearchParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("email", "jane@acme.com"),
				NextPage:   "3",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(`{
					"page": 3,
					"page_size": 100,
					"filters": [
						{"field":"email","condition":"equals","value":"jane@acme.com"}
					]
				}`),
				Then: mockserver.Response(http.StatusOK, responseContactsEmpty),
			}.Server(),
			Expected:     &common.SearchResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Malformed NextPage is rejected",
			Input: common.SearchParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("email", "jane@acme.com"),
				NextPage:   "not-a-page",
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrInvalidNextPage},
		},
		{
			Name: "Unknown operator is rejected before any request",
			Input: common.SearchParams{
				ObjectName: objectContacts,
				Fields:     connectors.Fields("id"),
				Filter: common.SearchFilter{
					FieldFilters: []common.FieldFilter{{
						FieldName: "email",
						Operator:  common.FilterOperator("contains"),
						Value:     "acme",
					}},
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrUnsupportedSearchOperator},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableSearcher, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
