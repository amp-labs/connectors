package pylon

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	contactsResponse := testutils.DataFromFile(t, "contacts-read.json")
	tagsResponse := testutils.DataFromFile(t, "tags-read.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"contacts", "tags"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/contacts"),
					Then: mockserver.Response(http.StatusOK, contactsResponse),
				}, {
					If:   mockcond.Path("/tags"),
					Then: mockserver.Response(http.StatusOK, tagsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"account": {
								DisplayName:  "account",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"avatar_url": {
								DisplayName:  "avatar_url",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"custom_fields": {
								DisplayName:  "custom_fields",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"emails": {
								DisplayName:  "emails",
								ValueType:    "other",
								ProviderType: "",
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"portal_role": {
								DisplayName:  "portal_role",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
					"tags": {
						DisplayName: "Tags",
						Fields: map[string]common.FieldMetadata{
							"hex_color": {
								DisplayName:  "hex_color",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
							"object_type": {
								DisplayName:  "object_type",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},

							"value": {
								DisplayName:  "value",
								ValueType:    "string",
								ProviderType: "",
								Values:       nil,
							},
						},
						FieldsMap: nil,
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
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
		AuthenticatedClient: mockutils.NewClient(),
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
