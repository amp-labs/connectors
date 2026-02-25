package workday

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

func TestWrite(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCreate := testutils.DataFromFile(t, "write/workers/create-response.json")
	responseUpdate := testutils.DataFromFile(t, "write/workers/update-response.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Input:        common.WriteParams{},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Create worker successfully",
			Input: common.WriteParams{
				ObjectName: "workers",
				RecordData: map[string]any{
					"descriptor":       "Logan McNeil",
					"primaryWorkEmail": "lmcneil@workday.net",
					"isManager":        true,
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/ccx/api/v1/testTenant/workers"),
				},
				Then: mockserver.Response(http.StatusCreated, responseCreate),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "3aa5550b7fe348b98d7b5741afc65534",
				Data: map[string]any{
					"id":               "3aa5550b7fe348b98d7b5741afc65534",
					"descriptor":       "Logan McNeil",
					"primaryWorkEmail": "lmcneil@workday.net",
					"isManager":        true,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Update worker successfully",
			Input: common.WriteParams{
				ObjectName: "workers",
				RecordId:   "3aa5550b7fe348b98d7b5741afc65534",
				RecordData: map[string]any{
					"descriptor": "Logan McNeil Updated",
					"isManager":  false,
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPatch),
					mockcond.Path("/ccx/api/v1/testTenant/workers/3aa5550b7fe348b98d7b5741afc65534"),
				},
				Then: mockserver.Response(http.StatusOK, responseUpdate),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "3aa5550b7fe348b98d7b5741afc65534",
				Data: map[string]any{
					"id":               "3aa5550b7fe348b98d7b5741afc65534",
					"descriptor":       "Logan McNeil Updated",
					"primaryWorkEmail": "lmcneil@workday.net",
					"isManager":        false,
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Write with empty response body",
			Input: common.WriteParams{
				ObjectName: "workers",
				RecordData: map[string]any{
					"descriptor": "Test Worker",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodPost),
					mockcond.Path("/ccx/api/v1/testTenant/workers"),
				},
				Then: mockserver.Response(http.StatusNoContent),
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
