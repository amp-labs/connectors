package braintree

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	createCustomerResponse := testutils.DataFromFile(t, "write-create-customer.json")
	updateCustomerResponse := testutils.DataFromFile(t, "write-update-customer.json")
	chargeTransactionResponse := testutils.DataFromFile(t, "write-charge-transaction.json")
	createPaymentMethodResponse := testutils.DataFromFile(t, "write-create-payment-method.json")
	updatePaymentMethodResponse := testutils.DataFromFile(t, "write-update-payment-method.json")

	// Expected request bodies for GraphQL mutations (loaded from files due to long query strings).
	createCustomerRequestBody := testutils.DataFromFile(t, "request-create-customer.json")
	updateCustomerRequestBody := testutils.DataFromFile(t, "request-update-customer.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Creating a customer with header and body validation",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordData: map[string]any{
					"firstName": "John",
					"lastName":  "Doe",
					"email":     "john.doe@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/graphql"),
					mockcond.Header(http.Header{
						braintreeVersionHeader: []string{braintreeVersion},
					}),
					mockcond.BodyBytes(createCustomerRequestBody),
				},
				Then: mockserver.Response(http.StatusOK, createCustomerResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "customer_123",
				Data: map[string]any{
					"id": "customer_123",
				},
			},
			Comparator:   testroutines.ComparatorSubsetWrite,
			ExpectedErrs: nil,
		},
		{
			Name: "Updating a customer with body validation",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordId:   "customer_123",
				RecordData: map[string]any{
					"firstName": "Jane",
					"email":     "jane.doe@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/graphql"),
					mockcond.Header(http.Header{
						braintreeVersionHeader: []string{braintreeVersion},
					}),
					mockcond.BodyBytes(updateCustomerRequestBody),
				},
				Then: mockserver.Response(http.StatusOK, updateCustomerResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "customer_123",
				Data: map[string]any{
					"id": "customer_123",
				},
			},
			Comparator:   testroutines.ComparatorSubsetWrite,
			ExpectedErrs: nil,
		},
		{
			Name: "Charging a payment method (creating transaction)",
			Input: common.WriteParams{
				ObjectName: "transactions",
				RecordData: map[string]any{
					"paymentMethodId": "pm_123",
					"amount":          "10.00",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Header(http.Header{
						braintreeVersionHeader: []string{braintreeVersion},
					}),
				},
				Then: mockserver.Response(http.StatusOK, chargeTransactionResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "txn_456",
				Data: map[string]any{
					"id": "txn_456",
				},
			},
			Comparator:   testroutines.ComparatorSubsetWrite,
			ExpectedErrs: nil,
		},
		{
			Name: "Vaulting a payment method (creating)",
			Input: common.WriteParams{
				ObjectName: "paymentMethods",
				RecordData: map[string]any{
					"paymentMethodId": "nonce_from_client",
					"customerId":      "customer_123",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Header(http.Header{
						braintreeVersionHeader: []string{braintreeVersion},
					}),
				},
				Then: mockserver.Response(http.StatusOK, createPaymentMethodResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "pm_123",
				Data: map[string]any{
					"id": "pm_123",
				},
			},
			Comparator:   testroutines.ComparatorSubsetWrite,
			ExpectedErrs: nil,
		},
		{
			Name: "Updating a payment method billing address",
			Input: common.WriteParams{
				ObjectName: "paymentMethods",
				RecordData: map[string]any{
					"paymentMethodId": "pm_123",
					"billingAddress": map[string]any{
						"streetAddress": "123 Main St",
						"locality":      "San Francisco",
						"region":        "CA",
						"postalCode":    "94105",
					},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Header(http.Header{
						braintreeVersionHeader: []string{braintreeVersion},
					}),
				},
				Then: mockserver.Response(http.StatusOK, updatePaymentMethodResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "pm_123",
				Data: map[string]any{
					"id": "pm_123",
				},
			},
			Comparator:   testroutines.ComparatorSubsetWrite,
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
