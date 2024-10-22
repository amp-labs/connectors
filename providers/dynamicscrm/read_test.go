package dynamicscrm

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseContactsGet := testutils.DataFromFile(t, "contacts-read.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "contact"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.ReadParams{ObjectName: "contact", Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "contact", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusBadRequest, `{
					"error": {
						"code": "0x80060888",
						"message":"Resource not found for the segment 'conacs'."
					}
				}`),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest, errors.New("Resource not found for the segment 'conacs'"), // nolint:goerr113
			},
		},
		{
			Name:  "Incorrect key in payload",
			Input: common.ReadParams{ObjectName: "contact", Fields: connectors.Fields("id")},
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
			Input: common.ReadParams{ObjectName: "contact", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"value": {}
				}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name:  "Next page cursor may be missing in payload",
			Input: common.ReadParams{ObjectName: "contact", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup: mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{
					"value": []
				}`),
			}.Server(),
			Expected: &common.ReadResult{
				Data: []common.ReadResultRow{},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read with chosen fields",
			Input: common.ReadParams{ObjectName: "contact", Fields: connectors.Fields("fullname", "fax")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseContactsGet),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"fullname": "Heriberto Nathan",
						"fax":      "614-555-0122",
					},
					Raw: map[string]any{
						"@odata.etag":   "W/\"4372108\"",
						"fullname":      "Heriberto Nathan",
						"emailaddress1": "heriberto@northwindtraders.com",
						"fax":           "614-555-0122",
						"familystatuscode@OData.Community.Display.V1.FormattedValue": "Single",
						"familystatuscode": float64(1),
						"contactid":        "cdcfa450-cb0c-ea11-a813-000d3a1b1223",
					},
				}, {
					Fields: map[string]any{
						"fullname": "Dwayne Elijah",
						"fax":      "281-555-0158",
					},
					Raw: map[string]any{
						"@odata.etag":   "W/\"4372115\"",
						"fullname":      "Dwayne Elijah",
						"emailaddress1": "dwayne@alpineskihouse.com",
						"fax":           "281-555-0158",
						"familystatuscode@OData.Community.Display.V1.FormattedValue": "Single",
						"familystatuscode": float64(1),
						"contactid":        "9fd4a450-cb0c-ea11-a813-000d3a1b1223",
					},
				}},
				NextPage: "https://org5bd08fdd.api.crm.dynamics.com/api/data/v9.2/contacts?$select=fullname,emailaddress1,fax,familystatuscode&$skiptoken=%3Ccookie%20pagenumber=%222%22%20pagingcookie=%22%253ccookie%2520page%253d%25221%2522%253e%253ccontactid%2520last%253d%2522%257b9FD4A450-CB0C-EA11-A813-000D3A1B1223%257d%2522%2520first%253d%2522%257bCDCFA450-CB0C-EA11-A813-000D3A1B1223%257d%2522%2520%252f%253e%253c%252fcookie%253e%22%20istracking=%22False%22%20/%3E", // nolint:lll
				Done:     false,
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
		WithWorkspace("test-workspace"),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
