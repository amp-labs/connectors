package revenuecat

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Delete{
		{
			Name:         "Delete object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Delete object ID must be included",
			Input:        common.DeleteParams{ObjectName: "entitlements"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:  "Delete entitlement returns 200",
			Input: common.DeleteParams{ObjectName: "entitlements", RecordId: "entl_abc123"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodDELETE(),
						mockcond.Path("/v2/projects/proj_123/entitlements/entl_abc123"),
					},
					Then: mockserver.Response(http.StatusOK),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name:  "Delete product returns 204",
			Input: common.DeleteParams{ObjectName: "products", RecordId: "prod_xyz789"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodDELETE(),
						mockcond.Path("/v2/projects/proj_123/products/prod_xyz789"),
					},
					Then: mockserver.Response(http.StatusNoContent),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestReadConnector(tt.Server.URL, "proj_123")
			})
		})
	}
}
