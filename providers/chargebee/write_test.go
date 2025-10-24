package chargebee

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

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseCustomerCreate := testutils.DataFromFile(t, "customer-create.json")
	responseCustomerUpdate := testutils.DataFromFile(t, "customer-update.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write record data is required",
			Input:        common.WriteParams{ObjectName: "customers"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Create customer successfully",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordData: map[string]any{
					"first_name": "John",
					"last_name":  "Doe",
					"email":      "john.doe@example.com",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/customers"),
					mockcond.MethodPOST(),
					mockcond.HeaderContentURLFormEncoded(),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomerCreate),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "__test__KyVnHhSBWlC1T2cj",
				Errors:   nil,
				Data: map[string]any{
					"allow_direct_debit":      false,
					"auto_collection":         "on",
					"card_status":             "no_card",
					"created_at":              float64(1517505747),
					"deleted":                 false,
					"email":                   "john.doe@example.com",
					"excess_payments":         float64(0),
					"first_name":              "John",
					"id":                      "__test__KyVnHhSBWlC1T2cj",
					"last_name":               "Doe",
					"net_term_days":           float64(0),
					"object":                  "customer",
					"pii_cleared":             "active",
					"preferred_currency_code": "USD",
					"promotional_credits":     float64(0),
					"refundable_credits":      float64(0),
					"resource_version":        float64(1517505747000),
					"taxability":              "taxable",
					"unbilled_charges":        float64(0),
					"updated_at":              float64(1517505747),
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update customer successfully",
			Input: common.WriteParams{
				ObjectName: "customers",
				RecordId:   "__test__KyVnHhSBWlC1T2cj",
				RecordData: map[string]any{
					"first_name": "Jane",
					"last_name":  "Smith",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/api/v2/customers/__test__KyVnHhSBWlC1T2cj"),
					mockcond.MethodPOST(),
					mockcond.HeaderContentURLFormEncoded(),
				},
				Then: mockserver.Response(http.StatusOK, responseCustomerUpdate),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "__test__KyVnHhSBWlC1T2cj",
				Errors:   nil,
				Data: map[string]any{
					"allow_direct_debit":      false,
					"auto_collection":         "on",
					"card_status":             "no_card",
					"created_at":              float64(1517505747),
					"deleted":                 false,
					"email":                   "john.doe@example.com",
					"excess_payments":         float64(0),
					"first_name":              "Jane",
					"id":                      "__test__KyVnHhSBWlC1T2cj",
					"last_name":               "Smith",
					"net_term_days":           float64(0),
					"object":                  "customer",
					"pii_cleared":             "active",
					"preferred_currency_code": "USD",
					"promotional_credits":     float64(0),
					"refundable_credits":      float64(0),
					"resource_version":        float64(1517505747000),
					"taxability":              "taxable",
					"unbilled_charges":        float64(0),
					"updated_at":              float64(1517505747),
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
