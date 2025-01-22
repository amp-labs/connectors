package salesloft

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
			Name:         "Unknown object requested",
			Input:        []string{"butterflies"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:   "Successfully describe one object with metadata",
			Input:  []string{"bulk_jobs_results"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"bulk_jobs_results": {
						DisplayName: "Job Data for a Completed Bulk Job.",
						FieldsMap: map[string]string{
							"error":    "error",
							"id":       "id",
							"record":   "record",
							"resource": "resource",
							"status":   "status",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe multiple objects with metadata",
			Input:  []string{"account_tiers", "actions"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"account_tiers": {
						DisplayName: "Account Tiers",
						FieldsMap: map[string]string{
							"created_at": "created_at",
							"id":         "id",
							"name":       "name",
							"order":      "order",
							"updated_at": "updated_at",
							"active":     "active",
						},
					},
					"actions": {
						DisplayName: "Actions",
						FieldsMap: map[string]string{
							"action_details":      "action_details",
							"cadence":             "cadence",
							"created_at":          "created_at",
							"due":                 "due",
							"due_on":              "due_on",
							"id":                  "id",
							"multitouch_group_id": "multitouch_group_id",
							"person":              "person",
							"status":              "status",
							"step":                "step",
							"task":                "task",
							"type":                "type",
							"updated_at":          "updated_at",
							"user":                "user",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
