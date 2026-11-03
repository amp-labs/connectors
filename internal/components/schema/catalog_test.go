package schema

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components/mocked"
	"github.com/amp-labs/connectors/internal/objectcatalog"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestCatalog(t *testing.T) {
	t.Parallel()

	catalog := objectcatalog.MustLoad([]byte(`{
		"orders": {
			"path": "/user/orders",
			"displayName": "Orders",
			"fields": {
				"id": {"displayName": "Identifier", "valueType": "string", "providerType": "Text"}
			}
		}
	}`))

	tests := []testconn.TestCaseListObjectMetadata{
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
			Comparator: testconn.ComparatorSubsetMetadata,
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
			Comparator: testconn.ComparatorSubsetMetadata,
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return &mockedConnector{
					Connector:      mocked.Connector{BaseURL: tt.Server.URL},
					SchemaProvider: NewCatalogSchemaProvider(catalog),
				}, nil
			})
		})
	}
}
