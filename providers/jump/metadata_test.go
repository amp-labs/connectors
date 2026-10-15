package jump

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Contacts metadata is served from the static schema",
			Input:      []string{"contacts"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeString,
								ProviderType: "ID!",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "ContactStatus",
								Values: []common.FieldValue{
									{Value: "UNCONFIRMED", DisplayValue: "UNCONFIRMED"},
									{Value: "CONFIRMED", DisplayValue: "CONFIRMED"},
								},
							},
							"updatedAt": {
								DisplayName:  "updatedAt",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "DateTime",
							},
							"contactInfo": {
								DisplayName:  "contactInfo",
								ValueType:    common.ValueTypeOther,
								ProviderType: "[ContactInfo]",
							},
						},
					},
				},
				Errors: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
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

	connector.SetUnitTestMockServerBaseUrl(serverURL)

	return connector, nil
}
