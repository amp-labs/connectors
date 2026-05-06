package marketing

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
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
			Name:       "Successfully describe campaigns",
			Input:      []string{"campaigns"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"campaigns": {
						DisplayName: "Campaigns",
						Fields: map[string]common.FieldMetadata{
							"hs_campaign_status": {
								DisplayName:  "Campaign Status",
								ValueType:    "singleSelect",
								ProviderType: "Enumeration",
								Values: common.FieldValues{{
									Value:        "planned",
									DisplayValue: "planned",
								}, {
									Value:        "in_progress",
									DisplayValue: "in_progress",
								}, {
									Value:        "active",
									DisplayValue: "active",
								}, {
									Value:        "paused",
									DisplayValue: "paused",
								}, {
									Value:        "completed",
									DisplayValue: "completed",
								}},
							},
							"hs_name": {
								DisplayName:  "Name",
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
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestAdapter(tt.Server.URL)
			})
		})
	}
}
