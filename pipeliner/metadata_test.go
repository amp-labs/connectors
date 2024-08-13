package pipeliner

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
			Input:  []string{"Notes"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"Notes": {
						DisplayName: "Notes",
						FieldsMap: map[string]string{
							"account":             "account",
							"account_id":          "account_id",
							"contact":             "contact",
							"contact_id":          "contact_id",
							"created":             "created",
							"custom_entity":       "custom_entity",
							"custom_entity_id":    "custom_entity_id",
							"id":                  "id",
							"is_delete_protected": "is_delete_protected",
							"is_deleted":          "is_deleted",
							"lead":                "lead",
							"lead_oppty_id":       "lead_oppty_id",
							"modified":            "modified",
							"note":                "note",
							"oppty":               "oppty",
							"owner":               "owner",
							"owner_id":            "owner_id",
							"project":             "project",
							"project_id":          "project_id",
							"quote":               "quote",
							"quote_id":            "quote_id",
							"revision":            "revision",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe multiple objects with metadata",
			Input:  []string{"Phones", "Tags"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"Phones": {
						DisplayName: "Phones",
						FieldsMap: map[string]string{
							"call_forwarding_phone":    "call_forwarding_phone",
							"created":                  "created",
							"id":                       "id",
							"is_delete_protected":      "is_delete_protected",
							"is_deleted":               "is_deleted",
							"message_forwarding_email": "message_forwarding_email",
							"modified":                 "modified",
							"name":                     "name",
							"owner":                    "owner",
							"owner_id":                 "owner_id",
							"phone_number":             "phone_number",
							"revision":                 "revision",
							"twilio_caller_sid":        "twilio_caller_sid",
							"type":                     "type",
						},
					},
					"Tags": {
						DisplayName: "Tags",
						FieldsMap: map[string]string{
							"color":               "color",
							"created":             "created",
							"creator":             "creator",
							"creator_id":          "creator_id",
							"id":                  "id",
							"is_delete_protected": "is_delete_protected",
							"is_deleted":          "is_deleted",
							"modified":            "modified",
							"name":                "name",
							"revision":            "revision",
							"supported_entities":  "supported_entities",
							"use_lang":            "use_lang",
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
