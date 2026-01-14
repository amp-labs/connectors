package chilipiper

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// nolint
func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	unsupportedResponse := testutils.DataFromFile(t, "unsupported.txt")
	addUsers := testutils.DataFromFile(t, "add_user_team.json")
	distribution := testutils.DataFromFile(t, "distribution.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},

		{
			Name:  "Unsupported object",
			Input: common.WriteParams{ObjectName: "arsenal", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusNotFound, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name: "Successful add users in a team",
			Input: common.WriteParams{ObjectName: "team/users/add", RecordData: map[string]any{
				"teamId": "4edf8761-e5ee-48b2-81c8-c5e4849481fc",
				"userIds": []string{
					"67929af0725ce43853fd2b8c",
				},
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, addUsers),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "4edf8761-e5ee-48b2-81c8-c5e4849481fc",
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully update a distribution",
			Input: common.WriteParams{
				ObjectName: "distribution",
				RecordId:   "66d573f1bb530101b230db6f",
				RecordData: map[string]any{
					"resetDistribution": true,
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, distribution),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
