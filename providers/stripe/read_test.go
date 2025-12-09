package stripe

import (
	"errors"
	"net/http"
	"testing"
	"time"

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

	errorBadRequest := testutils.DataFromFile(t, "general-errors/error-bad-request.json")
	responseEmptyAccounts := testutils.DataFromFile(t, "read/accounts/empty.json")
	responseCustomersFirstPage := testutils.DataFromFile(t, "read/customers/1-first-page.json")
	responseCustomersLastPage := testutils.DataFromFile(t, "read/customers/2-last-page.json")
	responsePaymentsExpandedCustomer := testutils.DataFromFile(t, "read/payment_intents/expand_customer.json")
	responseInvoices := testutils.DataFromFile(t, "read/invoices/incremental.json")

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
			Name:  "Error response is understood when payload is sent for GET operation",
			Input: common.ReadParams{ObjectName: "accounts", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorBadRequest),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New(
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
				If:    mockcond.Path("/v1/accounts"),
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
				If:    mockcond.Path("/v1/customers"),
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
				If:    mockcond.Path("/v1/customers?limit=100&starting_after=cus_Rd3NjdGWtynChD"),
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
					mockcond.Path("/v1/payment_intents"),
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
		{
			Name: "Incremental read of Invoices",
			Input: common.ReadParams{
				ObjectName: "invoices",
				Fields:     connectors.Fields("description"),
				Since:      time.Unix(1753116395, 0),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/v1/invoices"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("created[gte]", "1753116395"),
				},
				Then: mockserver.Response(http.StatusOK, responseInvoices),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"description": "Invoice 8887. Potatoes",
					},
					Raw: map[string]any{
						"billing_reason":    "manual",
						"collection_method": "charge_automatically",
						"customer":          "cus_Sbim60412VKvja",
						"customer_email":    "freddy.buckley@company.com",
						"customer_name":     "Freddy Buckley",
					},
				}},
				NextPage: testroutines.URLTestServer + "/v1/invoices?" +
					"created[gte]=1753116395&limit=100&starting_after=in_1RnN00ES6gLOjP91auKbmxwS",
				Done: false,
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
		WithAuthenticatedClient(mockutils.NewClient()),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
