package highlevelwhitelabel

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	businessesResponse := testutils.DataFromFile(t, "write_businesses.json")
	calendarsGroupsResponse := testutils.DataFromFile(t, "write_calendars_groups.json")
	productsCollectionsResponse := testutils.DataFromFile(t, "write_products_collections.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the businesses",
			Input: common.WriteParams{ObjectName: "businesses", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, businessesResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "63771dcac1116f0e21de8e12",
				Errors:   nil,
				Data: map[string]any{
					"id":        "63771dcac1116f0e21de8e12",
					"name":      "Microsoft",
					"phone":     "string",
					"email":     "abc@microsoft.com",
					"createdAt": "2019-08-24T14:15:22Z",
					"updatedAt": "2019-08-24T14:15:22Z",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "updating the businesses",
			Input: common.WriteParams{
				ObjectName: "businesses",
				RecordData: "dummy",
				RecordId:   "63771dcac1116f0e21de8e12",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, businessesResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "63771dcac1116f0e21de8e12",
				Errors:   nil,
				Data: map[string]any{
					"id":        "63771dcac1116f0e21de8e12",
					"name":      "Microsoft",
					"phone":     "string",
					"email":     "abc@microsoft.com",
					"createdAt": "2019-08-24T14:15:22Z",
					"updatedAt": "2019-08-24T14:15:22Z",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "creating the calendars groups",
			Input: common.WriteParams{ObjectName: "calendars/groups", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, calendarsGroupsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ocQHyuzHvysMo5N5VsXc",
				Errors:   nil,
				Data: map[string]any{
					"locationId":  "ocQHyuzHvysMo5N5VsXc",
					"name":        "group a",
					"description": "group description",
					"slug":        "15-mins",
					"isActive":    true,
					"id":          "ocQHyuzHvysMo5N5VsXc",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "updating the calendars groups",
			Input: common.WriteParams{
				ObjectName: "calendars/groups",
				RecordData: "dummy",
				RecordId:   "ocQHyuzHvysMo5N5VsXc",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, calendarsGroupsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ocQHyuzHvysMo5N5VsXc",
				Errors:   nil,
				Data: map[string]any{
					"locationId":  "ocQHyuzHvysMo5N5VsXc",
					"name":        "group a",
					"description": "group description",
					"slug":        "15-mins",
					"isActive":    true,
					"id":          "ocQHyuzHvysMo5N5VsXc",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "creating the products collections",
			Input: common.WriteParams{ObjectName: "products/collections", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, productsCollectionsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "655b33a82209e60b6adb87a5",
				Errors:   nil,
				Data: map[string]any{
					"_id":       "655b33a82209e60b6adb87a5",
					"altId":     "Z4Bxl8J4SaPEPLq9IQ8g",
					"name":      "Best Sellers",
					"slug":      "best-sellers",
					"createdAt": "2024-02-22T09:27:19.728Z",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "updating the products collections",
			Input: common.WriteParams{
				ObjectName: "products/collections",
				RecordData: "dummy",
				RecordId:   "655b33a82209e60b6adb87a5",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPUT(),
				Then:  mockserver.Response(http.StatusOK, nil),
			}.Server(),
			Expected: &common.WriteResult{
				Success: true,
			},
			ExpectedErrs: nil,
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
