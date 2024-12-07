package atlassian

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "jql-error.json")
	responseIssuesFirstPage := testutils.DataFromFile(t, "read-issues.json")
	responsePathNotFoundError := testutils.DataFromFile(t, "path-not-found.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "issues"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseErrorFormat),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Date value '-53s' for field 'updated' is invalid"), // nolint:goerr113
			},
		},
		{
			Name:  "Invalid path understood as not found error",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responsePathNotFoundError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Not Found - No message available"), // nolint:goerr113
			},
		},
		{
			Name:  "Incorrect key in payload",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id")},
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
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"issues": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name:  "Empty array produces no next page",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `
				{
				  "startAt": 6,
				  "issues": []
				}`),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 0, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Issue must have fields property",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `
				{
				  "issues": [{}]
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			Name:  "Issue must have id property",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `
				{
				  "issues": [{"fields":{}}]
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrKeyNotFound},
		},
		{
			Name:  "Missing starting index produces no next page",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `
				{
				  "issues": [
					{"fields":{}, "id": "0"},
					{"fields":{}, "id": "1"}
				]}`),
			}.Server(),
			Comparator:   testroutines.ComparatorPagination,
			Expected:     &common.ReadResult{Rows: 2, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Next page is implied from start index and issues size",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `
				{
				  "startAt": 6,
				  "issues": [
					{"fields":{}, "id": "0"},
					{"fields":{}, "id": "1"}
				]}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     2,
				NextPage: "8",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Since rounds to minute time frame",
			Input: common.ReadParams{
				ObjectName: "issues",
				Fields:     connectors.Fields("id"),
				Since:      time.Now().Add(-5 * time.Minute),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				// server was asked to get issues that occurred in the last 5 min
				If: mockcond.QueryParam("jql", `updated > "-5m"`),
				Then: mockserver.ResponseString(http.StatusOK, `
					{
					  "startAt": 0,
					  "issues": [{"fields":{}, "id": "0"}]
					}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				NextPage: "1",
				Done:     false,
			},
			ExpectedErrs: nil, // there must be no errors.
		},
		{
			Name: "Next page is propagated in query params",
			Input: common.ReadParams{
				ObjectName: "issues",
				Fields:     connectors.Fields("id"),
				NextPage:   "17",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.QueryParam("startAt", "17"),
				Then: mockserver.ResponseString(http.StatusOK, `
					{
					  "startAt": 17,
					  "issues": [
						{"fields":{}, "id": "0"},
						{"fields":{}, "id": "1"},
						{"fields":{}, "id": "2"}
					]}`),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     3,
				NextPage: "20",
				Done:     false,
			},
			ExpectedErrs: nil, // there must be no errors
		},
		{
			Name: "Successful list of rows",
			Input: common.ReadParams{
				ObjectName: "issues",
				Fields:     connectors.Fields("id", "summary"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseIssuesFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":      "10001",
						"summary": "Another one",
					},
					Raw: map[string]any{
						"id":                       "10001",
						"description":              nil,
						"created":                  "2024-07-22T22:41:48.474+0300",
						"statuscategorychangedate": "2024-07-22T22:41:48.686+0300",
					},
				}, {
					Fields: map[string]any{
						"id":      "10000",
						"summary": "First Issue on Jira",
					},
					Raw: map[string]any{
						"id":                       "10000",
						"description":              nil,
						"created":                  "2024-07-22T22:41:35.069+0300",
						"statuscategorychangedate": "2024-07-22T22:41:35.326+0300",
					},
				}},
				NextPage: "2",
				Done:     false,
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

func TestReadWithoutMetadata(t *testing.T) {
	t.Parallel()

	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithWorkspace("test-workspace"),
		WithModule(ModuleJira),
	)
	if err != nil {
		t.Fatal("failed to create connector")
	}

	_, err = connector.Read(context.Background(), common.ReadParams{
		ObjectName: "issues",
		Fields:     connectors.Fields("id"),
	})
	if !errors.Is(err, ErrMissingCloudId) {
		t.Fatalf("expected Read method to complain about missing cloud id")
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithWorkspace("test-workspace"),
		WithModule(ModuleJira),
		WithMetadata(map[string]string{
			"cloudId": "ebc887b2-7e61-4059-ab35-71f15cc16e12", // any value will work for the test
		}),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
