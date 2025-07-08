package fathom

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

func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	createWebhooksResponse := testutils.DataFromFile(t, "create-webhooks.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},

		{
			Name: "Successfully creation of a webhook",
			Input: common.WriteParams{ObjectName: "webhooks", RecordData: map[string]any{
				"destination_url":      "https://play.svix.com/in/e_5U95s0OihUbc32B8UDA1MoAaAG2/",
				"include_action_items": true,
			}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, createWebhooksResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "ikEoQ4bVoq4JYUmc",
				Data: map[string]any{

					"id":                   "ikEoQ4bVoq4JYUmc",
					"url":                  "https://example.com/webhook",
					"secret":               "whsec_x6EV6NIAAz3ldclszNJTwrow",
					"created_at":           "2025-06-30T10:40:46Z",
					"include_transcript":   true,
					"include_crm_matches":  true,
					"include_summary":      true,
					"include_action_items": true,
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
