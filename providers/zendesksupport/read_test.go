package zendesksupport

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestReadZendeskSupportModule(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "resource-not-found.json")
	responseForbiddenError := testutils.DataFromFile(t, "forbidden.json")
	responseUsersFirstPage := testutils.DataFromFile(t, "read-users-1-first-page.json")
	responseUsersEmptyPage := testutils.DataFromFile(t, "read-users-2-empty-page.json")
	responseReadTickets := testutils.DataFromFile(t, "read-list-tickets.json")

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
			Name:         "Object coming from different module is unknown",
			Input:        common.ReadParams{ObjectName: "user_segments", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "triggers", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseErrorFormat),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest, errors.New("[InvalidEndpoint]Not found"), // nolint:goerr113
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
				common.ErrForbidden, errors.New("[Forbidden]You do not have access to this page"), // nolint:goerr113
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
			Name:  "Next page cursor may be missing in payload",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUsersEmptyPage),
			}.Server(),
			Expected:     &common.ReadResult{Done: true, Data: []common.ReadResultRow{}},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is resolved, when provided with a string",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseUsersFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 3,
				NextPage: "https://d3v-ampersand.zendesk.com/api/v2/users" +
					"?page%5Bafter%5D=eyJvIjoiaWQiLCJ2IjoiYVJOc1cwVDZGd0FBIn0%3D&page%5Bsize%5D=3",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read with chosen fields",
			Input: common.ReadParams{
				ObjectName: "tickets",
				Fields:     connectors.Fields("type", "subject", "status"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/v2/tickets"),
				Then:  mockserver.Response(http.StatusOK, responseReadTickets),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"type":    "incident",
						"subject": "Updated to: Hello World",
						"status":  "open",
					},
					Raw: map[string]any{
						"url":           "https://d3v-ampersand.zendesk.com/api/v2/tickets/1.json",
						"problem_id":    nil,
						"has_incidents": false,
						"brand_id":      float64(26363596759827),
					},
				}},
				NextPage: "https://d3v-ampersand.zendesk.com/api/v2/tickets.json" +
					"?page%5Bafter%5D=eyJvIjoibmljZV9pZCIsInYiOiJhUUVBQUFBQUFBQUEifQ%3D%3D&page%5Bsize%5D=1",
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
				return constructTestConnector(tt.Server.URL, ModuleTicketing)
			})
		})
	}
}

func TestReadHelpCenterModule(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseReadPosts := testutils.DataFromFile(t, "read-posts.json")

	tests := []testroutines.Read{
		{
			Name:         "Object coming from different module is unknown",
			Input:        common.ReadParams{ObjectName: "triggers", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Successful read with chosen fields",
			Input: common.ReadParams{
				ObjectName: "posts",
				Fields:     connectors.Fields("title", "topic_id", "status"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/v2/community/posts"),
				Then:  mockserver.Response(http.StatusOK, responseReadPosts),
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
				return constructTestConnector(tt.Server.URL, ModuleHelpCenter)
			})
		})
	}
}

func constructTestConnector(serverURL string, moduleID common.ModuleID) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithWorkspace("test-workspace"),
		WithModule(moduleID),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
