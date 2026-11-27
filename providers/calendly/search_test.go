package calendly

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

const (
	testUserURI = "https://api.calendly.com/users/eacddb9f-f46d-4c98-ae7e-cd28651420a5"
	testOrgURI  = "https://api.calendly.com/organizations/a2bf01c6-e8d1-44ea-afa7-14592833ffe7"
)

// Search action is dispatched through this interface.
var _ connectors.SearchConnector = &Connector{}

func equalsFilter(pairs ...any) common.SearchFilter {
	filter := common.SearchFilter{}

	for index := 0; index+1 < len(pairs); index += 2 {
		filter.FieldFilters = append(filter.FieldFilters, common.FieldFilter{
			FieldName: pairs[index].(string), // nolint:forcetypeassert
			Operator:  common.FilterOperatorEQ,
			Value:     pairs[index+1],
		})
	}

	return filter
}

func TestSearch(t *testing.T) { // nolint:funlen,maintidx
	t.Parallel()

	responseEvent := testutils.DataFromFile(t, "search/event.json")
	responseEvents := testutils.DataFromFile(t, "search/events.json")
	responseEventsEmpty := testutils.DataFromFile(t, "search/events-empty.json")
	responseNotFound := testutils.DataFromFile(t, "search/not-found.json")
	responseInvalidArgument := testutils.DataFromFile(t, "search/invalid-argument.json")

	const (
		eventUUID = "bf1b889a-fecc-492d-a5d5-8d224bf5b1f9"
		eventURI  = "https://api.calendly.com/scheduled_events/" + eventUUID
	)

	event := common.ReadResultRow{
		Fields: map[string]any{"uri": eventURI, "status": "active"},
		Raw:    map[string]any{"name": "30 Minute Meeting", "start_time": "2026-11-30T06:00:00.000000Z"},
	}

	emptyResult := &common.SearchResult{
		Rows: 0,
		Data: []common.ReadResultRow{},
		Done: true,
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
				ObjectName: "scheduled_events",
				Filter:     equalsFilter("uuid", eventUUID),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "A filter is required",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingSearchFilters},
		},
		{
			Name: "Object without lookups cannot be searched",
			Input: common.SearchParams{
				ObjectName: "event_types",
				Fields:     connectors.Fields("uri"),
				Filter:     equalsFilter("uuid", "e680f436-8e38-440e-8a77-11a86e23be4a"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Unknown field cannot be searched by",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter:     equalsFilter("name", "30 Minute Meeting"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrUnsupportedSearchField},
		},
		{
			Name: "Only the equality operator is supported",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter: common.SearchFilter{FieldFilters: []common.FieldFilter{{
					FieldName: "min_start_time",
					Operator:  "gte",
					Value:     "2026-12-01T06:00:00Z",
				}}},
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrUnsupportedSearchOperator},
		},
		{
			Name: "UUID cannot be combined with other filters",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter:     equalsFilter("uuid", eventUUID, "status", "active"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrUnsupportedSearchFilterCombination},
		},
		{
			Name: "A field cannot be filtered twice",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter:     equalsFilter("status", "active", "status", "canceled"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrUnsupportedSearchFilterCombination},
		},
		{
			Name: "Filter value must be a string or time",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter:     equalsFilter("min_start_time", 1764568800),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrInvalidSearchValue},
		},
		{
			Name: "Filter value cannot be empty",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter:     equalsFilter("invitee_email", "  "),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrInvalidSearchValue},
		},
		{
			Name: "Event by uuid",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri", "status"),
				Filter:     equalsFilter("uuid", eventUUID),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/scheduled_events/" + eventUUID),
				},
				Then: mockserver.Response(http.StatusOK, responseEvent),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows: 1,
				Data: []common.ReadResultRow{event},
				Done: true,
			},
		},
		{
			Name: "Unknown uuid produces an empty result",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter:     equalsFilter("uuid", "00000000-0000-0000-0000-000000000000"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFound),
			}.Server(),
			Expected: emptyResult,
		},
		{
			Name: "Events by invitee email are scoped to the organization and user",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri", "status"),
				Filter:     equalsFilter("invitee_email", "search.alpha@example.com"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/scheduled_events"),
					mockcond.QueryParam("invitee_email", "search.alpha@example.com"),
					mockcond.QueryParam("organization", testOrgURI),
					mockcond.QueryParam("user", testUserURI),
					mockcond.QueryParam("count", "100"),
					mockcond.QueryParamsMissing("status", "min_start_time", "max_start_time"),
				},
				Then: mockserver.Response(http.StatusOK, responseEvents),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.SearchResult{
				Rows:     2,
				Data:     []common.ReadResultRow{event},
				NextPage: "https://api.calendly.com/scheduled_events?count=2&page_token=28qkRkmibstGWFNnXTldX2SVr0cN2wXA",
				Done:     false,
			},
		},
		{
			Name: "Events by status, limit is honoured",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter:     equalsFilter("status", "canceled"),
				Limit:      5,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/scheduled_events"),
					mockcond.QueryParam("status", "canceled"),
					mockcond.QueryParam("count", "5"),
				},
				Then: mockserver.Response(http.StatusOK, responseEventsEmpty),
			}.Server(),
			Expected: emptyResult,
		},
		{
			Name: "Events by start time range given as string and time",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter: equalsFilter(
					"min_start_time", "2026-11-30T00:00:00Z",
					"max_start_time", time.Date(2026, 12, 1, 23, 0, 0, 0, time.UTC),
				),
				Limit: 500,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/scheduled_events"),
					mockcond.QueryParam("min_start_time", "2026-11-30T00:00:00Z"),
					mockcond.QueryParam("max_start_time", "2026-12-01T23:00:00Z"),
					mockcond.QueryParam("count", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseEventsEmpty),
			}.Server(),
			Expected: emptyResult,
		},
		{
			Name: "All list filters are combined in one request",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter: equalsFilter(
					"invitee_email", "search.alpha@example.com",
					"status", "active",
					"min_start_time", "2026-11-29T00:00:00Z",
					"max_start_time", "2026-12-03T00:00:00Z",
				),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/scheduled_events"),
					mockcond.QueryParam("invitee_email", "search.alpha@example.com"),
					mockcond.QueryParam("status", "active"),
					mockcond.QueryParam("min_start_time", "2026-11-29T00:00:00Z"),
					mockcond.QueryParam("max_start_time", "2026-12-03T00:00:00Z"),
				},
				Then: mockserver.Response(http.StatusOK, responseEventsEmpty),
			}.Server(),
			Expected: emptyResult,
		},
		{
			Name: "Value rejected by the provider is an error",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter:     equalsFilter("status", "cancelled"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseInvalidArgument),
			}.Server(),
			ExpectedErrs: []error{common.ErrCaller},
		},
		{
			Name: "Next page is followed as is",
			Input: common.SearchParams{
				ObjectName: "scheduled_events",
				Fields:     connectors.Fields("uri"),
				Filter:     equalsFilter("status", "active"),
				NextPage:   testconn.URLTestServer + "/scheduled_events?count=2&page_token=abc",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/scheduled_events"),
					mockcond.QueryParam("page_token", "abc"),
					mockcond.QueryParam("count", "2"),
				},
				Then: mockserver.Response(http.StatusOK, responseEventsEmpty),
			}.Server(),
			Expected: emptyResult,
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
		Metadata: map[string]string{
			"userURI":         testUserURI,
			"organizationURI": testOrgURI,
		},
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseUrl(serverURL)

	return connector, nil
}
