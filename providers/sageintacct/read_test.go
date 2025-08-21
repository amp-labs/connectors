package sageintacct

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseAccounts := testutils.DataFromFile(t, "read-accounts.json")
	responseAccountsEmpty := testutils.DataFromFile(t, "read-accounts-empty.json")
	responseCustomers := testutils.DataFromFile(t, "read-customers.json")
	responseInvalidPath := testutils.DataFromFile(t, "read-invalid-path.html")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "account"},
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
			Name:  "Correct error message is understood from HTML response",
			Input: common.ReadParams{ObjectName: "account", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseInvalidPath),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Cannot GET /ia/api/v1/objects/general-ledger/account"), // nolint:goerr113
			},
		},
		{
			Name:  "Incorrect data type in payload",
			Input: common.ReadParams{ObjectName: "account", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{}`),
			}.Server(),
			ExpectedErrs: []error{jsonquery.ErrNotArray},
		},
		{
			Name:  "Empty read response",
			Input: common.ReadParams{ObjectName: "account", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseAccountsEmpty),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, NextPage: "", Done: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read with chosen fields",
			Input: common.ReadParams{ObjectName: "account", Fields: connectors.Fields("id", "key")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/ia/api/v1/objects/general-ledger/account"),
				Then:  mockserver.Response(http.StatusOK, responseAccounts),
				Else:  mockserver.Response(http.StatusNotFound, responseInvalidPath),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 3,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":  "1000",
							"key": "1",
						},
						Raw: map[string]any{
							"id":   "1000",
							"key":  "1",
							"href": "/objects/general-ledger/account/1",
						},
					},
					{
						Fields: map[string]any{
							"id":  "1100",
							"key": "2",
						},
						Raw: map[string]any{
							"id":   "1100",
							"key":  "2",
							"href": "/objects/general-ledger/account/2",
						},
					},
					{
						Fields: map[string]any{
							"id":  "1200",
							"key": "3",
						},
						Raw: map[string]any{
							"id":   "1200",
							"key":  "3",
							"href": "/objects/general-ledger/account/3",
						},
					},
				},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful read of customers",
			Input: common.ReadParams{ObjectName: "customer", Fields: connectors.Fields("id", "key")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/ia/api/v1/objects/accounts-receivable/customer"),
				Then:  mockserver.Response(http.StatusOK, responseCustomers),
				Else:  mockserver.Response(http.StatusNotFound, responseInvalidPath),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":  "CUST001",
							"key": "1",
						},
						Raw: map[string]any{
							"id":   "CUST001",
							"key":  "1",
							"href": "/objects/accounts-receivable/customer/1",
						},
					},
					{
						Fields: map[string]any{
							"id":  "CUST002",
							"key": "2",
						},
						Raw: map[string]any{
							"id":   "CUST002",
							"key":  "2",
							"href": "/objects/accounts-receivable/customer/2",
						},
					},
				},
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
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
	})

	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.HTTPClient().Base = mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL)

	return connector, err
}
