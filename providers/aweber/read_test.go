package aweber

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	errorBadRequest := testutils.DataFromFile(t, "get-with-req-body-not-allowed.html")
	errorNotFound := testutils.DataFromFile(t, "url-not-found.html")
	responseContactsFirstPage := testutils.DataFromFile(t, "read-contacts-1-first-page-v1.json")
	responseContactsEmptyPage := testutils.DataFromFile(t, "read-contacts-2-empty-page-v1.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "messages"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.ReadParams{ObjectName: "messages", Fields: connectors.Fields("id")},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Error cannot send request body on GET operation",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusBadRequest, errorBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("400 Bad Request: Your client has issued a malformed or illegal request."), // nolint:goerr113
			},
		},
		{
			Name:  "Error page not found",
			Input: common.ReadParams{ObjectName: "contacts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentHTML(),
				Always: mockserver.Response(http.StatusNotFound, errorNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Keap - Page Not Found"), // nolint:goerr113
			},
		},
		{
			Name: "Contacts first page has a link to next",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("given_name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/crm/rest/v1/contacts"),
				Then:  mockserver.Response(http.StatusOK, responseContactsFirstPage),
			}.Server(),
			Comparator: func(baseURL string, actual, expected *common.ReadResult) bool {
				return mockutils.ReadResultComparator.SubsetFields(actual, expected) &&
					mockutils.ReadResultComparator.SubsetRaw(actual, expected) &&
					actual.NextPage.String() == expected.NextPage.String() &&
					actual.Rows == expected.Rows &&
					actual.Done == expected.Done
			},
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"given_name": "Erica",
					},
					Raw: map[string]any{
						"id":          float64(22),
						"family_name": "Lewis",
					},
				}, {
					Fields: map[string]any{
						"given_name": "John",
					},
					Raw: map[string]any{
						"id":          float64(20),
						"family_name": "Doe",
					},
				}},
				NextPage: "https://api.infusionsoft.com/crm/rest/v1/contacts/?limit=2&offset=2&since=2024-06-03T22:17:59.039Z&order=id", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Read contacts empty page",
			Input: common.ReadParams{
				ObjectName: "contacts",
				Fields:     connectors.Fields("given_name"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactsEmptyPage),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "https://api.infusionsoft.com/crm/rest/v1/contacts/?limit=2&offset=4&since=2024-06-03T22:17:59.039Z&order=id", // nolint:lll
				Done:     true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
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
