package front

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// Search action is dispatched through this interface.
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

func TestSearch(t *testing.T) { // nolint:funlen,maintidx
	t.Parallel()

	responseContact := testutils.DataFromFile(t, "search/contact.json")
	responseConversation := testutils.DataFromFile(t, "search/conversation.json")
	responseConversations := testutils.DataFromFile(t, "search/conversations.json")
	responseConversationsEmpty := testutils.DataFromFile(t, "search/conversations-empty.json")
	responseNotFound := testutils.DataFromFile(t, "search/not-found.json")

	contact := common.ReadResultRow{
		Fields: map[string]any{"id": "crd_4mhsx24", "name": "Pep Hernandez"},
		Raw:    map[string]any{"description": "A market lead at Acme Inc", "is_private": false},
		Id:     "crd_4mhsx24",
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
				ObjectName: "contacts",
				Filter:     equalsFilter("id", "crd_4mhsx24"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "A filter is required",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingSearchFilters},
		},
		{
			Name: "Object without lookups cannot be searched",
			Input: common.SearchParams{
				ObjectName: "tags",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("id", "tag_1"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Contacts cannot be searched by status",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("status", "archived"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrUnsupportedSearchField},
		},
		{
			Name: "Conversations cannot be searched by email",
			Input: common.SearchParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("email", "pep@example.com"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrUnsupportedSearchField},
		},
		{
			Name: "Only the equality operator is supported",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
				Filter: common.SearchFilter{
					FieldFilters: []common.FieldFilter{{
						FieldName: "email",
						Operator:  common.FilterOperator("contains"),
						Value:     "example.com",
					}},
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrUnsupportedSearchOperator},
		},
		{
			Name: "Only one filter is accepted",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
				Filter: common.SearchFilter{
					FieldFilters: []common.FieldFilter{
						{FieldName: "email", Operator: common.FilterOperatorEQ, Value: "pep@example.com"},
						{FieldName: "phone", Operator: common.FilterOperatorEQ, Value: "+15555550100"},
					},
				},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrUnsupportedSearchFilterCombination},
		},
		{
			Name: "Filter value must be a string",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("id", 7),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrInvalidSearchValue},
		},
		{
			Name: "Status cannot be a multi term search query",
			Input: common.SearchParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("status", "archived is:assigned"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrInvalidSearchValue},
		},
		{
			Name: "Contact by identifier",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
				Filter:     equalsFilter("id", "crd_4mhsx24"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/contacts/crd_4mhsx24"),
				},
				Then: mockserver.Response(http.StatusOK, responseContact),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{contact},
				Done: true,
			},
		},
		{
			Name: "Contact by email uses the email alias",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
				Filter:     equalsFilter("email", "pep@example.com"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/contacts/alt:email:pep@example.com"),
				Then:  mockserver.Response(http.StatusOK, responseContact),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{contact},
				Done: true,
			},
		},
		{
			Name: "Contact by phone uses the phone alias",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
				Filter:     equalsFilter("phone", "+15555550100"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/contacts/alt:phone:+15555550100"),
				Then:  mockserver.Response(http.StatusOK, responseContact),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{contact},
				Done: true,
			},
		},
		{
			Name: "Phone with spaces is passed on, Front normalizes it",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id", "name"),
				Filter:     equalsFilter("phone", "+1 555 555 0100"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/contacts/alt:phone:+1 555 555 0100"),
				Then:  mockserver.Response(http.StatusOK, responseContact),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{contact},
				Done: true,
			},
		},
		{
			Name: "Missing contact produces no records",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("email", "nobody@example.com"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFound),
			}.Server(),
			Expected: &common.SearchResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
		},
		{
			Name: "Server errors other than not found are returned",
			Input: common.SearchParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("id", "crd_4mhsx24"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, responseNotFound),
			}.Server(),
			ExpectedErrs: []error{common.ErrForbidden},
		},
		{
			Name: "Conversation by identifier",
			Input: common.SearchParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("id", "subject", "status"),
				Filter:     equalsFilter("id", "cnv_1alzn15o"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/conversations/cnv_1alzn15o"),
				Then:  mockserver.Response(http.StatusOK, responseConversation),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "cnv_1alzn15o", "subject": "Package update?", "status": "archived"},
					Raw:    map[string]any{"type": "conversation"},
					Id:     "cnv_1alzn15o",
				}},
				Done: true,
			},
		},
		{
			Name: "Conversations by status use the search endpoint",
			Input: common.SearchParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("id", "subject", "status"),
				Filter:     equalsFilter("status", "archived"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/conversations/search/is:archived"),
					mockcond.QueryParam("limit", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseConversations),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": "cnv_1alzn15o", "subject": "Package update?", "status": "archived"},
					Raw:    map[string]any{"type": "conversation"},
					Id:     "cnv_1alzn15o",
				}, {
					Fields: map[string]any{"id": "cnv_2bmao26p", "subject": "Refund request", "status": "archived"},
					Raw:    map[string]any{"type": "conversation"},
					Id:     "cnv_2bmao26p",
				}},
				NextPage: "https://example.api.frontapp.com/conversations/search/is:archived?limit=100&page_token=abc",
				Done:     false,
			},
		},
		{
			Name: "Status is passed to the search as is, limit is honoured",
			Input: common.SearchParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("id"),
				Filter:     equalsFilter("status", "trashed"),
				Limit:      5,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/conversations/search/is:trashed"),
					mockcond.QueryParam("limit", "5"),
				},
				Then: mockserver.Response(http.StatusOK, responseConversationsEmpty),
			}.Server(),
			Expected: &common.SearchResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
		},
		{
			Name: "Next page of a status search is followed as is",
			Input: common.SearchParams{
				ObjectName: "conversations",
				Fields:     connectors.Fields("id", "subject", "status"),
				Filter:     equalsFilter("status", "archived"),
				NextPage:   testconn.URLTestServer + "/conversations/search/is:archived?limit=100&page_token=abc",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/conversations/search/is:archived"),
					mockcond.QueryParam("page_token", "abc"),
				},
				Then: mockserver.Response(http.StatusOK, responseConversationsEmpty),
			}.Server(),
			Expected: &common.SearchResult{
				Rows: 0,
				Data: []common.ReadResultRow{},
				Done: true,
			},
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseUrl(serverURL)

	return connector, nil
}
