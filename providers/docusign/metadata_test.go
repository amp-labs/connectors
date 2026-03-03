package docusign

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
			Name:       "Successful metadata for multiple objects",
			Input:      []string{"envelopes", "folders"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"envelopes": {
						DisplayName: "Envelopes",
						Fields: map[string]common.FieldMetadata{
							"documentsUri": {
								DisplayName:  "documentsUri",
								ValueType:    "string",
								ProviderType: "string",
							},
							"envelopeId": {
								DisplayName:  "envelopeId",
								ValueType:    "string",
								ProviderType: "string",
							},
							"envelopeDocuments": {
								DisplayName:  "envelope Documents",
								ValueType:    "other",
								ProviderType: "object",
							},
							"recipientsUri": {
								DisplayName:  "RecipientUri",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"folders": {
						DisplayName: "Folders",
						Fields: map[string]common.FieldMetadata{
							"errorDetails": {
								DisplayName:  "errorDetails",
								ValueType:    "other",
								ProviderType: "object",
							},
							"folderId": {
								DisplayName:  "folderId",
								ValueType:    "string",
								ProviderType: "string",
							},
							"folderItems": {
								DisplayName:  "folderItems",
								ValueType:    "other",
								ProviderType: "array",
							},
						},
					},
				},
			},
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
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
		WithMetadata(map[string]string{"server": "demo"}),
	)
	if err != nil {
		return nil, err
	}

	// Override the base URL to point to the test server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
