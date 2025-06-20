package lever

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	opportunitiesResponse := testutils.DataFromFile(t, "opportunities.json")
	requisitionFieldsResponse := testutils.DataFromFile(t, "requisition_fields.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"opportunities", "requisition_fields"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v1/opportunities"),
					Then: mockserver.Response(http.StatusOK, opportunitiesResponse),
				}, {
					If:   mockcond.Path("/v1/requisition_fields"),
					Then: mockserver.Response(http.StatusOK, requisitionFieldsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"opportunities": {
						DisplayName: "Opportunities",
						Fields:      nil,
						FieldsMap: map[string]string{
							"id":                  "id",
							"name":                "name",
							"contact":             "contact",
							"headline":            "headline",
							"stage":               "stage",
							"confidentiality":     "confidentiality",
							"location":            "location",
							"phones":              "phones",
							"emails":              "emails",
							"links":               "links",
							"archived":            "archived",
							"tags":                "tags",
							"sources":             "sources",
							"stageChanges":        "stageChanges",
							"origin":              "origin",
							"sourcedBy":           "sourcedBy",
							"owner":               "owner",
							"followers":           "followers",
							"applications":        "applications",
							"createdAt":           "createdAt",
							"updatedAt":           "updatedAt",
							"lastInteractionAt":   "lastInteractionAt",
							"lastAdvancedAt":      "lastAdvancedAt",
							"snoozedUntil":        "snoozedUntil",
							"urls":                "urls",
							"isAnonymized":        "isAnonymized",
							"dataProtection":      "dataProtection",
							"opportunityLocation": "opportunityLocation",
						},
					},
					"requisition_fields": {
						DisplayName: "Requisition_fields",
						Fields:      nil,
						FieldsMap: map[string]string{
							"id":         "id",
							"text":       "text",
							"type":       "type",
							"isRequired": "isRequired",
						},
					},
				},
				Errors: nil,
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Metadata: map[string]string{
			"opportunityId": "2087af84-f146-4535-9368-2309e33e049f",
		},
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
