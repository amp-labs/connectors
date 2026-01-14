package salesloft

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseCustomFieldsFirstPage := testutils.DataFromFile(t, "read/custom-fields/first-page.json")
	responseCustomFieldsSecondPage := testutils.DataFromFile(t, "read/custom-fields/last-page.json")

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
			Name:       "Successfully describe one object with metadata",
			Input:      []string{"activities/calls"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"activities/calls": {
						DisplayName: "Calls",
						FieldsMap: map[string]string{
							"disposition": "disposition",
							"duration":    "duration",
							"id":          "id",
							"note":        "note",
							"positive":    "positive",
							"recordings":  "recordings",
							"sentiment":   "sentiment",
						},
					},
				},
				Errors: make(map[string]error),
			},
			ExpectedErrs: nil,
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"account_tiers", "actions"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
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
		{
			Name:  "Successfully describe object people with custom fields",
			Input: []string{"people"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				// Custom fields are intentionally split across two pages to simulate
				// a provider account with more than 100 custom fields.
				//
				// This verifies that the connector:
				//   - Uses the maximum page size (per_page=100)
				//   - Follows pagination correctly
				//   - Aggregates custom fields from all pages before describing the object
				Cases: mockserver.Cases{{
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/v2/custom_fields"),
						mockcond.QueryParam("per_page", "100"),
						mockcond.QueryParamsMissing("page"),
					},
					Then: mockserver.Response(http.StatusOK, responseCustomFieldsFirstPage),
				}, {
					If: mockcond.And{
						mockcond.MethodGET(),
						mockcond.Path("/v2/custom_fields"),
						mockcond.QueryParam("per_page", "100"),
						mockcond.QueryParam("page", "2"),
					},
					Then: mockserver.Response(http.StatusOK, responseCustomFieldsSecondPage),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"people": {
						DisplayName: "People",
						Fields: map[string]common.FieldMetadata{
							"city": {
								DisplayName:  "city",
								ValueType:    "string",
								ProviderType: "string",
							},
							// All fields below originate from custom field definitions.
							"hobby": {
								DisplayName:  "hobby",
								ValueType:    "string",
								ProviderType: "text",
								IsCustom:     goutils.Pointer(true),
							},
							"test-field": {
								DisplayName:  "test-field",
								ValueType:    "string",
								ProviderType: "text",
								IsCustom:     goutils.Pointer(true),
							},
							"mails": {
								DisplayName:  "mails",
								ValueType:    "string",
								ProviderType: "text",
								IsCustom:     goutils.Pointer(true),
							},
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
