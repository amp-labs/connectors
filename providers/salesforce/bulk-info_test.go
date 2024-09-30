package salesforce

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestJobInfo(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseJobInProgress := testutils.DataFromFile(t, "bulk/info/in-progress.json")

	tests := []bulkJobInfoTestCase{
		{
			Name:         "Mime response header expected",
			Input:        "",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Request fails due to internal server error",
			Input: "",
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
			})),
			ExpectedErrs: []error{common.ErrRequestFailed},
		},
		{
			Name:  "Correct endpoint is invoked",
			Input: "750ak000009Bq9OAAS",
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // nolint:varnamelen
				w.Header().Set("Content-Type", "application/json")
				switch path := r.URL.Path; {
				case strings.HasSuffix(path, "/services/data/v59.0/jobs/ingest/750ak000009Bq9OAAS"):
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseJobInProgress)
				default:
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte{})
				}
			})),
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
		tt := tt // rebind, omit loop side effects for parallel goroutine
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // nolint:varnamelen
				w.Header().Set("Content-Type", "application/json")
				switch path := r.URL.Path; {
				case strings.HasSuffix(path, "/services/data/v59.0/jobs/query/750ak000009AVi5AAG"):
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseAccount)
				default:
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte{})
				}
			})),
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
		tt := tt // rebind, omit loop side effects for parallel goroutine
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // nolint:varnamelen
				w.Header().Set("Content-Type", "application/json")
				switch path := r.URL.Path; {
				case strings.HasSuffix(path, "/services/data/v59.0/jobs/ingest/750ak000009Dl5bAAC"):
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseJobPartialFailure)

				case strings.HasSuffix(path, "/services/data/v59.0/jobs/ingest/750ak000009Dl5bAAC/failedResults"):
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write(responseJobPartialFailureDescribed)

				default:
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte{})
				}
			})),
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // nolint:varnamelen
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseJobCompleteFailure)
			})),
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // nolint:varnamelen
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(responseJobSuccess)
			})),
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
		tt := tt // rebind, omit loop side effects for parallel goroutine
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
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch path := r.URL.Path; {
				case strings.HasSuffix(path,
					"/services/data/v59.0/jobs/ingest/750ak000009Dl5bAAC/successfulResults"):
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte{})
				default:
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte{})
				}
			})),
			Comparator:   statusCodeComparator,
			Expected:     &http.Response{StatusCode: http.StatusOK},
			ExpectedErrs: nil, // we expect no errors
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

func TestGetBulkQueryResults(t *testing.T) { // nolint:dupl
	t.Parallel()

	tests := []bulkGetBulkQueryResultsTestCase{
		{
			// this guards against unexpected URL changes
			Name:  "GetBulkQueryResults - endpoint is invoked",
			Input: "750ak000009Dl5bAAC",
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch path := r.URL.Path; {
				case strings.HasSuffix(path,
					"/services/data/v59.0/jobs/query/750ak000009Dl5bAAC/results"):
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte{})
				default:
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte{})
				}
			})),
			Comparator:   statusCodeComparator,
			Expected:     &http.Response{StatusCode: http.StatusOK},
			ExpectedErrs: nil, // we expect no errors
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
