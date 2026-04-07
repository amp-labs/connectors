package blackbaud

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

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	currenciesResponse := testutils.DataFromFile(t, "currencies.json")
	volunteersResponse := testutils.DataFromFile(t, "volunteers.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of currencies",
			Input: common.ReadParams{ObjectName: "crm-adnmg/currencies", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/crm-adnmg/currencies/list"),
				Then:  mockserver.Response(http.StatusOK, currenciesResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": "6802eae7-b10d-49b1-8ffd-ec5611d69af8",
						},
						Raw: map[string]any{
							"id":                    "6802eae7-b10d-49b1-8ffd-ec5611d69af8",
							"name":                  "US Dollar",
							"iso_4217":              "USD",
							"locale":                "United States",
							"decimal_digits":        float64(2),
							"currency_symbol":       "$",
							"rounding_type":         "Half rounds away from zero",
							"active":                true,
							"organization_currency": true,
						},
						Id: "6802eae7-b10d-49b1-8ffd-ec5611d69af8",
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of volunteers",
			Input: common.ReadParams{ObjectName: "crm-volmg/volunteers", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/crm-volmg/volunteers/search"),
				Then:  mockserver.Response(http.StatusOK, volunteersResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": "9ae04de3-0366-45b1-bdfa-3753ae64fc3f",
						},
						Raw: map[string]any{
							"id":                    "9ae04de3-0366-45b1-bdfa-3753ae64fc3f",
							"name":                  "Kyle Abrahms",
							"address":               "22 Baker Street",
							"city":                  "Charleston",
							"state":                 "South Carolina",
							"post_code":             "29964",
							"lookup_id":             "8-10000685",
							"constituent_type":      "Individual",
							"sort_constituent_name": "Abrahms, Kyle",
						},
						Id: "9ae04de3-0366-45b1-bdfa-3753ae64fc3f",
					},
				},
				Done: true,
			},
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
