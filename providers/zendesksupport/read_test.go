package zendesksupport

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "resource-not-found.json")
	responseForbiddenError := testutils.DataFromFile(t, "forbidden.json")
	responseTriggersFirstPage := testutils.DataFromFile(t, "read/triggers-1-first-page.json")
	responseTriggersLastPage := testutils.DataFromFile(t, "read/triggers-2-last-page.json")
	responseReadPosts := testutils.DataFromFile(t, "read/help-center-posts.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "triggers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "triggers", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseErrorFormat),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest, errors.New("[InvalidEndpoint]Not found"),
			},
		},
		{
			Name:  "Forbidden error code and response",
			Input: common.ReadParams{ObjectName: "triggers", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusForbidden, responseForbiddenError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrForbidden, errors.New("[Forbidden]You do not have access to this page"),
			},
		},
		{
			Name:  "Incorrect key in payload",
			Input: common.ReadParams{ObjectName: "triggers", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"garbage": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			Name:  "Incorrect data type in payload",
			Input: common.ReadParams{ObjectName: "triggers", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"triggers": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name:  "Next page URL is resolved, when provided with a string",
			Input: common.ReadParams{ObjectName: "triggers", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseTriggersFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "https://d3v-ampersand.zendesk.com/api/v2/triggers?page%5Bafter%5D=eyJvIjoicG9zaXRpb24scG9zaXRpb24sdGl0bGUsaWQiLCJ2IjoiYVFFQUFBQUFBQUFBYVFFQUFBQUFBQUFBY3gwQUFBQk9iM1JwWm5rZ1lYTnphV2R1WldVZ2IyWWdZWE56YVdkdWJXVnVkR21UOTlOQStoY0FBQT09In0%3D&page%5Bsize%5D=1", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Triggers is the last page",
			Input: common.ReadParams{ObjectName: "triggers", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/triggers"),
				Then:  mockserver.Response(http.StatusOK, responseTriggersLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"id": float64(26363596909203)},
					Raw:    map[string]any{"title": "Notify all agents of received request"},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read of help desk's posts with chosen fields",
			Input: common.ReadParams{
				ObjectName: "posts",
				Fields:     connectors.Fields("title", "topic_id", "status"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/community/posts"),
					mockcond.QueryParam("page[size]", "100"),
				},
				Then: mockserver.Response(http.StatusOK, responseReadPosts),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"title":    "How do I get around the community?",
						"topic_id": float64(27980065395091),
						"status":   "none",
					},
					Raw: map[string]any{
						"id":         float64(27980065413139),
						"created_at": "2024-04-01T13:01:11Z",
						"updated_at": "2024-04-01T13:01:11Z",
					},
				}},
				NextPage: "https://d3v-ampersand.zendesk.com" +
					"/api/v2/help_center/community/posts.json?page=2&per_page=30",
				Done: false,
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

func TestIncrementalRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseTickets := testutils.DataFromFile(t, "read/incremental/tickets.json")
	responseTicketsCustomFields := testutils.DataFromFile(t, "read/custom_fields/ticket_fields.json")
	responseUsersFirstPage := testutils.DataFromFile(t, "read/incremental/users-1-first-page.json")
	responseUsersLastPage := testutils.DataFromFile(t, "read/incremental/users-2-last-page.json")
	responseOrganizations := testutils.DataFromFile(t, "read/incremental/organizations.json")

	tests := []testroutines.Read{
		{
			Name: "Incremental Tickets no since with custom fields",
			Input: common.ReadParams{
				ObjectName: "tickets",
				Fields:     connectors.Fields("id", "Customer Type", "Topic"),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.QueryParam("per_page", "2000"),
						mockcond.QueryParam("start_time", "0"),
						mockcond.Path("/api/v2/incremental/tickets/cursor"),
					},
					Then: mockserver.Response(http.StatusOK, responseTickets),
				}, {
					If:   mockcond.Path("/api/v2/ticket_fields"),
					Then: mockserver.Response(http.StatusOK, responseTicketsCustomFields),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":            float64(5),
						"customer type": "standard_customer",
						"topic":         "inquiry",
					},
					Raw: map[string]any{
						"priority": "normal",
						"custom_fields": []any{
							map[string]any{
								"id":    float64(26363655924371),
								"value": "standard_customer",
							},
							map[string]any{
								"id":    float64(26363685850259),
								"value": "inquiry",
							},
						},
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Incremental Tickets with Since",
			Input: common.ReadParams{
				ObjectName: "tickets",
				Fields:     connectors.Fields("id"),
				Since:      time.Unix(1726674883, 0),
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.QueryParam("per_page", "2000"),
						mockcond.QueryParam("start_time", "1726674883"),
						mockcond.Path("/api/v2/incremental/tickets/cursor"),
					},
					Then: mockserver.Response(http.StatusOK, responseTickets),
				}, {
					If:   mockcond.Path("/api/v2/ticket_fields"),
					Then: mockserver.Response(http.StatusOK, responseTicketsCustomFields),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 1, NextPage: "", Done: true},
		},
		{
			Name: "Users first page no since",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.QueryParam("per_page", "1000"),
					mockcond.QueryParam("start_time", "0"),
					mockcond.Path("/api/v2/incremental/users/cursor"),
				},
				Then: mockserver.Response(http.StatusOK, responseUsersFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"name": "The Customer"},
					Raw:    map[string]any{"email": "customer@example.com"},
				}},
				NextPage: "https://d3v-ampersand.zendesk.com/api/v2/incremental/users/cursor?cursor=MTczOTkwMTY5OC4wfHwzODYzNDM4MTYyMDM3MXw%3D&per_page=19", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Users last page is missing",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.QueryParam("per_page", "1000"),
					mockcond.QueryParam("start_time", "0"),
					mockcond.Path("/api/v2/incremental/users/cursor"),
				},
				Then: mockserver.Response(http.StatusOK, responseUsersLastPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{"name": "TEST USER"},
					Raw:    map[string]any{"email": "admin@test.com"},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Organizations has no next page, response body returns time-based URL alongside indication of the end",
			Input: common.ReadParams{
				ObjectName: "organizations",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.QueryParam("per_page", "1000"),
					mockcond.QueryParam("start_time", "0"),
					mockcond.Path("/api/v2/incremental/organizations"),
				},
				Then: mockserver.Response(http.StatusOK, responseOrganizations),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
		WithWorkspace("test-workspace"),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
