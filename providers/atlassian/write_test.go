package atlassian

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { // nolint:funlen,cyclop
	t.Parallel()

	responseInvalidProjectError := testutils.DataFromFile(t, "create-issue-invalid-project.json")
	responseInvalidTypeError := testutils.DataFromFile(t, "create-issue-invalid-type.json")
	createIssueResponse := testutils.DataFromFile(t, "create-issue.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write needs data payload",
			Input:        common.WriteParams{ObjectName: "issues"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:  "Error missing project during write",
			Input: common.WriteParams{ObjectName: "issues", RecordId: "10003", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseInvalidProjectError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("project:Specify a valid project ID or key"), // nolint:goerr113
			},
		},
		{
			Name:  "Error missing issue type during write",
			Input: common.WriteParams{ObjectName: "issues", RecordId: "10003", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseInvalidTypeError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("issuetype:Specify an issue type"), // nolint:goerr113
			},
		},
		{
			Name:  "Write must act as a Create",
			Input: common.WriteParams{ObjectName: "issues", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Write must act as an Update",
			Input: common.WriteParams{ObjectName: "issues", RecordId: "10003", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.WriteResult{Success: true},
			ExpectedErrs: nil,
		},
		{
			Name:  "Valid creation of an Issue",
			Input: common.WriteParams{ObjectName: "issues", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createIssueResponse),
			}.Server(),
			Comparator: func(serverURL string, actual, expected *common.WriteResult) bool {
				return mockutils.WriteResultComparator.SubsetData(actual, expected)
			},
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "10004",
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

func TestWriteWithoutMetadata(t *testing.T) {
	t.Parallel()

	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithWorkspace("test-workspace"),
		WithModule(ModuleJira),
	)
	if err != nil {
		t.Fatal("failed to create connector")
	}

	_, err = connector.Write(context.Background(), common.WriteParams{ObjectName: "issues", RecordData: "dummy"})
	if !errors.Is(err, ErrMissingCloudId) {
		t.Fatalf("expected Write method to complain about missing cloud id")
	}
}
