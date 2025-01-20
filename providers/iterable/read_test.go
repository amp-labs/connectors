package iterable

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCatalogsFirst := testutils.DataFromFile(t, "read-catalogs-1-first-page.json")
	responseCatalogsLast := testutils.DataFromFile(t, "read-catalogs-2-second-page.json")
	errorTemplatesInvalidSince := testutils.DataFromFile(t, "read-templates-invalid-since.html")
	errorInvalidRestMethod := testutils.DataFromFile(t, "error-wrong-method.html")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "journeys"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id")},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Invalid usage of REST operation",
			Input: common.ReadParams{ObjectName: "templates", Fields: connectors.Fields("templateId")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound, errorInvalidRestMethod),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"not found: Action not found: Oh no- that url doesn't exist."),
			},
		},
		{
			Name: "Catalogs first page has a link to the next",
			Input: common.ReadParams{
				ObjectName: "catalogs",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/api/catalogs"),
				Then:  mockserver.Response(http.StatusOK, responseCatalogsFirst),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				Data:     []common.ReadResultRow{},
				NextPage: testroutines.URLTestServer + "/api/catalogs?pageSize=1&page=2", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Catalogs last page has no link to the next",
			Input: common.ReadParams{
				ObjectName: "catalogs",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/api/catalogs"),
				Then:  mockserver.Response(http.StatusOK, responseCatalogsLast),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read of templates with invalid time format error", // Test exists for diverse HTML error handling.
			Input: common.ReadParams{ObjectName: "templates", Fields: connectors.Fields("templateId")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound, errorTemplatesInvalidSince),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"Unable to bind 1734632773073 into a date time!Acceptable formats are ISO8601"),
			},
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
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
