package schema

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/components/mocked"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestOpenAPI(t *testing.T) { //nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object requested",
			Input:      []string{"someUnknownObject"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"someUnknownObject": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Successfully describe one object with metadata",
			Input:      []string{"orders"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"orders": {
						DisplayName: "Orders",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Identifier",
								ValueType:    common.ValueTypeString,
								ProviderType: "Text",
							},
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL), nil
			})
		})
	}
}

// Connector used to test SchemaProvider.
type mockedConnector struct {
	mocked.Connector
	components.SchemaProvider
}

func constructTestConnector(serverURL string) *mockedConnector {
	connector := mocked.Connector{
		BaseURL: serverURL,
	}

	metadata := staticschema.NewMetadata[staticschema.FieldMetadataMapV2]()
	metadata.Add(common.ModuleRoot, "orders", "Orders", "/user/orders", "result",
		map[string]staticschema.FieldMetadata{
			"id": {
				DisplayName:  "Identifier",
				ValueType:    common.ValueTypeString,
				ProviderType: "Text",
			},
		}, nil, nil)

	return &mockedConnector{
		Connector: connector,
		SchemaProvider: NewOpenAPISchemaProvider(
			common.ModuleRoot,
			metadata,
		),
	}
}
