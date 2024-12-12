package keap

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadataV1(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unknown object requested",
			Input:        []string{"butterflies"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{staticschema.ErrObjectNotFound},
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"campaigns", "products"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"campaigns": {
						DisplayName: "Campaigns",
						FieldsMap: map[string]string{
							"id":     "id",
							"name":   "name",
							"locked": "locked",
							"goals":  "goals",
						},
					},
					"products": {
						DisplayName: "Products",
						FieldsMap: map[string]string{
							"id":            "id",
							"sku":           "sku",
							"status":        "status",
							"product_price": "product_price",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL, ModuleV1)
			})
		})
	}
}

func TestListObjectMetadataV2(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"automation_categories", "tags"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"automation_categories": {
						DisplayName: "Automation Categories",
						FieldsMap: map[string]string{
							"id":               "id",
							"name":             "name",
							"automation_count": "automation_count",
						},
					},
					"tags": {
						DisplayName: "Tags",
						FieldsMap: map[string]string{
							"id":          "id",
							"name":        "name",
							"description": "description",
							"category":    "category",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL, ModuleV2)
			})
		})
	}
}
