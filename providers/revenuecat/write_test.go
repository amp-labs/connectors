package revenuecat

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

func TestWrite(t *testing.T) {
	t.Parallel()

	responseCreateEntitlement := testutils.DataFromFile(t, "create-entitlement.json")
	responseUpdateEntitlement := testutils.DataFromFile(t, "update-entitlement.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "entitlements"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name: "Create entitlement via POST",
			Input: common.WriteParams{
				ObjectName: "entitlements",
				RecordData: map[string]any{
					"lookup_key":   "premium",
					"display_name": "Premium Access",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/v2/projects/proj_123/entitlements"),
					},
					Then: mockserver.Response(http.StatusCreated, responseCreateEntitlement),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "entl_abc123",
				Data: map[string]any{
					"id":           "entl_abc123",
					"lookup_key":   "premium",
					"display_name": "Premium Access",
					"project_id":   "proj_123",
				},
			},
		},
		{
			Name: "Update entitlement via PATCH",
			Input: common.WriteParams{
				ObjectName: "entitlements",
				RecordId:   "entl_abc123",
				RecordData: map[string]any{
					"display_name": "Premium Plus Access",
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.MethodPATCH(),
						mockcond.Path("/v2/projects/proj_123/entitlements/entl_abc123"),
					},
					Then: mockserver.Response(http.StatusOK, responseUpdateEntitlement),
				}},
				Default: mockserver.ResponseString(http.StatusInternalServerError, `{"error":"unexpected request"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "entl_abc123",
				Data: map[string]any{
					"id":           "entl_abc123",
					"lookup_key":   "premium",
					"display_name": "Premium Plus Access",
					"project_id":   "proj_123",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestReadConnector(tt.Server.URL, "proj_123")
			})
		})
	}
}
