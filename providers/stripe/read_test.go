package stripe

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

	errorBadRequest := testutils.DataFromFile(t, "general-errors/error-bad-request.json")
	responseEmptyAccounts := testutils.DataFromFile(t, "read/accounts/empty.json")
	responseCustomersFirstPage := testutils.DataFromFile(t, "read/customers/1-first-page.json")
	responseCustomersLastPage := testutils.DataFromFile(t, "read/customers/2-last-page.json")
	responsePaymentsExpandedCustomer := testutils.DataFromFile(t, "read/payment_intents/expand_customer.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "accounts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:     "Unknown object name is not supported",
			Input:    common.ReadParams{ObjectName: "videos", Fields: connectors.Fields("id")},
			Server:   mockserver.Dummy(),
			Expected: nil,
			ExpectedErrs: []error{
				common.ErrOperationNotSupportedForObject,
			},
		},
		{
			Name:  "Error response is understood when payload is sent for GET operation",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"Received unknown parameter: pineapple"),
			},
		},
		{
			Name: "Accounts has no records",
			Input: common.ReadParams{
				ObjectName: "accounts",
				Fields:     connectors.Fields("id"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/v1/accounts"),
				Then:  mockserver.Response(http.StatusOK, responseEmptyAccounts),
			}.Server(),
			Expected: &common.ReadResult{
				Rows:     0,
				Data:     []common.ReadResultRow{},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page is implied from customers first page",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/v1/customers"),
				Then:  mockserver.Response(http.StatusOK, responseCustomersFirstPage),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Hayley Huffman",
					},
					Raw: map[string]any{
						"id":    "cus_Rd3ODBxHt9M5xK",
						"email": "hayley.huffman@example.com",
					},
				}, {
					Fields: map[string]any{
						"name": "Linda Morgan",
					},
					Raw: map[string]any{
						"id":    "cus_Rd3NjdGWtynChD",
						"email": "linda.morgan@example.com",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v1/customers?limit=100&starting_after=cus_Rd3NjdGWtynChD",
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page is missing for the last customer page",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("name"),
				NextPage:   "/v1/customers?limit=100&starting_after=cus_Rd3NjdGWtynChD",
			},
			Comparator: testroutines.ComparatorSubsetRead,
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/v1/customers?limit=100&starting_after=cus_Rd3NjdGWtynChD"),
				Then:  mockserver.Response(http.StatusOK, responseCustomersLastPage),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"name": "Sean Foster",
					},
					Raw: map[string]any{
						"id":    "cus_Rd3NKXxTV0Hzpp",
						"email": "sean.foster@example.com",
					},
				}},
				NextPage: "",
				Done:     true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Passing associated objects will trigger expandable query of nested objects",
			Input: common.ReadParams{
				ObjectName:        "payment_intents",
				Fields:            connectors.Fields("capture_method"),
				AssociatedObjects: []string{"application", "customer"},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/v1/payment_intents"),
					mockcond.QueryParam("expand[]", "data.customer"),
					mockcond.QueryParam("expand[]", "data.application"),
				},
				Then: mockserver.Response(http.StatusOK, responsePaymentsExpandedCustomer),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"capture_method": "automatic",
					},
					Raw: map[string]any{
						"id":                 "pi_3QjoLeES6gLOjP910d0QwpzI",
						"object":             "payment_intent",
						"setup_future_usage": "off_session",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v1/payment_intents?expand%5B%5D=data.application&expand%5B%5D=data.customer&limit=100&starting_after=pi_3QjoLeES6gLOjP910d0QwpzI", // nolint:lll
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetURL(serverURL)

	return connector, nil
}
