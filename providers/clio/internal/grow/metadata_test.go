package grow

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
			Name:       "Successful metadata for selected Grow objects",
			Input:      []string{"contacts", "custom_actions", "inbox_leads", "users"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "name",
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
					"custom_actions": {
						DisplayName: "Custom Actions",
						Fields: map[string]common.FieldMetadata{
							"label": {
								DisplayName:  "label",
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
					"inbox_leads": {
						DisplayName: "Inbox Leads",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "email",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"state": {
								DisplayName:  "state",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
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
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
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
		Module:              providers.ModuleClioGrow,
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "api.clio.com",
		Metadata: map[string]string{
			"region": "",
		},
	})
}
