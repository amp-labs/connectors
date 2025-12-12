package shopify

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:  "Successfully fetch metadata for Product object",
			Input: []string{"products"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/admin/api/2025-01/graphql.json"),
				Then:  mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "metadata/product.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"products": {
						DisplayName: "Products",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "ID",
							},
							"title": {
								DisplayName:  "title",
								ValueType:    "string",
								ProviderType: "String",
							},
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully fetch metadata for Order object",
			Input: []string{"orders"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/admin/api/2025-01/graphql.json"),
				Then:  mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "metadata/order.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"orders": {
						DisplayName: "Orders",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "ID",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "String",
							},
							"totalPrice": {
								DisplayName:  "totalPrice",
								ValueType:    "float",
								ProviderType: "Money",
							},
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully fetch metadata for Customer object",
			Input: []string{"customers"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/admin/api/2025-01/graphql.json"),
				Then:  mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "metadata/customer.json")),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customers": {
						DisplayName: "Customers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "ID",
							},
							"firstName": {
								DisplayName:  "firstName",
								ValueType:    "string",
								ProviderType: "String",
							},
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "String",
							},
							"numberOfOrders": {
								DisplayName:  "numberOfOrders",
								ValueType:    "int",
								ProviderType: "UnsignedInt64",
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
		tt := tt
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
			AuthenticatedClient: mockutils.NewClient(),
			Workspace:           "test-store",
		},
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	// Use SetUnitTest BaseURL to replace the base URL
	baseURL := mockutils.ReplaceURLOrigin(connector.ProviderInfo().BaseURL, serverURL)
	connector.SetUnitTestBaseURL(baseURL)

	return connector, nil
}
