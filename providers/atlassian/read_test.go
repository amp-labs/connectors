package atlassian

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "jql-error.json")
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
		WithModule(providers.ModuleAtlassianJira),
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
