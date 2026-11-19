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
								DisplayName:  "envelopeDocuments",
								ValueType:    "other",
								ProviderType: "array",
							},
							"recipientsUri": {
								DisplayName:  "recipientsUri",
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
		{
			Name:         "Must query at least one object",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object",
			Input:      []string{"doggo"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"doggo": common.ErrObjectNotSupported,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				connMetadata := map[string]string{
					"server":     "devTest",
					"account_id": "devTest-123",
				}
				return constructTestConnector(tt.Server.URL, connMetadata)
			})
		})
	}
}

func constructTestConnector(serverURL string, metadata map[string]string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(mockutils.NewClient()),
		WithMetadata(metadata),
	)
	if err != nil {
		return nil, err
	}

	// Override the base URL to point to the test server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
