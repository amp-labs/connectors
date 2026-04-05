package fastspring

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
			Name:         "Object name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Record id is required",
			Input:        common.DeleteParams{ObjectName: "products"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordID},
		},
		{
			Name:         "Unsupported object",
			Input:        common.DeleteParams{ObjectName: "accounts", RecordId: "x"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Delete product",
			Input: common.DeleteParams{
				ObjectName: "products",
				RecordId:   "my-product-path",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/products/my-product-path"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{"products":[{"product":"my-product-path","result":"success"}]}`)),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
		{
			Name: "Cancel subscription",
			Input: common.DeleteParams{
				ObjectName: "subscriptions",
				RecordId:   "sub_abc",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodDELETE(),
					mockcond.Path("/subscriptions/sub_abc"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(`{}`)),
			}.Server(),
			Expected: &common.DeleteResult{Success: true},
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.DeleteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
