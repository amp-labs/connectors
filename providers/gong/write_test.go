package gong

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseInvalidRequest := testutils.DataFromFile(t, "write-invalid-request.json")
	responseCreateCall := testutils.DataFromFile(t, "write-call.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "calls"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Unsupported object name",
			Input:        common.WriteParams{ObjectName: "butterflies", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.WriteParams{ObjectName: "calls", RecordData: "dummy"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Error on invalid json request",
			Input: common.WriteParams{ObjectName: "calls", RecordData: "dummy"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write(responseInvalidRequest)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New( // nolint:goerr113
					"parties: must not be null, direction: must not be null, actualStart: must not be null",
				),
			},
		},
		{
			Name:  "Valid creation of a call",
			Input: common.WriteParams{ObjectName: "calls", RecordData: "dummy"},
			Server: mockserver.Reactive{
				Setup:     mockserver.ContentJSON(),
				Condition: mockcond.MethodPOST(),
				OnSuccess: mockserver.Response(http.StatusOK, responseCreateCall),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "1102881687159885703",
				Errors:   nil,
				Data:     nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
