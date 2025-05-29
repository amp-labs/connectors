package capsule

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/parameters"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successful metadata for multiple objects",
			Input:      []string{"activitytypes", "projects"},
			Server:     mockserver.Dummy(),
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
								ReadOnly:     true,
							},
							"updateLastContacted": {
								DisplayName:  "Update Last Contacted",
								ValueType:    "boolean",
								ProviderType: "Boolean",
							},
						},
					},
					"projects": {
						DisplayName: "Projects",
						Fields: map[string]common.FieldMetadata{
							"createdAt": {
								DisplayName:  "Created At",
								ValueType:    "date",
								ProviderType: "Date",
								ReadOnly:     true,
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
	connector, err := NewConnector(parameters.Connector{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
