package paddle

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Unknown object requested",
			Input:      []string{"butterflies"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"butterflies": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Successful metadata for multiple objects",
			Input:      []string{"client-tokens", "discount-groups"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"client-tokens": {
						DisplayName: "Client Tokens",
						Fields: map[string]common.FieldMetadata{
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"description": {
								DisplayName:  "description",
								ValueType:    "string",
								ProviderType: "string",
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"revoked_at": {
								DisplayName:  "revoked_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "string",
								ProviderType: "string",
							},
							"token": {
								DisplayName:  "token",
								ValueType:    "string",
								ProviderType: "string",
							},
							"updated_at": {
								DisplayName:  "updated_at",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"discount-groups": {
						DisplayName: "Discount Groups",
						Fields: map[string]common.FieldMetadata{
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    "string",
								ProviderType: "string",
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"import_meta": {
								DisplayName: "import_meta",
								ValueType:   "other",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "string",
								ProviderType: "string",
							},
							"updated_at": {
								DisplayName:  "updated_at",
								ValueType:    "string",
								ProviderType: "string",
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
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
