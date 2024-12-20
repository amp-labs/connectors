package iterable

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unknown object requested",
			Input:        []string{"butterflies"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{staticschema.ErrObjectNotFound},
		},
		{
			Name:   "Successfully describe multiple objects with metadata",
			Input:  []string{"campaigns", "messageTypes"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"campaigns": {
						DisplayName: "Campaigns",
						FieldsMap: map[string]string{
							"campaignState":       "campaignState",
							"createdAt":           "createdAt",
							"createdByUserId":     "createdByUserId",
							"endedAt":             "endedAt",
							"id":                  "id",
							"labels":              "labels",
							"listIds":             "listIds",
							"messageMedium":       "messageMedium",
							"name":                "name",
							"recurringCampaignId": "recurringCampaignId",
							"sendSize":            "sendSize",
							"startAt":             "startAt",
							"suppressionListIds":  "suppressionListIds",
							"templateId":          "templateId",
							"type":                "type",
							"updatedAt":           "updatedAt",
							"updatedByUserId":     "updatedByUserId",
							"workflowId":          "workflowId",
						},
					},
					"messageTypes": {
						DisplayName: "Message Types",
						FieldsMap: map[string]string{
							"channelId":          "channelId",
							"createdAt":          "createdAt",
							"frequencyCap":       "frequencyCap",
							"id":                 "id",
							"name":               "name",
							"rateLimitPerMinute": "rateLimitPerMinute",
							"subscriptionPolicy": "subscriptionPolicy",
							"updatedAt":          "updatedAt",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
