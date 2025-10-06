package chilipiper

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	zeroRecords := testutils.DataFromFile(t, "empty.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.txt")
	team := testutils.DataFromFile(t, "team.json")

	tests := []testroutines.Read{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is required",
			Input:        common.ReadParams{ObjectName: "deals"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unsupported object",
			Input: common.ReadParams{ObjectName: "meme", Fields: datautils.NewStringSet("name")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:  "Zero records response",
			Input: common.ReadParams{ObjectName: "distribution", Fields: connectors.Fields("assistant")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, zeroRecords),
			}.Server(),
			Expected:     &common.ReadResult{Rows: 0, Data: []common.ReadResultRow{}, Done: true},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully Read Teams",
			Input: common.ReadParams{
				ObjectName: "team",
				Fields:     connectors.Fields("id", "name"),
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, team),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"id":   "4edf8761-e5ee-48b2-81c8-c5e4849481fc",
						"name": "Dev Team",
					},
					Raw: map[string]any{
						"id": "4edf8761-e5ee-48b2-81c8-c5e4849481fc",
						"members": []any{
							"67929af0725ce43853fd2b8c",
						},
						"metadata": map[string]any{
							"createdAt": "2025-01-24T09:37:47.321631Z",
							"createdBy": "user/67929af0725ce43853fd2b8c",
							"revision":  float64(0),
							"teamMembersMetadata": map[string]any{
								"addedAt": map[string]any{
									"67929af0725ce43853fd2b8c": "2025-01-24T09:37:47.321631Z",
								},
							},
							"updatedAt": "2025-01-24T09:37:47.321631Z",
							"updatedBy": "user/67929af0725ce43853fd2b8c",
						},
						"name":        "Dev Team",
						"workspaceId": "cad33722-df27-4691-bc11-1f2c89c1dd31",
					},
				}},
				Done: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
