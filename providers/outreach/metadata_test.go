package outreach

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	schemaResponse := testutils.DataFromFile(t, "schema.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe supported objects",
			Input: []string{"opportunities"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/v2/schema.json"),
					Then: mockserver.Response(http.StatusOK, schemaResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"opportunities": {
						DisplayName: "opportunities",
						Fields: map[string]common.FieldMetadata{
							"amount": {
								DisplayName:  "amount",
								ValueType:    common.ValueTypeString,
								ProviderType: "number",
							},
							"closeDate": {
								DisplayName:  "closeDate",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"createdAt": {
								DisplayName:  "createdAt",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(true),
							},
							"currencyType": {
								DisplayName:  "currencyType",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"description": {
								DisplayName:  "description",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"mapNumberOfOverdueTasks": {
								DisplayName:  "mapNumberOfOverdueTasks",
								ValueType:    common.ValueTypeString,
								ProviderType: "integer",
							},
							"mapStatus": {
								DisplayName:  "mapStatus",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"nextStep": {
								DisplayName:  "nextStep",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"opportunityType": {
								DisplayName:  "opportunityType",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"probability": {
								DisplayName:  "probability",
								ValueType:    common.ValueTypeString,
								ProviderType: "number",
							},
							"prospectingRepId": {
								DisplayName:  "prospectingRepId",
								ValueType:    common.ValueTypeString,
								ProviderType: "integer",
							},
							"sharingTeamId": {
								DisplayName:  "sharingTeamId",
								ValueType:    common.ValueTypeString,
								ProviderType: "integer",
							},
							"tags": {
								DisplayName:  "tags",
								ValueType:    common.ValueTypeString,
								ProviderType: "array",
							},
							"territoryId": {
								DisplayName:  "territoryId",
								ValueType:    common.ValueTypeString,
								ProviderType: "integer",
							},
							"touchedAt": {
								DisplayName:  "touchedAt",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"trashedAt": {
								DisplayName:  "trashedAt",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"updatedAt": {
								DisplayName:  "updatedAt",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								ReadOnly:     goutils.Pointer(true),
							},
						},
						FieldsMap: map[string]string{
							"mapNumberOfOverdueTasks": "mapNumberOfOverdueTasks",
							"mapStatus":               "mapStatus",
							"name":                    "name",
							"nextStep":                "nextStep",
							"opportunityType":         "opportunityType",
							"probability":             "probability",
							"prospectingRepId":        "prospectingRepId",
							"sharingTeamId":           "sharingTeamId",
							"tags":                    "tags",
							"territoryId":             "territoryId",
							"touchedAt":               "touchedAt",
							"trashedAt":               "trashedAt",
							"updatedAt":               "updatedAt",
						},
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
