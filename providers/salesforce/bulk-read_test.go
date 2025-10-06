// nolint:gocritic
package salesforce

import (
	"context"
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

func TestBulkRead(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseUnknownObject := testutils.DataFromFile(t, "unknown-object.json")
	responseAccount := testutils.DataFromFile(t, "bulk/read-launch-job-account.json")

	tests := []bulkReadTestCase{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "Orders"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.ReadParams{ObjectName: "Accout", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseUnknownObject),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest, errors.New("sObject type 'Accout' is not supported"), // nolint:goerr113
			},
		},
		{
			Name: "Launch bulk job with SOQL query",
			Input: common.ReadParams{
				ObjectName: "Account",
				Fields:     connectors.Fields("Id", "Name"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Or{
					// Select may have fields in different order.
					mockcond.Body(`{"operation":"query",
						"query":"SELECT Id,Name FROM Account"}`),
					mockcond.Body(`{"operation":"query",
						"query":"SELECT Name,Id FROM Account"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Expected: &GetJobInfoResult{
				Id:              "750ak000009AVi5AAG",
				Operation:       "query",
				Object:          "Account",
				CreatedById:     "005ak000005hvjJAAQ",
				CreatedDate:     "2024-09-09T13:08:34.000+0000",
				SystemModstamp:  "2024-09-09T13:08:34.000+0000",
				State:           "UploadComplete",
				ConcurrencyMode: "Parallel",
				ContentType:     "CSV",
				ApiVersion:      59.0,
				LineEnding:      "LF",
				ColumnDelimiter: "COMMA",
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (*Connector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

type (
	bulkReadTestCaseType = testroutines.TestCase[common.ReadParams, *GetJobInfoResult]
	bulkReadTestCase     bulkReadTestCaseType
)

func (c bulkReadTestCase) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Connector]) {
	t.Helper()
	conn := builder.Build(t, c.Name)
	output, err := conn.BulkRead(context.Background(), c.Input)
	bulkReadTestCaseType(c).Validate(t, err, output)
}
