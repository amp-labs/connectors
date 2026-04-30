package manage

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successful metadata for selected Manage objects",
			Input:      []string{"activities", "contacts", "matters", "users"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"activities": {
						DisplayName: "Activities",
						Fields: map[string]common.FieldMetadata{
							"created_at": {
								DisplayName:  "created_at",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
							},
						},
					},
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"is_client": {
								DisplayName:  "is_client",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
							},
						},
					},
					"matters": {
						DisplayName: "Matters",
						Fields: map[string]common.FieldMetadata{
							"description": {
								DisplayName:  "description",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
							},
						},
					},
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "email",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"enabled": {
								DisplayName:  "enabled",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
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
				return constructTestAdapter()
			})
		})
	}
}

func constructTestAdapter() (*Adapter, error) {
	return NewAdapter(common.ConnectorParams{
		Module:              providers.ModuleClioManage,
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "app.clio.com",
		Metadata: map[string]string{
			"region": "",
		},
	})
}
