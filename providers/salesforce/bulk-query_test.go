package salesforce

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestBulkQuery(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseUnknownObject := testutils.DataFromFile(t, "unknown-object.json")
	responseAccount := testutils.DataFromFile(t, "bulk/read-launch-job-account.json")

	account := &GetJobInfoResult{
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
	}

	tests := []bulkQueryTestCase{
		{
			Name: "Mime response header expected",
			Input: bulkQueryInput{
				query:          "",
				includeDeleted: false,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name: "Correct error message is understood from JSON response",
			Input: bulkQueryInput{
				query:          "",
				includeDeleted: false,
			},
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
			Input: bulkQueryInput{
				query:          "SELECT Id,Name,BillingCity FROM Account",
				includeDeleted: false,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.Body(`{
					"operation":"query",
					"query":"SELECT Id,Name,BillingCity FROM Account"}`),
				Then: mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Expected:     account,
			ExpectedErrs: nil,
		},
		{
			Name: "Include deleted items using Query All",
			Input: bulkQueryInput{
				query:          "SELECT Id,Name,BillingCity FROM Account",
				includeDeleted: true,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.PathSuffix("/services/data/v59.0/jobs/query"),
					mockcond.Body(`{
						"operation":"queryAll",
						"query":"SELECT Id,Name,BillingCity FROM Account"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Expected:     account,
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests { // nolint:dupl
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (*Connector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

type (
	bulkQueryInput struct {
		query          string
		includeDeleted bool
	}
	bulkQueryTestCaseType = testroutines.TestCase[bulkQueryInput, *GetJobInfoResult]
	bulkQueryTestCase     bulkQueryTestCaseType
)

func (c bulkQueryTestCase) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Connector]) {
	t.Helper()
	conn := builder.Build(t, c.Name)
	output, err := conn.BulkQuery(context.Background(), c.Input.query, c.Input.includeDeleted)
	bulkQueryTestCaseType(c).Validate(t, err, output)
}
