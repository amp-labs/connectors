package heyreach

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

	listResponse := testutils.DataFromFile(t, "create_list.json")
	addLeadToCampaignAndListResponse := testutils.DataFromFile(t, "add_lead_list_campaign.json")

	tests := []testroutines.Write{
		{
			Name:         "Write object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "creating the list",
			Input: common.WriteParams{ObjectName: "list/CreateEmptyList", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, listResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "123",
				Errors:   nil,
				Data: map[string]any{
					"id":           float64(123),
					"name":         "My List",
					"count":        float64(0),
					"listType":     "COMPANY_LIST",
					"creationTime": "2024-08-29T09:34:56.5417789Z",
					"isDeleted":    false,
					"campaigns":    nil,
					"search":       nil,
					"status":       "UNKNOWN",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Add Leads to campaign",
			Input: common.WriteParams{ObjectName: "campaign/AddLeadsToCampaignV2", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, addLeadToCampaignAndListResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "0",
				Errors:   nil,
				Data: map[string]any{
					"addedLeadsCount":   float64(1),
					"updatedLeadsCount": float64(0),
					"failedLeadsCount":  float64(0),
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Add Leads to list",
			Input: common.WriteParams{ObjectName: "list/AddLeadsToListV2", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, addLeadToCampaignAndListResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "0",
				Errors:   nil,
				Data: map[string]any{
					"addedLeadsCount":   float64(1),
					"updatedLeadsCount": float64(0),
					"failedLeadsCount":  float64(0),
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Sending message to LinkedIn conversation",
			Input: common.WriteParams{ObjectName: "inbox/SendMessage", RecordData: "dummy"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
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
