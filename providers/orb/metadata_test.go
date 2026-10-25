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
			Input:      []string{"plans"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"plans": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
			},
		},
		{
			Name:       "Objects come from the catalog generated from the OpenAPI spec",
			Input:      []string{"customers", "subscriptions", "metrics"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"customers": {
						DisplayName: "Customers",
						Fields: map[string]common.FieldMetadata{
							"external_customer_id": {
								DisplayName:  "External Customer Id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"created_at": {
								DisplayName:  "Created At",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "string",
							},
							"timezone": {
								DisplayName:  "Timezone",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
						},
					},
					"subscriptions": {
						DisplayName: "Subscriptions",
						Fields: map[string]common.FieldMetadata{
							"end_date": {
								DisplayName:  "End Date",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "Status",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "active", DisplayValue: "active"},
									{Value: "ended", DisplayValue: "ended"},
									{Value: "upcoming", DisplayValue: "upcoming"},
								},
							},
							"customer": {
								DisplayName:  "Customer",
								ValueType:    common.ValueTypeOther,
								ProviderType: "object",
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
