package keap

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseContactsModel := testutils.DataFromFile(t, "custom-fields/contacts-v2.json")

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
			Name: "Successfully describe multiple objects with metadata",
			Input: []string{
				"campaigns", "contacts",
				"automationCategory", "tags",
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/crm/rest/v2/contacts/model"),
				Then:  mockserver.Response(http.StatusOK, responseContactsModel),
			}.Server(),
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
					"contacts": {
						DisplayName: "Contacts",
						FieldsMap: map[string]string{
							"addresses":        "addresses",
							"anniversary_date": "anniversary_date",
							"birth_date":       "birth_date",
							"company":          "company",
							"contact_type":     "contact_type",
							"create_time":      "create_time",
							"custom_fields":    "custom_fields",
							"email_addresses":  "email_addresses",
							"family_name":      "family_name",
							"fax_numbers":      "fax_numbers",
							"given_name":       "given_name",
							"id":               "id",
							"job_title":        "job_title",
							"leadsource_id":    "leadsource_id",
							"links":            "links",
							"middle_name":      "middle_name",
							"origin":           "origin",
							"owner_id":         "owner_id",
							"phone_numbers":    "phone_numbers",
							"preferred_locale": "preferred_locale",
							"preferred_name":   "preferred_name",
							"prefix":           "prefix",
							"referral_code":    "referral_code",
							"score_value":      "score_value",
							"social_accounts":  "social_accounts",
							"source_type":      "source_type",
							"spouse_name":      "spouse_name",
							"suffix":           "suffix",
							"tag_ids":          "tag_ids",
							"time_zone":        "time_zone",
							"update_time":      "update_time",
							"utm_parameters":   "utm_parameters",
							"website":          "website",
							// Custom fields.
							"jobtitle":       "jobtitle",
							"jobdescription": "jobdescription",
							"experience":     "experience",
							"age":            "age",
						},
					},
					"automationCategory": {
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
							"category":    "category",
							"description": "description",
						},
					},
				},

				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Metadata error due to failed custom data requests",
			Input: []string{"contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/crm/rest/v2/contacts/model"),
				Then:  mockserver.Response(http.StatusInternalServerError),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"contacts": mockutils.ExpectedSubsetErrors{
						common.ErrResolvingCustomFields,
						common.ErrServer,
					},
				},
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
