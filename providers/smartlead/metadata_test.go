package smartlead

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/test/utils/mockutils"
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
			ExpectedErrs: []error{staticschema.ErrObjectNotFound},
		},
		{
			Name:   "Successfully describe multiple objects with metadata",
			Input:  []string{"campaigns", "leads"},
			Server: mockserver.Dummy(),
			Comparator: func(baseURL string, actual, expected *common.ListObjectMetadataResult) bool {
				return mockutils.MetadataResultComparator.SubsetFields(actual, expected)
			},
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"campaigns": {
						DisplayName: "Campaigns",
						FieldsMap: map[string]string{
							"id":         "ID",
							"user_id":    "User ID",
							"created_at": "Created at",
							"updated_at": "Updated at",
							"status":     "Status",
							"name":       "Name",
						},
					},
					"leads": {
						DisplayName: "Leads",
						FieldsMap: map[string]string{
							"first_name":   "First name",
							"last_name":    "Last name",
							"email":        "Email",
							"created_at":   "Created at",
							"phone_number": "Phone number",
							"company_name": "Company name",
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
