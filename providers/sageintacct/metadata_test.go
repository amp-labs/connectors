package sageintacct

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
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
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"account", "contact"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"account": {
						DisplayName: "Account",
						Fields: map[string]common.FieldMetadata{
							"href": {
								DisplayName:  "href",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       nil,
							},
							"key": {
								DisplayName:  "key",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
					"contact": {
						DisplayName: "Contact",
						Fields: map[string]common.FieldMetadata{
							"href": {
								DisplayName:  "href",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       nil,
							},
							"key": {
								DisplayName:  "key",
								ValueType:    "string",
								ProviderType: "string",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: nil,
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
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
