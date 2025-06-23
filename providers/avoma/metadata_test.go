package avoma

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
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object requested",
			Input:      []string{"templates"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"templates": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"notes", "template"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"notes": {
						DisplayName: "Notes",
						Fields: map[string]common.FieldMetadata{
							"created": {
								DisplayName:  "created",
								ValueType:    "string",
								ProviderType: "string",
							},
							"data": {
								DisplayName: "data",
								ValueType:   "other",
							},
							"modified": {
								DisplayName:  "modified",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"template": {
						DisplayName: "Template",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
							"is_default": {
								DisplayName:  "is_default",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"meeting_types": {
								DisplayName:  "meeting_types",
								ValueType:    "other",
								ProviderType: "array",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"privacy": {
								DisplayName:  "privacy",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: common.FieldValues{
									{Value: "private", DisplayValue: "private"},
									{Value: "organization", DisplayValue: "organization"},
								},
							},
							"text_slate": {
								DisplayName:  "text_slate",
								ValueType:    "string",
								ProviderType: "string",
							},
							"uuid": {
								DisplayName:  "uuid",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
			},
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

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
