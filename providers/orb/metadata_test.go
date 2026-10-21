package orb

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object is reported per object",
			Input:      []string{"events"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"events": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
			},
		},
		{
			Name:       "Objects come from the catalog generated from the OpenAPI spec",
			Input:      []string{"customers", "invoices", "metrics", "subscriptions"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customers": {
						DisplayName: "Customers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"external_customer_id": {
								DisplayName:  "External Customer Id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"timezone": {
								DisplayName:  "Timezone",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"created_at": {
								DisplayName:  "Created At",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "string",
							},
						},
					},
					"invoices": {
						DisplayName: "Invoices",
						Fields: map[string]common.FieldMetadata{
							"status": {
								DisplayName:  "Status",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "issued", DisplayValue: "issued"},
									{Value: "paid", DisplayValue: "paid"},
									{Value: "synced", DisplayValue: "synced"},
									{Value: "void", DisplayValue: "void"},
									{Value: "draft", DisplayValue: "draft"},
								},
							},
							"paid_at": {
								DisplayName:  "Paid At",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "string",
							},
						},
					},
					"metrics": {
						DisplayName: "Metrics",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "Name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
						},
					},
					"subscriptions": {
						DisplayName: "Subscriptions",
						Fields: map[string]common.FieldMetadata{
							"customer": {
								DisplayName:  "Customer",
								ValueType:    common.ValueTypeOther,
								ProviderType: "object",
							},
							"end_date": {
								DisplayName:  "End Date",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
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
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
