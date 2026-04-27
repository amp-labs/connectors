package getresponse

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object requested",
			Input:      []string{"nonexistent_object"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"nonexistent_object": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Successfully describe contacts object with metadata",
			Input:      []string{"contacts"},
			Server:     testServerCustomFieldsListJSON(`[]`),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
			},
		},
		{
			Name:  "Describe contacts merges custom field definitions",
			Input: []string{"contacts"},
			Server: testServerCustomFieldsListJSON(`[{"customFieldId":"fld1","name":"Tier","fieldType":"text","valueType":"single_select","values":["gold","silver"]}]`),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
							"cf_fld1": {
								DisplayName:  "Tier",
								ValueType:    common.ValueTypeSingleSelect,
								ProviderType: "single_select",
								Values: []common.FieldValue{
									{Value: "gold", DisplayValue: "gold"},
									{Value: "silver", DisplayValue: "silver"},
								},
								IsCustom: boolPtr(true),
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

func boolPtr(b bool) *bool {
	return &b
}

// testServerCustomFieldsListJSON serves GET /v3/custom-fields with the given JSON body.
// Other requests get an empty custom-field list so ListObjectMetadata does not hit a teapot dummy.
func testServerCustomFieldsListJSON(customFieldsJSON string) *httptest.Server {
	return mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{
			{
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/v3/custom-fields"),
				},
				Then: mockserver.Response(http.StatusOK, []byte(customFieldsJSON)),
			},
		},
		Default: mockserver.Response(http.StatusOK, []byte(`[]`)),
	}.Server()
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
