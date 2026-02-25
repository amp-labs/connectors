package highlevelstandard

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

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	businessesResponse := testutils.DataFromFile(t, "businesses.json")
	calendarsGroupsResponse := testutils.DataFromFile(t, "calendars_groups.json")
	productsCollectionsResponse := testutils.DataFromFile(t, "products_collections.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of all businesses",
			Input: common.ReadParams{ObjectName: "businesses", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/businesses/"),
					mockcond.QueryParam("locationId", "iV1BEzddaWWLqU2kXhcN"),
				},
				Then: mockserver.Response(http.StatusOK, businessesResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": "67922facc1844a515a6d72e5",
						},
						Raw: map[string]any{
							"customFields": []any{},
							"name":         "Msoft",
							"locationId":   "iV1BEzddaWWLqU2kXhcN",
							"createdBy": map[string]any{
								"source":    "INTEGRATION",
								"channel":   "OAUTH",
								"sourceId":  "6789ed53f93f9a428151c88f-m60czadz",
								"createdAt": "2025-01-23T12:01:48.429Z",
							},
							"createdAt": "2025-01-23T12:01:48.430Z",
							"updatedAt": "2025-01-23T12:01:48.430Z",
							"id":        "67922facc1844a515a6d72e5",
						},
						Id: "67922facc1844a515a6d72e5",
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all calendars groups",
			Input: common.ReadParams{ObjectName: "calendars/groups", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/calendars/groups"),
					mockcond.QueryParam("locationId", "iV1BEzddaWWLqU2kXhcN"),
				},
				Then: mockserver.Response(http.StatusOK, calendarsGroupsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"id": "c5d87HDX906XNUdQD3rS",
						},
						Raw: map[string]any{
							"id":          "c5d87HDX906XNUdQD3rS",
							"name":        "test",
							"description": "test description",
							"slug":        "update",
							"isActive":    true,
							"dateAdded":   "2025-01-24T13:13:05.541Z",
							"dateUpdated": "2025-01-24T13:22:00.493Z",
						},
						Id: "c5d87HDX906XNUdQD3rS",
					},
				},
				Done: true,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all products collections",
			Input: common.ReadParams{ObjectName: "products/collections", Fields: connectors.Fields(""), NextPage: "101"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/products/collections"),
					mockcond.QueryParam("altId", "iV1BEzddaWWLqU2kXhcN"),
					mockcond.QueryParam("altType", "location"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("offset", "101"),
				},
				Then: mockserver.Response(http.StatusOK, productsCollectionsResponse),
			}.Server(),
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"_id":       "68906b97862c0c73e894b772",
							"altId":     "iV1BEzddaWWLqU2kXhcN",
							"name":      "Best Sellers",
							"slug":      "best-sellers",
							"createdAt": "2025-08-04T08:13:11.574Z",
						},
					},
				},
				NextPage: "201",
				Done:     false,
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
