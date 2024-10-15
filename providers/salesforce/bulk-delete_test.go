package salesforce

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestBulkDelete(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCreateJob := testutils.DataFromFile(t, "bulk/delete/launch-job-opportunity.json")
	responseUpdateJob := testutils.DataFromFile(t, "bulk/delete/update-job-opportunity.json")

	bodyRequest := `{"object":"Opportunity","operation":"delete"}`

	tests := []bulkDeleteTestCase{
		{
			Name: "Mime response header expected",
			Input: BulkOperationParams{
				ObjectName: "Opportunity",
				CSVData:    strings.NewReader(""),
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Read object must be included",
			Input: BulkOperationParams{},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNoContent)
			})),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "CSV is required for upload",
			Input: BulkOperationParams{
				ObjectName: "Opportunity",
				CSVData:    nil,
			},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingCSVData},
		},
		{
			Name: "Successful Job Create, CSV upload, Job Update",
			Input: BulkOperationParams{
				ObjectName: "Opportunity",
				CSVData:    strings.NewReader(""),
			},
			Server: createBulkJobServer(bodyRequest, responseCreateJob, responseUpdateJob, "750ak000009BkrxAAC"),
			Expected: &BulkOperationResult{
				State: "UploadComplete",
				JobId: "750ak000009BkrxAAC",
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

type bulkDeleteTestCase bulkWriteTestCase

func (c bulkDeleteTestCase) Run(t *testing.T, builder testroutines.ConnectorBuilder[*Connector]) {
	t.Helper()
	conn := builder.Build(t, c.Name)
	output, err := conn.BulkDelete(context.Background(), c.Input)
	bulkWriteTestCaseType(c).Validate(t, err, output)
}
