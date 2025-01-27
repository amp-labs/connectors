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

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	errorSetupPaymentIntent := testutils.DataFromFile(t, "general-errors/error-setup-payment-intent.json")
	responseCreateCustomer := testutils.DataFromFile(t, "write/customers/new.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Bad request from provider",
			Input: common.WriteParams{
				ObjectName: "payment_intents",
				RecordId:   "9891d452-56fe-4397-b431-a92e79cdc980",
				RecordData: make(map[string]any),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, errorSetupPaymentIntent),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"Missing required param: mandate_data.",
				),
			},
		},
		{
			Name:  "Write must act as a Create",
			Input: common.WriteParams{ObjectName: "customers", RecordData: make(map[string]any)},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/v1/customers"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Write must act as an Update",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordId:   "cus_ReFi6S3vuH5Lea",
				RecordData: make(map[string]any),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/v1/customers/cus_ReFi6S3vuH5Lea"),
				},
				Then: mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of a customer",
			Input: common.WriteParams{ObjectName: "customers", RecordData: make(map[string]any)},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.PathSuffix("/v1/customers"),
					mockcond.HeaderContentURLFormEncoded(),
				},
				Then: mockserver.Response(http.StatusOK, responseCreateCustomer),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "cus_RfIicOKc1UqLHJ",
				Errors:   nil,
				Data: map[string]any{
					"id":    "cus_RfIicOKc1UqLHJ",
					"name":  "DarkBlueLiterature",
					"email": "leraquitzon@kovacek.com",
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
