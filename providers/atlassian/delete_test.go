package atlassian

import (
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

func TestDelete(t *testing.T) { // nolint:funlen,cyclop,dupl
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "delete-issue-not-found.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete issue must include ID",
			Input:        common.DeleteParams{ObjectName: "issues"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Not found returned on removing missing entry",
			Input: common.DeleteParams{ObjectName: "issues", RecordId: "10010"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseErrorFormat),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Issue does not exist or you do not have permission to see it"),
			},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "issues", RecordId: "10010"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/ex/jira/ebc887b2-7e61-4059-ab35-71f15cc16e12/rest/api/3/issue/10010"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func TestDeleteConfluence(t *testing.T) { // nolint:funlen,cyclop,dupl
	t.Parallel()

	responseErrorFormat := testutils.DataFromFile(t, "confluence/delete/blogposts/err-not-found.json")

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete blogpost must include ID",
			Input:        common.DeleteParams{ObjectName: "blogposts"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Not found returned on removing missing entry",
			Input: common.DeleteParams{ObjectName: "blogposts", RecordId: "1802301"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseErrorFormat),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				errors.New("Cannot find a blogpost with id [1802301]"), // nolint:goerr113
			},
		},
		{
			Name:  "Successful delete",
			Input: common.DeleteParams{ObjectName: "blogposts", RecordId: "1802301"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodDELETE(),
				Then:  mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected:     &common.DeleteResult{Success: true},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnectorConfluence(tt.Server.URL)
			})
		})
	}
}
