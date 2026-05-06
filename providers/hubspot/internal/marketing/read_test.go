package marketing

import (
	"net/http"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) {
	t.Parallel()

	responseCampaignsFirst := testutils.DataFromFile(t, "read/campaigns/1-first-page.json")
	responseCampaignsLast := testutils.DataFromFile(t, "read/campaigns/2-last-page.json")

	tests := []testroutines.Read{
		{
			Name: "Read campaigns first page",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("hs_name", "hs_notes", "hs_budget_items_sum_amount"),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/marketing/campaigns/2026-03"),
					mockcond.QueryParam("limit", "100"),
					mockcond.QueryParam("sort", "-updatedAt"),
				},
				Then: mockserver.Response(http.StatusOK, responseCampaignsFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Fields: map[string]any{
						"hs_name":                    "Nurture",
						"hs_notes":                   "Creating campaign from the Dashboard",
						"hs_budget_items_sum_amount": "2.0",
					},
					Raw: map[string]any{
						"id": "84f199fa-beb7-4dca-ad94-3d778cdce157",
						"properties": map[string]any{
							"hs_name":                    "Nurture",
							"hs_notes":                   "Creating campaign from the Dashboard",
							"hs_budget_items_sum_amount": "2.0",
						},
						"createdAt": "2026-05-05T23:41:20.330Z",
						"updatedAt": "2026-05-05T23:45:04.200Z",
					},
					Id: "84f199fa-beb7-4dca-ad94-3d778cdce157",
				}, {
					Fields: map[string]any{
						"hs_name": "Kiwi",
					},
					Raw: map[string]any{
						"id": "fc4583f7-5cfc-4773-8fa4-076cd4f4ae6d",
						"properties": map[string]any{
							"hs_name": "Kiwi",
						},
						"createdAt": "2026-05-05T23:09:27.549Z",
						"updatedAt": "2026-05-05T23:09:27.713Z",
					},
					Id: "fc4583f7-5cfc-4773-8fa4-076cd4f4ae6d",
				}},
				NextPage: "https://api.hubapi.com/marketing/campaigns/2026-03?limit=2&sort=-updatedAt&properties=hs_name%2Chs_notes%2Chs_budget_items_sum_amount&after=Mg%3D%3D",
				Done:     false,
			},
		},
		{
			Name: "Read campaigns with connector side filtering",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("hs_name"),
				// The first item will be returned, last filtered out.
				// Due to the sort order there will be no next page.
				// The record which is excluded has this timestamp: 2026-05-05T23:09:27.713Z
				Since: time.Date(2026, 5, 5, 23, 10, 0, 0, time.UTC),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/marketing/campaigns/2026-03"),
				},
				Then: mockserver.Response(http.StatusOK, responseCampaignsFirst),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"hs_name": "Nurture",
						},
						Raw: map[string]any{
							"updatedAt": "2026-05-05T23:45:04.200Z",
						},
						Id: "84f199fa-beb7-4dca-ad94-3d778cdce157",
					},
				},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name: "Read campaigns last page",
			Input: common.ReadParams{
				ObjectName: "campaigns",
				Fields:     connectors.Fields("hs_name"),
				NextPage:   testroutines.URLTestServer + "/marketing/campaigns/2026-03?limit=2&sort=-updatedAt&properties=hs_name%2Chs_notes%2Chs_budget_items_sum_amount&after=Mg%3D%3D",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/marketing/campaigns/2026-03"),
					mockcond.QueryParam("after", "Mg=="),
				},
				Then: mockserver.Response(http.StatusOK, responseCampaignsLast),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{
							"hs_name": "Inbound",
						},
						Raw: map[string]any{
							"createdAt": "2026-05-05T23:07:11.797Z",
							"updatedAt": "2026-05-05T23:07:12.040Z",
						},
						Id: "5f7bff76-193f-43af-968b-f13c6576ca76",
					},
				},
				NextPage: "",
				Done:     true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestAdapter(tt.Server.URL)
			})
		})
	}
}
