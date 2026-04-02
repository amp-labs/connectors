package housecallpro

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
			Name: "Delete price book material successfully",
			Input: common.DeleteParams{
				ObjectName: "price_book/materials",
				RecordId:   "mat_8c92ab4f1e",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodDelete),
					mockcond.Path("/api/price_book/materials/mat_8c92ab4f1e"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Delete material category successfully",
			Input: common.DeleteParams{
				ObjectName: "price_book/material_categories",
				RecordId:   "pbmcat_db619b02f05d40d79470a38fc50332db",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodDelete),
					mockcond.Path("/api/price_book/material_categories/pbmcat_db619b02f05d40d79470a38fc50332db"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Delete price form successfully",
			Input: common.DeleteParams{
				ObjectName: "price_book/price_forms",
				RecordId:   "pf_12345abcd",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Method(http.MethodDelete),
					mockcond.Path("/api/price_book/price_forms/pf_12345abcd"),
				},
				Then: mockserver.Response(http.StatusNoContent),
			}.Server(),
			Expected: &common.DeleteResult{
				Success: true,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
