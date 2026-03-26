package greenhouse

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

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	responseCustomFieldDefs := testutils.DataFromFile(t, "read/custom-fields/definitions.json")
	responseCustomFieldOpts := testutils.DataFromFile(t, "read/custom-fields/options.json")

	tests := []testroutines.Metadata{
		{
			Name:       "Unknown object requested",
			Input:      []string{"pipelines"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"pipelines": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Successfully describe users object with metadata",
			Input:      []string{"users"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"first_name": {
								DisplayName:  "first_name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"site_admin": {
								DisplayName:  "site_admin",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"emails": {
								DisplayName:  "emails",
								ValueType:    "other",
								ProviderType: "array",
							},
						},
					},
				},
			},
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"applications", "scorecards"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"applications": {
						DisplayName: "Applications",
						Fields: map[string]common.FieldMetadata{
							"candidate_id": {
								DisplayName:  "candidate_id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: common.FieldValues{
									{Value: "rejected", DisplayValue: "rejected"},
									{Value: "hired", DisplayValue: "hired"},
									{Value: "converted", DisplayValue: "converted"},
									{Value: "in_process", DisplayValue: "in_process"},
								},
							},
							"prospect": {
								DisplayName:  "prospect",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
						},
					},
					"scorecards": {
						DisplayName: "Scorecards",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "int",
								ProviderType: "integer",
							},
						},
					},
				},
			},
		},
		{
			Name:  "Metadata includes custom fields from API",
			Input: []string{"applications"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: mockserver.Cases{
					{
						If:   mockcond.Path("/v3/custom_fields"),
						Then: mockserver.Response(http.StatusOK, responseCustomFieldDefs),
					},
					{
						If:   mockcond.Path("/v3/custom_field_options"),
						Then: mockserver.Response(http.StatusOK, responseCustomFieldOpts),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"applications": {
						DisplayName: "Applications",
						Fields: map[string]common.FieldMetadata{
							// Base schema field
							"candidate_id": {
								DisplayName:  "candidate_id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							// Custom field: number type
							"interview_score": {
								DisplayName:  "Interview Score",
								ValueType:    "int",
								ProviderType: "number",
								IsCustom:     goutils.Pointer(true),
							},
							// Custom field: single_select with options
							"referral_source": {
								DisplayName:  "Referral Source",
								ValueType:    "singleSelect",
								ProviderType: "single_select",
								IsCustom:     goutils.Pointer(true),
								Values: common.FieldValues{
									{Value: "Employee Referral", DisplayValue: "Employee Referral"},
									{Value: "LinkedIn", DisplayValue: "LinkedIn"},
									{Value: "Job Board", DisplayValue: "Job Board"},
								},
							},
						},
					},
				},
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
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
