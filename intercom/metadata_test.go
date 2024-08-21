package intercom

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/tools/scrapper"
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
			Name:         "Unknown object requested",
			Input:        []string{"butterflies"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{scrapper.ErrObjectNotFound},
		},
		{
			Name:   "Successfully describe one object with metadata",
			Input:  []string{"help_centers"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"help_centers": {
						DisplayName: "Help Centers",
						FieldsMap: map[string]string{
							"created_at":        "created_at",
							"display_name":      "display_name",
							"id":                "id",
							"identifier":        "identifier",
							"updated_at":        "updated_at",
							"website_turned_on": "website_turned_on",
							"workspace_id":      "workspace_id",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe multiple objects with metadata",
			Input:  []string{"segments", "teams"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"segments": {
						DisplayName: "Segments",
						FieldsMap: map[string]string{
							"count":       "count",
							"created_at":  "created_at",
							"id":          "id",
							"name":        "name",
							"person_type": "person_type",
							"type":        "type",
							"updated_at":  "updated_at",
						},
					},
					"teams": {
						DisplayName: "Teams",
						FieldsMap: map[string]string{
							"admin_ids":            "admin_ids",
							"admin_priority_level": "admin_priority_level",
							"id":                   "id",
							"name":                 "name",
							"type":                 "type",
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
