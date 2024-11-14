package klaviyo

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/test/utils/mockutils"
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
			Name:   "Successfully describe multiple objects with metadata",
			Input:  []string{"campaigns", "lists"},
			Server: mockserver.Dummy(),
			Comparator: func(baseURL string, actual, expected *common.ListObjectMetadataResult) bool {
				return mockutils.MetadataResultComparator.SubsetFields(actual, expected)
			},
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"campaigns": {
						DisplayName: "Campaigns",
						FieldsMap: map[string]string{
							"archived":         "archived",
							"audiences":        "audiences",
							"created_at":       "created_at",
							"id":               "id",
							"links":            "links",
							"name":             "name",
							"relationships":    "relationships",
							"scheduled_at":     "scheduled_at",
							"send_options":     "send_options",
							"send_strategy":    "send_strategy",
							"send_time":        "send_time",
							"status":           "status",
							"tracking_options": "tracking_options",
							"type":             "type",
							"updated_at":       "updated_at",
						},
					},
					"lists": {
						DisplayName: "Lists",
						FieldsMap: map[string]string{
							"created":        "created",
							"id":             "id",
							"links":          "links",
							"name":           "name",
							"opt_in_process": "opt_in_process",
							"relationships":  "relationships",
							"type":           "type",
							"updated":        "updated",
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
		tt := tt // rebind, omit loop side effects for parallel goroutine
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
