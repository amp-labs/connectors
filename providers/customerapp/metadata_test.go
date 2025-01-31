package customerapp

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
			Name:   "Successfully describe multiple objects with metadata",
			Input:  []string{"reporting_webhooks", "workspaces"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"reporting_webhooks": {
						DisplayName: "Reporting Webhooks",
						FieldsMap: map[string]string{
							"disabled":        "disabled",
							"endpoint":        "endpoint",
							"events":          "events",
							"full_resolution": "full_resolution",
							"id":              "id",
							"name":            "name",
							"with_content":    "with_content",
						},
					},
					"workspaces": {
						DisplayName: "Workspaces",
						FieldsMap: map[string]string{
							"billable_messages_sent": "billable_messages_sent",
							"id":                     "id",
							"messages_sent":          "messages_sent",
							"name":                   "name",
							"object_types":           "object_types",
							"objects":                "objects",
							"people":                 "people",
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
