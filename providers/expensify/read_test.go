package expensify

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

func TestRead(t *testing.T) { //nolint:funlen
	t.Parallel()

	responsePolicies := testutils.DataFromFile(t, "read-policy.json")
	responseAuthError := testutils.DataFromFile(t, "error-auth.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field must be requested",
			Input:        common.ReadParams{ObjectName: "policy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name: "Unsupported object returns operation not supported error",
			Input: common.ReadParams{
				ObjectName: "employees",
				Fields:     connectors.Fields("id", "name"),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "API authentication error is propagated",
			Input: common.ReadParams{
				ObjectName: "policy",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseAuthError),
			}.Server(),
			ExpectedErrs: []error{common.ErrRequestFailed},
		},
		{
			Name: "Successfully read policy objects with selected fields",
			Input: common.ReadParams{
				ObjectName: "policy",
				Fields:     connectors.Fields("id", "name", "outputCurrency"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, responsePolicies),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":             "POLICY_ABC123",
						"name":           "Test Company Policy",
						"outputcurrency": "USD",
					},
					Raw: map[string]any{
						"id":               "POLICY_ABC123",
						"name":             "Test Company Policy",
						"outputCurrency":   "USD",
						"type":             "team",
						"owner":            "admin@testcompany.com",
						"autoReporting":    true,
						"requiresCategory": false,
						"employees":        float64(10),
					},
				}},
				NextPage: "",
				Done:     true,
			},
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
