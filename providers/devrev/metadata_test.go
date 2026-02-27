package devrev

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successfully describe articles and commands",
			Input:      []string{"articles", "commands"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"articles": {
						DisplayName: "Articles",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"description": {
								DisplayName:  "description",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"created_date": {
								DisplayName:  "created_date",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"article_type": {
								DisplayName:  "article_type",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "article", DisplayValue: "article"},
									{Value: "content_block", DisplayValue: "content_block"},
									{Value: "page", DisplayValue: "page"},
								},
							},
						},
					},
					"commands": {
						DisplayName: "Commands",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"display_name": {
								DisplayName:  "display_name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"created_date": {
								DisplayName:  "created_date",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								Values:       nil,
							},
							"status": {
								DisplayName:  "status",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "disabled", DisplayValue: "disabled"},
									{Value: "draft", DisplayValue: "draft"},
									{Value: "enabled", DisplayValue: "enabled"},
								},
							},
						},
					},
				},
				Errors: map[string]error{},
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
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
