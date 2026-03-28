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

func TestRead(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseBudget := testutils.DataFromFile(t, "read-budget.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "account"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Successful read with chosen fields",
			Input: common.ReadParams{ObjectName: "general-ledger/budget", Fields: connectors.Fields("$['id']", "$['key']", "$['href']")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/ia/api/v1/services/core/query"),
					mockcond.Method(http.MethodPost),
				},
				Then: mockserver.Response(http.StatusOK, responseBudget),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"key": "1",
							"id":  "Std_Budget",
						},
						Raw: map[string]any{
							"key":  "1",
							"id":   "Std_Budget",
							"href": "/objects/general-ledger/budget/1",
						},
					},
					{
						Fields: map[string]any{
							"key": "2",
							"id":  "KPI_BUDGET",
						},
						Raw: map[string]any{
							"key":  "2",
							"id":   "KPI_BUDGET",
							"href": "/objects/general-ledger/budget/2",
						},
					},
				},
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
