package salesforce

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestJobInfo(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseJobInProgress := testutils.DataFromFile(t, "bulk/info/in-progress.json")

	tests := []bulkJobInfoTestCase{
		{
			Name:  "Request fails due to internal server error",
			Input: "",
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusInternalServerError),
			}.Server(),
			ExpectedErrs: []error{common.ErrRequestFailed},
		},
		{
			Name:  "Correct endpoint is invoked",
			Input: "750ak000009Bq9OAAS",
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/services/data/v59.0/jobs/ingest/750ak000009Bq9OAAS"),
				Then:  mockserver.Response(http.StatusOK, responseJobInProgress),
			}.Server(),
			Comparator: testConciseJobInfoComparator,
			Expected: &GetJobInfoResult{
				Id:          "750ak000009Bq9OAAS",
				Object:      "Opportunity",
				State:       "InProgress",
				CreatedDate: "2024-09-09T21:32:31.000+0000",
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

func TestGetBulkQueryInfo(t *testing.T) { // nolint:dupl
	t.Parallel()

	responseAccount := testutils.DataFromFile(t, "bulk/read-launch-job-account.json")

	tests := []bulkJobInfoQueryTestCase{
		{
			Name:  "Requesting BulkQuery information invokes correct endpoint",
			Input: "750ak000009AVi5AAG",
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/services/data/v59.0/jobs/query/750ak000009AVi5AAG"),
				Then:  mockserver.Response(http.StatusOK, responseAccount),
			}.Server(),
			Comparator: testConciseJobInfoComparator,
			Expected: &GetJobInfoResult{
				Id:          "750ak000009AVi5AAG",
				Object:      "Account",
				State:       "UploadComplete",
				CreatedDate: "2024-09-09T13:08:34.000+0000",
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

func TestJobResults(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseJobPartialFailure := testutils.DataFromFile(t, "bulk/info/partial-failure.json")
	responseJobPartialFailureDescribed := testutils.DataFromFile(t, "bulk/info/partial-failure.csv")
	responseJobCompleteFailure := testutils.DataFromFile(t, "bulk/info/complete-failure.json")
	responseJobSuccess := testutils.DataFromFile(t, "bulk/info/success.json")

	tests := []bulkJobResultTestCase{
		{
			Name:  "Partial failure is parsed",
			Input: "750ak000009Dl5bAAC",
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("/services/data/v59.0/jobs/ingest/750ak000009Dl5bAAC"),
					Then: mockserver.Response(http.StatusOK, responseJobPartialFailure),
				}, {
					If:   mockcond.PathSuffix("/services/data/v59.0/jobs/ingest/750ak000009Dl5bAAC/failedResults"),
					Then: mockserver.Response(http.StatusOK, responseJobPartialFailureDescribed),
				}},
			}.Server(),
			Comparator: testJobResultsComparator,
			Expected: &JobResults{
				JobId: "750ak000009Dl5bAAC",
				State: "JobComplete",
				FailureDetails: &FailInfo{
					FailureType:   "Partial",
					FailedUpdates: make(map[string][]string),
					FailedCreates: map[string][]string{
						"INVALID_FIELD:Failed to deserialize field at col 3. " +
							"Due to, '2003-04-987654321987654321' is not a valid value " +
							"for the type xsd:date:CloseDate --": {"external-id-3"},
					},
					Reason: "",
				},
				JobInfo: nil, // this is ignored for brevity
				Message: "Some records are not processed successfully. " +
					"Please refer to the 'failureDetails' for more details.",
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Complete failure with descriptive message",
			Input: "750ak000009E1YXAA0",
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseJobCompleteFailure),
			}.Server(),
			Comparator: testJobResultsComparator,
			Expected: &JobResults{
				JobId:          "750ak000009E1YXAA0",
				State:          "Failed",
				FailureDetails: nil,
				JobInfo:        nil, // this is ignored for brevity
				Message: "No records processed successfully. " +
					"This is likely due the CSV being empty or issues with CSV column names.",
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successful info parsed from JobCompleted response",
			Input: "750ak000009BWKLAA4",
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseJobSuccess),
			}.Server(),
			Comparator: testJobResultsComparator,
			Expected: &JobResults{
				JobId:          "750ak000009BWKLAA4",
				State:          "JobComplete",
				FailureDetails: nil,
				JobInfo:        nil, // this is ignored for brevity
				Message:        "",
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

func TestGetSuccessfulJobResults(t *testing.T) { // nolint:dupl
	t.Parallel()

	tests := []bulkGetSuccessfulJobResultsTestCase{
		{
			// this guards against unexpected URL changes
			Name:  "GetSuccessfulJobResults - endpoint is invoked",
			Input: "750ak000009Dl5bAAC",
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/services/data/v59.0/jobs/ingest/750ak000009Dl5bAAC/successfulResults"),
				Then:  mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			Comparator:   statusCodeComparator,
			Expected:     &http.Response{StatusCode: http.StatusOK},
			ExpectedErrs: nil, // we expect no errors
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

func TestGetBulkQueryResults(t *testing.T) { // nolint:dupl
	t.Parallel()

	tests := []bulkGetBulkQueryResultsTestCase{
		{
			// this guards against unexpected URL changes
			Name:  "GetBulkQueryResults - endpoint is invoked",
			Input: "750ak000009Dl5bAAC",
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/services/data/v59.0/jobs/query/750ak000009Dl5bAAC/results"),
				Then:  mockserver.Response(http.StatusOK, []byte{}),
			}.Server(),
			Comparator:   statusCodeComparator,
			Expected:     &http.Response{StatusCode: http.StatusOK},
			ExpectedErrs: nil, // we expect no errors
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

func statusCodeComparator(serverURL string, actual, expected *http.Response) bool {
	return actual.StatusCode == expected.StatusCode
}

func testJobResultsComparator(serverURL string, actual, expected *JobResults) bool {
	actual.JobInfo = nil // ignore JobInfo when comparing

	return reflect.DeepEqual(actual, expected)
}

func testConciseJobInfoComparator(serverURL string, actual *GetJobInfoResult, expected *GetJobInfoResult) bool {
	return actual.Id == expected.Id &&
		actual.Object == expected.Object &&
		actual.State == expected.State &&
		actual.CreatedDate == expected.CreatedDate
}

type (
	testCaseTypeJobInfo                 = testroutines.TestCase[string, *GetJobInfoResult]
	bulkJobInfoTestCase                 testCaseTypeJobInfo
	bulkJobInfoQueryTestCase            testCaseTypeJobInfo
	testCaseTypeJobResults              = testroutines.TestCase[string, *JobResults]
	bulkJobResultTestCase               testCaseTypeJobResults
	testCaseTypeHTTPResponse            = testroutines.TestCase[string, *http.Response]
	bulkGetSuccessfulJobResultsTestCase testCaseTypeHTTPResponse
	bulkGetBulkQueryResultsTestCase     testCaseTypeHTTPResponse
)

func (c bulkJobInfoTestCase) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Connector]) {
	t.Helper()
	conn := builder.Build(t, c.Name)
	output, err := conn.GetJobInfo(context.Background(), c.Input)
	testCaseTypeJobInfo(c).Validate(t, err, output)
}

func (c bulkJobInfoQueryTestCase) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Connector]) {
	t.Helper()
	conn := builder.Build(t, c.Name)
	output, err := conn.GetBulkQueryInfo(context.Background(), c.Input)
	testCaseTypeJobInfo(c).Validate(t, err, output)
}

func (c bulkJobResultTestCase) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Connector]) {
	t.Helper()
	conn := builder.Build(t, c.Name)
	output, err := conn.GetJobResults(context.Background(), c.Input)
	testCaseTypeJobResults(c).Validate(t, err, output)
}

func (c bulkGetSuccessfulJobResultsTestCase) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Connector]) {
	t.Helper()
	conn := builder.Build(t, c.Name)

	output, err := conn.GetSuccessfulJobResults(context.Background(), c.Input)
	if err != nil {
		_ = output.Body.Close()
	}

	testCaseTypeHTTPResponse(c).Validate(t, err, output)
}

func (c bulkGetBulkQueryResultsTestCase) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Connector]) {
	t.Helper()
	conn := builder.Build(t, c.Name)

	output, err := conn.GetBulkQueryResults(context.Background(), c.Input)
	if err != nil {
		_ = output.Body.Close()
	}

	testCaseTypeHTTPResponse(c).Validate(t, err, output)
}
