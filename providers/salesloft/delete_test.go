package salesloft

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/interpreter"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	listSchema := testutils.DataFromFile(t, "write-signals-error.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete param object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete param object and its ID must be included",
			Input:        common.DeleteParams{ObjectName: "signals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:         "Mime response header expected",
			Input:        common.DeleteParams{ObjectName: "signals", RecordId: "22165"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Correct error message is understood from JSON response",
			Input: common.DeleteParams{ObjectName: "signals", RecordId: "22165"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write(listSchema)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("no Signal Registration found for integration id 5167 and given type"), // nolint:goerr113
			},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "signals", RecordId: "22165"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				mockutils.RespondNoContentForMethod(w, r, "DELETE")
			})),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
