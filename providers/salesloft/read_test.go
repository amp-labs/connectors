package salesloft

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

	responseEmptyRead := testutils.DataFromFile(t, "read-empty.json")
	responseListPeople := testutils.DataFromFile(t, "read-list-people.json")
	responseListUsers := testutils.DataFromFile(t, "read-list-users.json")
	responseListAccounts := testutils.DataFromFile(t, "read-list-accounts.json")
	responseListAccountsSince := testutils.DataFromFile(t, "read-list-accounts-since.json")
	accountsSince, err := time.Parse(time.RFC3339Nano, "2024-06-07T10:51:20.851224-04:00")
	mockutils.NoErrors(t, err)

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unsupported object name",
			Input:        common.ReadParams{ObjectName: "butterflies", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusNotFound, `{
					"error": "Not Found"
				}`),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest, errors.New("Not Found"),
			},
		},
		{
			Name:  "Incorrect key in payload",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")},
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
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"data": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name:  "Next page cursor may be missing in payload",
			Input: common.ReadParams{ObjectName: "users", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseEmptyRead),
			}.Server(),
			Expected:     &common.ReadResult{Done: true, Data: []common.ReadResultRow{}},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page URL is correctly inferred",
			Input: common.ReadParams{ObjectName: "people", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/people"),
				Then:  mockserver.Response(http.StatusOK, responseListPeople),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     25,
				NextPage: testroutines.URLTestServer + "/v2/people?page=2&per_page=100",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read with 25 entries, checking one row",
			Input: common.ReadParams{ObjectName: "people", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/people"),
				Then:  mockserver.Response(http.StatusOK, responseListPeople),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 25,
				// We are only interested to validate only first Read Row!
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id": float64(164510523),
					},
					Raw: map[string]any{
						"first_name":             "Lynnelle",
						"email_address":          "losbourn29@paypal.com",
						"full_email_address":     "\"Lynnelle new\" <losbourn29@paypal.com>",
						"person_company_website": "http://paypal.com",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v2/people?page=2&per_page=100",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read with chosen fields",
			Input: common.ReadParams{
				ObjectName: "people",
				Fields:     connectors.Fields("email_address", "person_company_website"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/people"),
				Then:  mockserver.Response(http.StatusOK, responseListPeople),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 25,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"email_address":          "losbourn29@paypal.com",
						"person_company_website": "http://paypal.com",
					},
					Raw: map[string]any{
						"first_name":             "Lynnelle",
						"email_address":          "losbourn29@paypal.com",
						"full_email_address":     "\"Lynnelle new\" <losbourn29@paypal.com>",
						"person_company_website": "http://paypal.com",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v2/people?page=2&per_page=100",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Listing Users without pagination payload",
			Input: common.ReadParams{
				ObjectName: "users",
				Fields:     connectors.Fields("email", "guid"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v2/users"),
				Then:  mockserver.Response(http.StatusOK, responseListUsers),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"guid":  "0863ed13-7120-479b-8650-206a3679e2fb",
						"email": "somebody@withampersand.com",
					},
					Raw: map[string]any{
						"name":       "Int User",
						"first_name": "Int",
						"last_name":  "User",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read accounts without since query",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/accounts"),
					mockcond.QueryParamsMissing("updated_at[gte]"),
				},
				Then: mockserver.Response(http.StatusOK, responseListAccounts),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 4, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read accounts since point in time",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Since:      accountsSince,
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v2/accounts"),
					mockcond.QueryParam("updated_at[gte]", "2024-06-07T10:51:20.851224-04:00"),
				},
				Then: mockserver.Response(http.StatusOK, responseListAccountsSince),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 2, NextPage: "", Done: true},
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
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
