package braintree

import (
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	customersResponse := testutils.DataFromFile(t, "read-customers.json")
	transactionsResponse := testutils.DataFromFile(t, "read-transactions.json")
	disputesResponse := testutils.DataFromFile(t, "read-disputes.json")
	verificationsResponse := testutils.DataFromFile(t, "read-verifications.json")
	merchantAccountsResponse := testutils.DataFromFile(t, "read-merchant_accounts.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Successful read of customers",
			Input: common.ReadParams{
				ObjectName: "customers",
				Fields:     connectors.Fields("id", "email"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, customersResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":    "customer_123",
							"email": "john.doe@example.com",
						},
						Raw: map[string]any{
							"node": map[string]any{
								"id":          "customer_123",
								"legacyId":    "123456",
								"email":       "john.doe@example.com",
								"firstName":   "John",
								"lastName":    "Doe",
								"createdAt":   "2025-01-15T10:30:00Z",
								"company":     "Acme Corp",
								"phoneNumber": "+1234567890",
							},
						},
					},
					{
						Fields: map[string]any{
							"id":    "customer_456",
							"email": "jane.smith@example.com",
						},
						Raw: map[string]any{
							"node": map[string]any{
								"id":          "customer_456",
								"legacyId":    "456789",
								"email":       "jane.smith@example.com",
								"firstName":   "Jane",
								"lastName":    "Smith",
								"createdAt":   "2025-01-16T14:20:00Z",
								"company":     "Tech Inc",
								"phoneNumber": "+0987654321",
							},
						},
					},
				},
				Done: true,
			},
			Comparator:   testroutines.ComparatorSubsetRead,
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read of transactions",
			Input: common.ReadParams{
				ObjectName: "transactions",
				Fields:     connectors.Fields("id", "status"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, transactionsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":     "txn_123",
							"status": "settled",
						},
						Raw: map[string]any{
							"node": map[string]any{
								"id":        "txn_123",
								"legacyId":  "tx123456",
								"status":    "settled",
								"amount":    map[string]any{"value": "100.00", "currencyCode": "USD"},
								"customer":  map[string]any{"id": "customer_123", "email": "john.doe@example.com"},
								"orderId":   "order_789",
								"createdAt": "2025-01-20T09:15:00Z",
							},
						},
					},
					{
						Fields: map[string]any{
							"id":     "txn_456",
							"status": "authorized",
						},
						Raw: map[string]any{
							"node": map[string]any{
								"id":        "txn_456",
								"legacyId":  "tx456789",
								"status":    "authorized",
								"amount":    map[string]any{"value": "50.00", "currencyCode": "USD"},
								"customer":  map[string]any{"id": "customer_456", "email": "jane.smith@example.com"},
								"orderId":   "order_012",
								"createdAt": "2025-01-21T11:30:00Z",
							},
						},
					},
				},
				Done: true,
			},
			Comparator:   testroutines.ComparatorSubsetRead,
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read of disputes",
			Input: common.ReadParams{
				ObjectName: "disputes",
				Fields:     connectors.Fields("id", "status"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, disputesResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":     "dispute_123",
							"status": "OPEN",
						},
						Raw: map[string]any{
							"node": map[string]any{
								"id":              "dispute_123",
								"legacyId":        "dp123456",
								"createdAt":       "2025-01-20T09:15:00Z",
								"amountDisputed":  map[string]any{"value": "100.00", "currencyCode": "USD"},
								"status":          "OPEN",
								"type":            "CHARGEBACK",
								"caseNumber":      "CASE123",
								"referenceNumber": "REF123",
							},
						},
					},
					{
						Fields: map[string]any{
							"id":     "dispute_456",
							"status": "WON",
						},
						Raw: map[string]any{
							"node": map[string]any{
								"id":              "dispute_456",
								"legacyId":        "dp456789",
								"createdAt":       "2025-01-21T11:30:00Z",
								"amountDisputed":  map[string]any{"value": "50.00", "currencyCode": "USD"},
								"status":          "WON",
								"type":            "RETRIEVAL",
								"caseNumber":      "CASE456",
								"referenceNumber": "REF456",
							},
						},
					},
				},
				Done: true,
			},
			Comparator:   testroutines.ComparatorSubsetRead,
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read of verifications",
			Input: common.ReadParams{
				ObjectName: "verifications",
				Fields:     connectors.Fields("id", "status"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, verificationsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":     "verification_123",
							"status": "VERIFIED",
						},
						Raw: map[string]any{
							"node": map[string]any{
								"id":                "verification_123",
								"legacyId":          "vf123456",
								"createdAt":         "2025-01-20T09:15:00Z",
								"status":            "VERIFIED",
								"merchantAccountId": "merchant_123",
								"processorResponse": map[string]any{"legacyCode": "1000", "message": "Approved"},
							},
						},
					},
					{
						Fields: map[string]any{
							"id":     "verification_456",
							"status": "GATEWAY_REJECTED",
						},
						Raw: map[string]any{
							"node": map[string]any{
								"id":                "verification_456",
								"legacyId":          "vf456789",
								"createdAt":         "2025-01-21T11:30:00Z",
								"status":            "GATEWAY_REJECTED",
								"merchantAccountId": "merchant_456",
								"processorResponse": map[string]any{"legacyCode": "2000", "message": "Processor Declined"},
							},
						},
					},
				},
				Done: true,
			},
			Comparator:   testroutines.ComparatorSubsetRead,
			ExpectedErrs: nil,
		},
		{
			Name: "Successful read of merchant_accounts",
			Input: common.ReadParams{
				ObjectName: "merchant_accounts",
				Fields:     connectors.Fields("id", "dbaName"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, merchantAccountsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id":      "merchant_account_123",
							"dbaname": "Acme Corp",
						},
						Raw: map[string]any{
							"node": map[string]any{
								"id":           "merchant_account_123",
								"currencyCode": "USD",
								"dbaName":      "Acme Corp",
							},
						},
					},
					{
						Fields: map[string]any{
							"id":      "merchant_account_456",
							"dbaname": "Acme Europe",
						},
						Raw: map[string]any{
							"node": map[string]any{
								"id":           "merchant_account_456",
								"currencyCode": "EUR",
								"dbaName":      "Acme Europe",
							},
						},
					},
				},
				Done: true,
			},
			Comparator:   testroutines.ComparatorSubsetRead,
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
