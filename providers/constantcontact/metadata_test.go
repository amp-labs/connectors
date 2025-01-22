package constantcontact

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
			Input:  []string{"email_campaigns", "contact_tags"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"email_campaigns": {
						DisplayName: "Email Campaigns",
						FieldsMap: map[string]string{
							"campaign_id":    "campaign_id",
							"created_at":     "created_at",
							"current_status": "current_status",
							"name":           "name",
							"type":           "type",
							"type_code":      "type_code",
							"updated_at":     "updated_at",
						},
					},
					"contact_tags": {
						DisplayName: "Contact Tags",
						FieldsMap: map[string]string{
							"contacts_count": "contacts_count",
							"created_at":     "created_at",
							"name":           "name",
							"tag_id":         "tag_id",
							"tag_source":     "tag_source",
							"updated_at":     "updated_at",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
