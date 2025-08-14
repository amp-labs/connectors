package xero

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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responsePaymentOrdersFirst := testutils.DataFromFile(t, "purchase-orders-first.json")
	responsePaymentOrdersSecond := testutils.DataFromFile(t, "purchase-orders-second.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "users"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},

		{
			Name:  "Successful read with chosen fields",
			Input: common.ReadParams{ObjectName: "purchaseOrders", Fields: connectors.Fields("purchaseorderid", "purchaseordernumber", "type")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api.xro/2.0/PurchaseOrders"),
				Then:  mockserver.Response(http.StatusOK, responsePaymentOrdersFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"purchaseorderid":     "f6c52cc7-5e7f-4924-af07-ec848dd0043f",
						"purchaseordernumber": "PO-0001",
						"type":                "PURCHASEORDER",
					},
					Raw: map[string]any{
						"PurchaseOrderID":      "f6c52cc7-5e7f-4924-af07-ec848dd0043f",
						"PurchaseOrderNumber":  "PO-0001",
						"DateString":           "2025-08-14T00:00:00",
						"Date":                 "/Date(1755129600000+0000)/",
						"DeliveryDateString":   "2025-08-11T00:00:00",
						"DeliveryDate":         "/Date(1754870400000+0000)/",
						"DeliveryAddress":      "",
						"AttentionTo":          "",
						"Telephone":            "",
						"DeliveryInstructions": "",
						"HasErrors":            false,
						"IsDiscounted":         false,
						"Reference":            "feqpowejrpqw",
						"Type":                 "PURCHASEORDER",
						"CurrencyRate":         1.0000000000,
						"CurrencyCode":         "USD",
					},
				}},
				NextPage: "2", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Next page is the last page",
			Input: common.ReadParams{
				ObjectName: "purchaseOrders",
				Fields:     connectors.Fields("purchaseorderid", "purchaseordernumber", "type"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api.xro/2.0/PurchaseOrders"),
				Then:  mockserver.Response(http.StatusOK, responsePaymentOrdersSecond),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows:     1,
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
