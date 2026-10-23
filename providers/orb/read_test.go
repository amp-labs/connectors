package orb

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

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	responseUnauthorized := testutils.DataFromFile(t, "error-unauthorized.json")

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Unknown object is not supported",
			Input:        common.ReadParams{ObjectName: "plans", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:  "Error response is parsed",
			Input: common.ReadParams{ObjectName: "customers", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, responseUnauthorized),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrAccessToken,
				testutils.StringError("The Authorization header could not be verified. An invalid API key was provided."),
			},
		},
		// No successful live response has been captured yet, so the request-shape cases below answer
		// with the saved 401 only when the request matches, and anything else with 418.
		{
			Name: "Invoices request every status, including draft and void",
			Input: common.ReadParams{
				ObjectName: "invoices",
				Fields:     connectors.Fields("id"),
				PageSize:   2,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/invoices"),
					mockcond.QueryParam("limit", "2"),
					mockcond.QueryParam("status[]", "draft", "issued", "paid", "synced", "void"),
					mockcond.QueryParamsMissing("cursor"),
				},
				Then: mockserver.Response(http.StatusUnauthorized, responseUnauthorized),
				Else: mockserver.Response(http.StatusTeapot),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
		},
		{
			Name: "Next page sends the cursor and keeps the invoice statuses",
			Input: common.ReadParams{
				ObjectName: "invoices",
				Fields:     connectors.Fields("id"),
				NextPage:   "cursor-page-2",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/invoices"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("cursor", "cursor-page-2"),
					mockcond.QueryParam("status[]", "draft", "issued", "paid", "synced", "void"),
				},
				Then: mockserver.Response(http.StatusUnauthorized, responseUnauthorized),
				Else: mockserver.Response(http.StatusTeapot),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
		},
		{
			Name:  "Other objects send no status filter",
			Input: common.ReadParams{ObjectName: "customers", Fields: connectors.Fields("id")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/customers"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParamsMissing("status[]", "cursor"),
				},
				Then: mockserver.Response(http.StatusUnauthorized, responseUnauthorized),
				Else: mockserver.Response(http.StatusTeapot),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
