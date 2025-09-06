package nutshell

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestCalendarListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successful metadata for CalendarList and Settings",
			Input:      []string{"competitormaps", "contacts", "products"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"competitormaps": {
						DisplayName: "Competitor Maps",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{{
									Value:        "potential",
									DisplayValue: "potential",
								}, {
									Value:        "stole",
									DisplayValue: "stole",
								}, {
									Value:        "beat",
									DisplayValue: "beat",
								}},
							},
						},
					},
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"firstName": {
								DisplayName: "firstName",
								ValueType:   "other",
							},
							"href": {
								DisplayName:  "href",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"products": {
						DisplayName: "Products",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName: "name",
								ValueType:   "other",
							},
							"price": {
								DisplayName:  "price",
								ValueType:    "other",
								ProviderType: "object",
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
	connector, err := NewConnector(
		common.ConnectorParams{
			AuthenticatedClient: mockutils.NewClient(),
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(mockutils.ReplaceURLOrigin(connector.ModuleInfo().BaseURL, serverURL))

	return connector, nil
}
