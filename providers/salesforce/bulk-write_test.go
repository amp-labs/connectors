package salesforce

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestBulkWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCreateJob := testutils.DataFromFile(t, "bulk/write/launch-job-opportunity.json")
	responseUpdateJob := testutils.DataFromFile(t, "bulk/write/update-job-opportunity.json")

	bodyRequest := `{
		"contentType":"CSV",
		"externalIdFieldName":"external_id__c",
		"lineEnding":"LF",
		"object":"Opportunity",
		"operation":"upsert"
	}`

	tests := []bulkWriteTestCase{
		{
			Name: "Mime response header expected",
			Input: BulkOperationParams{
				ObjectName:      "Opportunity",
				ExternalIdField: "fieldName8",
				CSVData:         strings.NewReader(""),
				Mode:            UpsertMode,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name: "Read object must be included",
			Input: BulkOperationParams{
				Mode: UpsertMode,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNoContent),
			}.Server(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Delete mode cannot be used in BulkWrite",
			Input: BulkOperationParams{
				ObjectName: "Opportunity",
				Mode:       DeleteMode,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrUnsupportedMode},
		},
		{
			Name: "Upsert requires External ID",
			Input: BulkOperationParams{
				ObjectName: "Opportunity",
				Mode:       UpsertMode,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrExternalIdEmpty},
		},
		{
			Name: "CSV is required for upload",
			Input: BulkOperationParams{
				ObjectName:      "Opportunity",
				ExternalIdField: "external_id__c",
				CSVData:         nil,
				Mode:            UpsertMode,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingCSVData},
		},
		{
			Name: "Creating Job fails on bad response",
			Input: BulkOperationParams{
				ObjectName:      "Opportunity",
				ExternalIdField: "external_id__c",
				CSVData:         strings.NewReader(""),
				Mode:            UpsertMode,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusInternalServerError, []byte{}),
			}.Server(),
			ExpectedErrs: []error{ErrCreateJob},
		},
		{
			Name: "Create job id must be string",
			Input: BulkOperationParams{
				ObjectName:      "Opportunity",
				ExternalIdField: "external_id__c",
				CSVData:         strings.NewReader(""),
				Mode:            UpsertMode,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{"id":true, "state":"Open"}`),
			}.Server(),
			ExpectedErrs: []error{common.ErrParseError},
		},
		{
			Name: "Created job must have 'Open' state",
			Input: BulkOperationParams{
				ObjectName:      "Opportunity",
				ExternalIdField: "external_id__c",
				CSVData:         strings.NewReader(""),
				Mode:            UpsertMode,
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.ResponseString(http.StatusOK, `{"id":"132", "state":"UploadComplete"}`),
			}.Server(),
			ExpectedErrs: []error{ErrInvalidJobState},
		},
		{
			Name: "Server rejects CSV upload with internal server error",
			Input: BulkOperationParams{
				ObjectName:      "Opportunity",
				ExternalIdField: "external_id__c",
				CSVData:         strings.NewReader(""),
				Mode:            UpsertMode,
			},
			Server: mockserver.Reactive{
				Setup: mockserver.ContentJSON(),
				Condition: mockcond.And{
					mockcond.PathSuffix("/services/data/v59.0/jobs/ingest"),
					mockcond.Body(bodyRequest),
				},
				OnSuccess: mockserver.Response(http.StatusOK, responseCreateJob),
			}.Server(),
			ExpectedErrs: []error{ErrCSVUploadFailure},
		},
		{
			Name: "Updating Job status fails",
			Input: BulkOperationParams{
				ObjectName:      "Opportunity",
				ExternalIdField: "external_id__c",
				CSVData:         strings.NewReader(""),
				Mode:            UpsertMode,
			},
			Server:       createBulkJobServer(bodyRequest, responseCreateJob, []byte(`{...]`), "750ak000009BWKLAA4"),
			ExpectedErrs: []error{ErrUpdateJob},
		},
		{
			Name: "Successful Job Create, CSV upload, Job Update",
			Input: BulkOperationParams{
				ObjectName:      "Opportunity",
				ExternalIdField: "external_id__c",
				CSVData:         strings.NewReader(""),
				Mode:            UpsertMode,
			},
			Server: createBulkJobServer(bodyRequest, responseCreateJob, responseUpdateJob, "750ak000009BWKLAA4"),
			Expected: &BulkOperationResult{
				State: "UploadComplete",
				JobId: "750ak000009BWKLAA4",
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

// This server knows how to respond to the following events:
// * Job created for certain request body.
// * CSV was uploaded using JobID.
// * Updated Job status as upload completed.
// Otherwise, responds with internal server error that will break the tests which is intended.
func createBulkJobServer(
	bodyRequest string,
	responseCreateJob []byte,
	responseUpdateJob []byte,
	jobID string,
) *httptest.Server {
	return mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{ // Create job if body matches.
				If: mockcond.And{
					mockcond.PathSuffix("/services/data/v59.0/jobs/ingest"),
					mockcond.Body(bodyRequest),
				},
				Then: mockserver.Response(http.StatusOK, responseCreateJob),
			},
			{ // We expect CSV to be uploaded via this endpoint.
				If:   mockcond.PathSuffix(fmt.Sprintf("/services/data/v59.0/jobs/ingest/%v/batches", jobID)),
				Then: mockserver.Response(http.StatusCreated, []byte{}),
			},
			{ // Mark job Completed.
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.PathSuffix(fmt.Sprintf("/services/data/v59.0/jobs/ingest/%v", jobID)),
					mockcond.Body(`{"state":"UploadComplete"}`),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdateJob),
			},
		},
	}.Server()
}

type (
	bulkWriteTestCaseType = testroutines.TestCase[BulkOperationParams, *BulkOperationResult]
	bulkWriteTestCase     bulkWriteTestCaseType
)

func (c bulkWriteTestCase) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Connector]) {
	t.Helper()
	conn := builder.Build(t, c.Name)
	output, err := conn.BulkWrite(context.Background(), c.Input)
	bulkWriteTestCaseType(c).Validate(t, err, output)
}
