package atlassian

import (
	"context"
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

func TestDelete(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "delete-issue-not-found.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Write issue must include ID",
			Input: common.DeleteParams{ObjectName: "issues"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Mime response header expected",
			Input: common.DeleteParams{ObjectName: "issues", RecordId: "10010"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusTeapot)
			})),
			ExpectedErrs: []error{interpreter.ErrMissingContentType},
		},
		{
			Name:  "Not found returned on removing missing entry",
			Input: common.DeleteParams{ObjectName: "issues", RecordId: "10010"},
			Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write(responseErrorFormat)
			})),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Issue does not exist or you do not have permission to see it"), // nolint:goerr113
			},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "issues", RecordId: "10010"},
			Server: mockserver.Reactive{
				Setup:     mockserver.ContentJSON(),
				Condition: mockcond.MethodDELETE(),
				OnSuccess: mockserver.Response(http.StatusNoContent),
			}.Server(),
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

func TestDeleteWithoutMetadata(t *testing.T) {
	t.Parallel()

	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithWorkspace("test-workspace"),
		WithModule(ModuleJira),
	)
	if err != nil {
		t.Fatal("failed to create connector")
	}

	_, err = connector.Delete(context.Background(), common.DeleteParams{ObjectName: "issues", RecordId: "123"})
	if !errors.Is(err, ErrMissingCloudId) {
		t.Fatalf("expected Delete method to complain about missing cloud id")
	}
}
