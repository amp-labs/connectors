package lightfield

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

	responseOpportunityDefinitions := testutils.DataFromFile(t, "opportunities-definitions.json")
	responseForbidden := testutils.DataFromFile(t, "error-forbidden.json")

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object is reported per object",
			Input:      []string{"notes"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{},
				Errors: map[string]error{
					"notes": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
			},
		},
		{
			Name:       "Members come from the static catalog",
			Input:      []string{"members"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"members": {
						DisplayName: "Members",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    common.ValueTypeString,
								ProviderType: "string",
							},
							"updatedAt": {
								DisplayName: "Updated At",
								ValueType:   common.ValueTypeDateTime,
							},
							"$email": {
								DisplayName:  "Email",
								ValueType:    common.ValueTypeString,
								ProviderType: "EMAIL",
							},
							"$name": {
								DisplayName:  "Name",
								ValueType:    common.ValueTypeOther,
								ProviderType: "FULL_NAME",
							},
							"$user": {
								DisplayName:  "User",
								ValueType:    common.ValueTypeReference,
								ProviderType: "HAS_ONE",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "Opportunities are described by the definitions endpoint",
			Input: []string{"opportunities"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v1/opportunities/definitions"),
					mockcond.Header(http.Header{"Lightfield-Version": []string{"2026-03-01"}}),
				},
				Then: mockserver.Response(http.StatusOK, responseOpportunityDefinitions),
			}.Server(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"opportunities": {
						DisplayName: "Opportunities",
						Fields: map[string]common.FieldMetadata{
							"createdAt": {
								DisplayName:  "Created At",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "string",
							},
							"$name": {
								DisplayName:  "Name",
								ValueType:    common.ValueTypeString,
								ProviderType: "TEXT",
								ReadOnly:     new(false),
								IsCustom:     new(false),
							},
							"$stage": {
								DisplayName:  "Stage",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "SINGLE_SELECT",
								IsCustom:     new(false),
								Values: []common.FieldValue{
									{Value: "opt_prospecting", DisplayValue: "Prospecting"},
									{Value: "opt_closed_won", DisplayValue: "Closed Won"},
								},
							},
							"$amount": {
								DisplayName:  "Amount",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "CURRENCY",
								IsCustom:     new(false),
							},
							"$summary": {
								DisplayName:  "Summary",
								ValueType:    common.ValueTypeString,
								ProviderType: "READONLY_MARKDOWN",
								ReadOnly:     new(true),
								IsCustom:     new(false),
							},
							"priority": {
								DisplayName:  "Priority",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "NUMBER",
								IsCustom:     new(true),
							},
							"$account": {
								DisplayName:  "Account",
								ValueType:    common.ValueTypeReference,
								ProviderType: "HAS_ONE",
								IsCustom:     new(false),
								ReferenceTo:  []string{"accounts"},
							},
							"$owner": {
								DisplayName:  "Owner",
								ValueType:    common.ValueTypeReference,
								ProviderType: "HAS_ONE",
								IsCustom:     new(false),
							},
							"stakeholders": {
								DisplayName:  "Stakeholders",
								ValueType:    common.ValueTypeOther,
								ProviderType: "HAS_MANY",
								IsCustom:     new(true),
								ReferenceTo:  []string{"contacts"},
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:  "Failed definitions request is reported only for that object",
			Input: []string{"accounts", "members"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/v1/accounts/definitions"),
				Then:  mockserver.Response(http.StatusForbidden, responseForbidden),
			}.Server(),
			// accounts must not come back with its catalog-only fields.
			Comparator: func(serverURL string, actual, expected *common.ListObjectMetadataResult) *testutils.CompareResult {
				result := testconn.ComparatorSubsetMetadata(serverURL, actual, expected)
				if _, ok := actual.Result["accounts"]; ok {
					result.AddDiff("Result[accounts] exists, but the object failed")
				}

				return result
			},
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"members": {
						DisplayName: "Members",
						Fields: map[string]common.FieldMetadata{
							"$role": {
								DisplayName:  "Role",
								ValueType:    common.ValueTypeString,
								ProviderType: "TEXT",
							},
						},
					},
				},
				Errors: map[string]error{
					"accounts": mockutils.ExpectedSubsetErrors{
						common.ErrForbidden,
						testutils.StringError("API key is missing the required scope."),
					},
				},
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

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
