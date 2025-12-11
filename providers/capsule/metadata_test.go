package capsule

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

	responsePartiesCustomFields := testutils.DataFromFile(t, "read/parties/custom-fields.json")
	responseProjectsCustomFields := testutils.DataFromFile(t, "read/projects/custom-fields.json")

	tests := []testroutines.Metadata{
		{
			Name:  "Successful metadata for multiple objects",
			Input: []string{"activitytypes", "parties"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.Path("/api/v2/parties/fields/definitions"),
				Then:  mockserver.Response(http.StatusOK, responsePartiesCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"activitytypes": {
						DisplayName: "Activity Types",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Id",
								ValueType:    "int",
								ProviderType: "Long",
								ReadOnly:     goutils.Pointer(true),
							},
							"updateLastContacted": {
								DisplayName:  "Update Last Contacted",
								ValueType:    "boolean",
								ProviderType: "Boolean",
							},
						},
					},
					"parties": {
						DisplayName: "Parties",
						Fields: map[string]common.FieldMetadata{
							"lastContactedAt": {
								DisplayName:  "Last Contacted At",
								ValueType:    "date",
								ProviderType: "Date",
								ReadOnly:     goutils.Pointer(true),
							},
							"type": {
								DisplayName:  "Type",
								ValueType:    "singleSelect",
								ProviderType: "String",
								Values: common.FieldValues{{
									Value:        "person",
									DisplayValue: "person",
								}, {
									Value:        "organisation",
									DisplayValue: "organisation",
								}},
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Projects metadata include custom fields",
			Input: []string{"projects"},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				// API still uses the old name "kases" instead of "projects"
				If:   mockcond.Path("/api/v2/kases/fields/definitions"),
				Then: mockserver.Response(http.StatusOK, responseProjectsCustomFields),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"projects": {
						DisplayName: "Projects",
						Fields: map[string]common.FieldMetadata{
							"createdAt": {
								DisplayName:  "Created At",
								ValueType:    "date",
								ProviderType: "Date",
								ReadOnly:     goutils.Pointer(true),
							},
							"status": {
								DisplayName:  "Status",
								ValueType:    "singleSelect",
								ProviderType: "String",
								Values: common.FieldValues{{
									Value:        "OPEN",
									DisplayValue: "OPEN",
								}, {
									Value:        "CLOSED",
									DisplayValue: "CLOSED",
								}},
							},
							// Custom field which comes from a dedicated API call.
							"Interests": {
								DisplayName:  "Interests",
								ValueType:    "singleSelect",
								ProviderType: "list",
								Values: common.FieldValues{{
									Value:        "Traveling",
									DisplayValue: "Traveling",
								}, {
									Value:        "Food",
									DisplayValue: "Food",
								}, {
									Value:        "Art",
									DisplayValue: "Art",
								}},
							},
						},
					},
				},
				Errors: map[string]error{},
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
