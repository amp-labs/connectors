package instantlyai

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	campaignsResponse := testutils.DataFromFile(t, "campaigns.json")
	customTagsResponse := testutils.DataFromFile(t, "custom-tags.json")
	leadListsResponse := testutils.DataFromFile(t, "lead-lists.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of campaigns",
			Input: common.ReadParams{ObjectName: "campaigns", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, campaignsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":                "0196d267-8d93-73c3-b563-1a6443e78eb8",
							"name":              "My First Campaign",
							"timestamp_created": "2025-05-15T05:25:23.987Z",
							"timestamp_updated": "2025-05-15T05:25:23.987Z",
							"organization":      "0196d267-8d93-73c3-b563-1a6685313ea8",
						},
					},
				},
				NextPage: testroutines.URLTestServer +
					"/v2/campaigns?limit=100&starting_after=0196d267-98dd-7720-b740-f132c2c547d9",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of custom tags",
			Input: common.ReadParams{ObjectName: "custom-tags", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, customTagsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":                "01966b52-73c2-7a5d-9df3-ea5a95a7fb09",
							"timestamp_created": "2025-04-25T05:01:27.874Z",
							"timestamp_updated": "2025-04-25T05:01:27.875Z",
							"organization_id":   "01966b52-73c3-7cc6-8081-efbe11b617f2",
							"label":             "Important",
							"description":       "Used for marking important items",
						},
					},
				},
				NextPage: testroutines.URLTestServer +
					"/v2/custom-tags?limit=100&starting_after=01966b52-881d-76de-acb0-8b12432533f6",
				Done: false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of lead lists",
			Input: common.ReadParams{ObjectName: "lead-lists", Fields: connectors.Fields("")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, leadListsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":                  "01966b52-73bf-79c0-bf01-933de40a96af",
							"organization_id":     "01966b52-73bf-79c0-bf01-933e7e9d7123",
							"has_enrichment_task": false,
							"owned_by":            "01966b52-73bf-79c0-bf01-933fa9d5415a",
							"name":                "My Lead List",
							"timestamp_created":   "2025-04-25T05:01:27.871Z",
						},
					},
				},
				NextPage: testroutines.URLTestServer +
					"/v2/lead-lists?limit=100&starting_after=01966b52-883f-760d-b7dc-cdd0fc3f7841",
				Done: false,
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
