package clickup

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	genericError := testutils.DataFromFile(t, "write/generic-err.json")
	responseChannel := testutils.DataFromFile(t, "write/channel.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Generic error is processed",
			Input: common.WriteParams{ObjectName: "channels", RecordData: "dummy"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, genericError),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("Generic error message"),
			},
		},
		{
			Name:  "Create channel",
			Input: common.WriteParams{ObjectName: "channels", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/api/v3/workspaces/test-workspace-123/chat/channels"),
				},
				Then: mockserver.Response(http.StatusOK, responseChannel),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2kxv6e49-572",
				Data:     map[string]any{"name": "Collaboration"},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Update channel",
			Input: common.WriteParams{ObjectName: "channels", RecordData: "dummy", RecordId: "2kxv6e49-572"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPATCH(),
					mockcond.Path("/api/v3/workspaces/test-workspace-123/chat/channels/2kxv6e49-572"),
				},
				Then: mockserver.Response(http.StatusOK, responseChannel),
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "2kxv6e49-572",
				Data:     map[string]any{"name": "Collaboration"},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server)
			})
		})
	}
}
