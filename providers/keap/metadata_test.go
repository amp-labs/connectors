package keap

import (
	"errors"
	"log/slog"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadataV1(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseContactsModel := testutils.DataFromFile(t, "custom-fields-contacts.json")

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
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"campaigns", "products", "contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/crm/rest/v1/contacts/model"),
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
					"products": {
						DisplayName: "Products",
						FieldsMap: map[string]string{
							"id":            "id",
							"sku":           "sku",
							"status":        "status",
							"product_price": "product_price",
						},
					},
					"contacts": {
						DisplayName: "Contacts",
						FieldsMap: map[string]string{
							"ScoreValue":       "ScoreValue",
							"addresses":        "addresses",
							"anniversary":      "anniversary",
							"birthday":         "birthday",
							"company":          "company",
							"company_name":     "company_name",
							"contact_type":     "contact_type",
							"custom_fields":    "custom_fields",
							"date_created":     "date_created",
							"email_addresses":  "email_addresses",
							"email_opted_in":   "email_opted_in",
							"email_status":     "email_status",
							"family_name":      "family_name",
							"fax_numbers":      "fax_numbers",
							"given_name":       "given_name",
							"id":               "id",
							"job_title":        "job_title",
							"last_updated":     "last_updated",
							"lead_source_id":   "lead_source_id",
							"middle_name":      "middle_name",
							"owner_id":         "owner_id",
							"phone_numbers":    "phone_numbers",
							"preferred_locale": "preferred_locale",
							"preferred_name":   "preferred_name",
							"prefix":           "prefix",
							"social_accounts":  "social_accounts",
							"source_type":      "source_type",
							"spouse_name":      "spouse_name",
							"suffix":           "suffix",
							"time_zone":        "time_zone",
							"website":          "website",
							// Custom fields.
							"jobtitle":       "title",
							"jobdescription": "job_description",
							"experience":     "experience",
							"age":            "age",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Partial metadata due to failed custom data requests",
			Input: []string{"contacts"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/crm/rest/v1/contacts/model"),
				Then:  mockserver.Response(http.StatusInternalServerError),
			}.Server(),
			Comparator: metadataExpectAbsentFields,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						FieldsMap: map[string]string{
							// These custom fields MUST be absent due to not responding server.
							"jobtitle":       "title",
							"jobdescription": "job_description",
							"experience":     "experience",
							"age":            "age",
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

func metadataExpectAbsentFields(serverURL string, actual, expected *common.ListObjectMetadataResult) bool {
	const contacts = "contacts"
	if !errors.Is(actual.Errors[contacts], ErrResolvingCustomFields) {
		slog.Info("missing metadata error", "errors", ErrResolvingCustomFields)

		return false
	}

	for fieldName := range expected.Result[contacts].FieldsMap {
		_, present := actual.Result[contacts].FieldsMap[fieldName]
		if present {
			slog.Info("custom field should NOT be present", "field", fieldName)

			return false
		}
	}

	return true
}
