package clarify

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen
	t.Parallel()

	responseSchemas := testutils.DataFromFile(t, "schemas.json")
	responseNotFound := testutils.DataFromFile(t, "error-workspace-not-found.json")

	schemasServer := func() *mockserver.Conditional {
		return &mockserver.Conditional{
			Setup: mockserver.ContentJSON(),
			If: mockcond.And{
				mockcond.MethodGET(),
				mockcond.Path("/v1/workspaces/test-ws/schemas"),
			},
			Then: mockserver.Response(http.StatusOK, responseSchemas),
		}
	}

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Entities the connector does not read are reported per object",
			Input:      []string{"agent", "c_unknown"},
			Server:     schemasServer().Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"agent":     mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
					"c_unknown": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
			},
		},
		{
			Name:       "Fields, enums and to-one relationships come from the workspace schemas",
			Input:      []string{"company", "deal", "meeting", "users", "c_sample_contract"},
			Server:     schemasServer().Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"company": {
						DisplayName: "Companies",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "ID",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								ReadOnly:     new(true),
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"domains": {
								DisplayName:  "Domains",
								ValueType:    common.ValueTypeOther,
								ProviderType: "collectionOfHostnames",
							},
							"_updated_at": {
								DisplayName:  "Updated at",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "object:date-time",
								ReadOnly:     new(true),
							},
							"employee_range": {
								DisplayName:  "Employee range",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "1-10", DisplayValue: "1-10"},
									{Value: "11-50", DisplayValue: "11-50"},
									{Value: "51-250", DisplayValue: "51-250"},
									{Value: "251-1K", DisplayValue: "251-1K"},
									{Value: "1K-5K", DisplayValue: "1K-5K"},
									{Value: "5K-10K", DisplayValue: "5K-10K"},
									{Value: "10K-50K", DisplayValue: "10K-50K"},
									{Value: "50K-100K", DisplayValue: "50K-100K"},
									{Value: "100K+", DisplayValue: "100K+"},
								},
							},
						},
					},
					"deal": {
						DisplayName: "Deals",
						Fields: map[string]common.FieldMetadata{
							"amount": {
								DisplayName:  "Amount",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "number",
							},
							"close_date": {
								DisplayName:  "Close date",
								ValueType:    common.ValueTypeDate,
								ProviderType: "string:date",
							},
							"company_id": {
								DisplayName:  "Company",
								ValueType:    common.ValueTypeReference,
								ProviderType: "string:uuid",
								ReferenceTo:  []string{"company"},
							},
							"owner_id": {
								DisplayName:  "Owner",
								ValueType:    common.ValueTypeReference,
								ProviderType: "string",
								ReferenceTo:  []string{"users"},
							},
						},
					},
					"meeting": {
						DisplayName: "Meetings",
						Fields: map[string]common.FieldMetadata{
							// Links to an entity the connector doesn't read keep the id's own type.
							"recurrence_rule_id": {
								DisplayName:  "Recurrence rule",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
								ReadOnly:     new(true),
							},
						},
					},
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "Email",
								ValueType:    common.ValueTypeString,
								ProviderType: "string:email",
								ReadOnly:     new(true),
							},
							"last_active_at": {
								DisplayName:  "Last active at",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "object:date-time",
								ReadOnly:     new(true),
							},
						},
					},
					"c_sample_contract": {
						DisplayName: "Sample Contracts",
						Fields: map[string]common.FieldMetadata{
							"contract_value": {
								DisplayName:  "Contract value",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "number",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "To-many relationships are not listed because Read does not return them",
			Input: []string{"company"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseSchemas),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadataWithMissingFields(map[string][]string{
				"company": {"deals", "people", "tasks", "attachments"},
			}),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"company": {DisplayName: "Companies", Fields: map[string]common.FieldMetadata{}},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "Users list only what GET /users returns",
			Input: []string{"users"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseSchemas),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadataWithMissingFields(map[string][]string{
				"users": {"auth0_id", "workos_id", "_deleted_at", "_created_by", "owned_deals"},
			}),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"is_service_account": {
								DisplayName:  "Is service account",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
								ReadOnly:     new(true),
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "Error response fails the call",
			Input: []string{"company"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, responseNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrNotFound,
				testutils.StringError("Workspace not found"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
