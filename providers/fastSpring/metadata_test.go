package fastspring

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
			Name:  "Successful metadata for commerce objects",
			Input: []string{"accounts", "orders", "products", "subscriptions", "events-processed", "events-unprocessed"},
			Server: mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"accounts": {
						DisplayName: "Accounts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Account Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"account": {
								DisplayName:  "Account",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"orders": {
						DisplayName: "Orders",
						Fields: map[string]common.FieldMetadata{
							"order": {
								DisplayName:  "Order Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"id": {
								DisplayName:  "Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"products": {
						DisplayName: "Products",
						Fields: map[string]common.FieldMetadata{
							"path": {
								DisplayName:  "Product Path",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"subscriptions": {
						DisplayName: "Subscriptions",
						Fields: map[string]common.FieldMetadata{
							"subscription": {
								DisplayName:  "Subscription Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"id": {
								DisplayName:  "Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"events-processed": {
						DisplayName: "Processed Events",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Event Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"type": {
								DisplayName:  "Type",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"events-unprocessed": {
						DisplayName: "Unprocessed Events",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Event Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"type": {
								DisplayName:  "Type",
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
		{
			Name:         "Empty objects returns missing objects error",
			Input:        nil,
			Server:       mockserver.Dummy(),
			Expected:     nil,
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unsupported object returns object not supported error",
			Input:      []string{"accounts", "unknown_object"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"accounts": {
						DisplayName: "Accounts",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Account Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{
					"unknown_object": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
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
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: mockutils.NewClient(),
			Workspace:           "test-workspace",
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
