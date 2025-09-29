package quickbooks

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	createCustomerResponse := testutils.DataFromFile(t, "create-customer.json")
	createPurchaseResponse := testutils.DataFromFile(t, "create-purchase.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},

		{
			Name: "Successfully creation of Customer",
			Input: common.WriteParams{ObjectName: "customer", RecordData: map[string]any{
				"FullyQualifiedName": "King Groceries",
				"Suffix":             "Jr",
				"Title":              "Mr",
				"MiddleName":         "B",
				"Notes":              "Here are other details.",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createCustomerResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "67",
				Data: map[string]any{
					"domain": "QBO",
					"PrimaryEmailAddr": map[string]any{
						"Address": "jdrew@myemail.com",
					},
					"DisplayName": "King's Groceries",
					"CurrencyRef": map[string]any{
						"name":  "United States Dollar",
						"value": "USD",
					},
					"DefaultTaxCodeRef": map[string]any{
						"value": "2",
					},
					"PreferredDeliveryMethod": "Print",
					"GivenName":               "James",
					"FullyQualifiedName":      "King's Groceries",
					"BillWithParent":          false,
					"Title":                   "Mr",
					"Job":                     false,
					"MiddleName":              "B",
					"Notes":                   "Here are other details.",
					"Active":                  true,
					"SyncToken":               "0",
					"Suffix":                  "Jr",
					"CompanyName":             "King Groceries",
					"FamilyName":              "King",
					"PrintOnCheckName":        "King Groceries",
					"sparse":                  false,
					"Id":                      "67",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully creation of Purchase",
			Input: common.WriteParams{
				ObjectName: "purchase",
				RecordId:   "055712c7-0fcf-4ba2-a900-1200f10cfe28",
				RecordData: map[string]any{
					"Name": "Eusebio Damore",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createPurchaseResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "247",
				Data: map[string]any{
					"SyncToken":   "0",
					"domain":      "QBO",
					"Credit":      false,
					"TotalAmt":    10.0,
					"PaymentType": "CreditCard",
					"TxnDate":     "2015-07-27",
					"sparse":      false,
					"AccountRef": map[string]any{
						"name":  "Visa",
						"value": "42",
					},
					"Id": "247",
					"MetaData": map[string]any{
						"CreateTime":      "2015-07-27T10:27:01-07:00",
						"LastUpdatedTime": "2015-07-27T10:27:01-07:00",
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
