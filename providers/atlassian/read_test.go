package atlassian

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestReadJira(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "jql-error.json")
	responseIssuesFirstPage := testutils.DataFromFile(t, "read-issues.json")
	responsePathNotFoundError := testutils.DataFromFile(t, "path-not-found.json")
	responseUnauthenticatedHTML := testutils.DataFromFile(t, "connect-un-auth.html")

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
			Name:  "Error can be parsed from HTML response",
			Input: common.ReadParams{ObjectName: "issues", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusBadRequest, responseUnauthenticatedHTML),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Unauthorized (401)"), // nolint:goerr113
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
						"id":  "10001",
						"key": "AM-2",
					},
				}, {
					Fields: map[string]any{
						"id":      "10000",
						"summary": "First Issue on Jira",
					},
					Raw: map[string]any{
						"id":  "10000",
						"key": "AM-1",
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

func TestReadConfluence(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorInvalidPath := testutils.DataFromFile(t, "confluence/read/err-invalid-path.json")
	errorBadPageSize := testutils.DataFromFile(t, "confluence/read/err-page-size.json")
	responseBlogpostsFirstPage := testutils.DataFromFile(t, "confluence/read/blogposts/1-first-page.json")
	responseBlogpostsLastPage := testutils.DataFromFile(t, "confluence/read/blogposts/2-last-page.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "blogposts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Error message page size out of range is parsed",
			Input: common.ReadParams{ObjectName: "blogposts", Fields: connectors.Fields("body")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorBadPageSize),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Provided size {300} for 'limit' is greater than the max allowed: 250"), // nolint:goerr113
			},
		},
		{
			Name:  "Error message invalid path is parsed",
			Input: common.ReadParams{ObjectName: "blogposts", Fields: connectors.Fields("body")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorInvalidPath),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("No message available"), // nolint:goerr113
			},
		},
		{
			Name:  "First page has next page reference",
			Input: common.ReadParams{ObjectName: "blogposts", Fields: connectors.Fields("title")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/ex/confluence/ebc887b2-7e61-4059-ab35-71f15cc16e12/wiki/api/v2/blogposts"),
					mockcond.QueryParam("limit", "250"),
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link",
						`</wiki/api/v2/blogposts?limit=2&cursor=eyJpZCI6Ijc1MzY2OSIsImNvbnRlbnRPcmRlciI6ImlkIiwiY29udGVudE9yZGVyVmFsdWUiOjc1MzY2OX0=>; rel="next", <https://withampersand-team-oqo0hkaj.atlassian.net/wiki>; rel="base"`), // nolint:lll
					mockserver.Response(http.StatusOK, responseBlogpostsFirstPage),
				),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"title": "First blog post ever!",
					},
					Raw: map[string]any{
						"id":        "688129",
						"spaceId":   "131189",
						"createdAt": "2025-06-26T20:39:16.596Z",
					},
				}, {
					Fields: map[string]any{
						"title": "Second blog post!",
					},
					Raw: map[string]any{
						"id":        "753669",
						"spaceId":   "131189",
						"createdAt": "2025-06-26T20:39:37.501Z",
					},
				}},
				NextPage: testroutines.URLTestServer + "/ex/confluence/ebc887b2-7e61-4059-ab35-71f15cc16e12/wiki/api/v2/blogposts?limit=2&cursor=eyJpZCI6Ijc1MzY2OSIsImNvbnRlbnRPcmRlciI6ImlkIiwiY29udGVudE9yZGVyVmFsdWUiOjc1MzY2OX0=", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read with chosen fields using next page token",
			Input: common.ReadParams{
				ObjectName: "blogposts",
				NextPage:   testroutines.URLTestServer + "/wiki/api/v2/blogposts?limit=2&cursor=eyJpZCI6Ijc1MzY2OSIsImNvbnRlbnRPcmRlciI6ImlkIiwiY29udGVudE9yZGVyVmFsdWUiOjc1MzY2OX0=", // nolint:lll
				Fields:     connectors.Fields("title"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/wiki/api/v2/blogposts"),
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParam("cursor", "eyJpZCI6Ijc1MzY2OSIsImNvbnRlbnRPcmRlciI6ImlkIiwiY29udGVudE9yZGVyVmFsdWUiOjc1MzY2OX0="), // nolint:lll
				},
				Then: mockserver.ResponseChainedFuncs(
					mockserver.Header("Link",
						`<https://withampersand-team-oqo0hkaj.atlassian.net/wiki>; rel="base"`),
					mockserver.Response(http.StatusOK, responseBlogpostsLastPage),
				),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"title": "Even a third blog post, on a roll!",
					},
					Raw: map[string]any{
						"id":        "819206",
						"spaceId":   "131189",
						"createdAt": "2025-06-26T20:39:48.350Z",
					},
				}},
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
				return constructTestConnectorConfluence(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	return constructTestConnectorGeneral(serverURL, providers.ModuleAtlassianJira)
}

func constructTestConnectorConfluence(serverURL string) (*Connector, error) {
	return constructTestConnectorGeneral(serverURL, providers.ModuleAtlassianConfluence)
}

func constructTestConnectorGeneral(serverURL string, module common.ModuleID) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
		WithWorkspace("test-workspace"),
		WithModule(module),
		WithMetadata(map[string]string{
			"cloudId": "ebc887b2-7e61-4059-ab35-71f15cc16e12", // any value will work for the test
		}),
	)
	if err != nil {
		return nil, err
	}

	connector.setBaseURL(
		mockutils.ReplaceURLOrigin(connector.providerInfo.BaseURL, serverURL),
		mockutils.ReplaceURLOrigin(connector.moduleInfo.BaseURL, serverURL),
	)

	return connector, nil
}
