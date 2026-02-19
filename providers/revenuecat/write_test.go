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

func TestWrite(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Write data must be included",
			Input:        common.WriteParams{ObjectName: "products"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},
		{
			Name:         "Only products are supported for write",
			Input:        common.WriteParams{ObjectName: "customers", RecordData: map[string]any{"x": "y"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name:         "Product updates are not supported",
			Input:        common.WriteParams{ObjectName: "products", RecordId: "prod1a2b3c4", RecordData: map[string]any{"display_name": "New"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrOperationNotSupportedForObject},
		},
		{
			Name: "Create product",
			Input: common.WriteParams{
				ObjectName: "products",
				RecordData: map[string]any{
					"store_identifier": "com.example.product.monthly",
					"app_id":           "app1a2b3c4",
					"type":             "subscription",
					"display_name":     "Premium Monthly",
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodPOST(),
					mockcond.Path("/v2/projects/proje51f9111/products"),
					mockcond.Body(`{"app_id":"app1a2b3c4","display_name":"Premium Monthly","store_identifier":"com.example.product.monthly","type":"subscription"}`),
				},
				Then: mockserver.ResponseString(http.StatusCreated, `{"id":"prod1a2b3c4d5","store_identifier":"com.example.product.monthly","app_id":"app1a2b3c4","type":"subscription","display_name":"Premium Monthly"}`),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "prod1a2b3c4d5",
				Data: map[string]any{
					"id": "prod1a2b3c4d5",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

