package sageintacct

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

	createPaymentResponse := testutils.DataFromFile(t, "create-payment.json")
	updatePaymentSummaryResponse := testutils.DataFromFile(t, "update-payment-summary.json")

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
			Name: "Successfully creation of a payment",
			Input: common.WriteParams{ObjectName: "payment", RecordData: map[string]any{
				"paymentMethod": "cash",
				"customer": map[string]any{
					"id": "Cust-00064",
				},
				"documentNumber": "1567",
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createPaymentResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2096",
				Data: map[string]any{
					"id":   "2096",
					"key":  "2096",
					"href": "/objects/accounts-receivable/payment/2096",
				},
			},
			ExpectedErrs: nil,
		},

		{
			Name: "Successfully update a payment summary",
			Input: common.WriteParams{
				ObjectName: "payment-summary",
				RecordId:   "110",
				RecordData: map[string]any{
					"name": "Johntest updated",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, updatePaymentSummaryResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "110",
				Data: map[string]any{
					"key":  "110",
					"id":   "110",
					"href": "/objects/accounts-receivable/payment-summary/110",
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
