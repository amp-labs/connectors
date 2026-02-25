package workday

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	responseCustomFieldDefs := testutils.DataFromFile(t, "custom-fields/workers/definitions.json")
	responseEmptyCustomFieldDefs := testutils.DataFromFile(t, "custom-fields/workers/empty-definitions.json")

	tests := []testroutines.Metadata{
		{
			Name:         "Unknown object requested",
			Input:        []string{"nonexistent_object"},
			Server:       mockserver.Dummy(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"nonexistent_object": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:  "Successfully describe workers object with metadata",
			Input: []string{"workers"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/ccx/api/v1/testTenant/customObjects/workers/fields"),
					Then: mockserver.Response(http.StatusOK, responseEmptyCustomFieldDefs),
				}},
			}.Server(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"workers": {
						DisplayName: "Workers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"descriptor": {
								DisplayName:  "descriptor",
								ValueType:    "string",
								ProviderType: "string",
							},
							"primaryWorkEmail": {
								DisplayName:  "primaryWorkEmail",
								ValueType:    "string",
								ProviderType: "string",
							},
							"isManager": {
								DisplayName:  "isManager",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:         "Successfully describe organizations object with metadata",
			Input:        []string{"organizations"},
			Server:       mockserver.Dummy(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"organizations": {
						DisplayName: "Organizations",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"descriptor": {
								DisplayName:  "descriptor",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:         "Successfully describe supervisory organizations object with metadata",
			Input:        []string{"supervisoryOrganizations"},
			Server:       mockserver.Dummy(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"supervisoryOrganizations": {
						DisplayName: "Supervisory Organizations",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"code": {
								DisplayName:  "code",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:         "Successfully describe currencies object with metadata",
			Input:        []string{"currencies"},
			Server:       mockserver.Dummy(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"currencies": {
						DisplayName: "Currencies",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"code": {
								DisplayName:  "code",
								ValueType:    "string",
								ProviderType: "string",
							},
							"precision": {
								DisplayName:  "precision",
								ValueType:    "int",
								ProviderType: "integer",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"workers", "organizations", "auditLogs", "jobChangeReasons"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/ccx/api/v1/testTenant/customObjects/workers/fields"),
					Then: mockserver.Response(http.StatusOK, responseEmptyCustomFieldDefs),
				}},
			}.Server(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"workers": {
						DisplayName: "Workers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"organizations": {
						DisplayName: "Organizations",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"auditLogs": {
						DisplayName: "Audit Logs",
						Fields: map[string]common.FieldMetadata{
							"activityAction": {
								DisplayName:  "activityAction",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"jobChangeReasons": {
						DisplayName: "Job Change Reasons",
						Fields: map[string]common.FieldMetadata{
							"isForEmployee": {
								DisplayName:  "isForEmployee",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "Workers metadata includes custom fields",
			Input: []string{"workers"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/ccx/api/v1/testTenant/customObjects/workers/fields"),
					Then: mockserver.Response(http.StatusOK, responseCustomFieldDefs),
				}},
			}.Server(),
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"workers": {
						DisplayName: "Workers",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"descriptor": {
								DisplayName:  "descriptor",
								ValueType:    "string",
								ProviderType: "string",
							},
							"custom_field_department_code": {
								DisplayName:  "Department Code",
								ValueType:    "string",
								ProviderType: "Text",
								IsCustom:     goutils.Pointer(true),
							},
							"custom_field_years_experience": {
								DisplayName:  "Years Experience",
								ValueType:    "float",
								ProviderType: "Numeric",
								IsCustom:     goutils.Pointer(true),
							},
						},
					},
				},
				Errors: map[string]error{},
			},
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "test",
		Metadata: map[string]string{
			"tenantName": "testTenant",
		},
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
