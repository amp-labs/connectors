package heyreach

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	campaignResponse := testutils.DataFromFile(t, "campaign.json")
	listResponse := testutils.DataFromFile(t, "list.json")
	liAccountResponse := testutils.DataFromFile(t, "li_account.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"campaign/GetAll", "list/GetAll", "li_account/GetAll"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("public/campaign/GetAll"),
					Then: mockserver.Response(http.StatusOK, campaignResponse),
				}, {
					If:   mockcond.PathSuffix("public/list/GetAll"),
					Then: mockserver.Response(http.StatusOK, listResponse),
				}, {
					If:   mockcond.PathSuffix("public/li_account/GetAll"),
					Then: mockserver.Response(http.StatusOK, liAccountResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"campaign/GetAll": {
						DisplayName: "GetAll campaign",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"creationTime": {
								DisplayName:  "creationTime",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"id":           "id",
							"name":         "name",
							"creationTime": "creationTime",
							"status":       "status",
						},
					},
					"list/GetAll": {
						DisplayName: "GetAll list",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"name": {
								DisplayName:  "name",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"creationTime": {
								DisplayName:  "creationTime",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"listType": {
								DisplayName:  "listType",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"id":           "id",
							"name":         "name",
							"creationTime": "creationTime",
							"listType":     "listType",
						},
					},
					"li_account/GetAll": {
						DisplayName: "GetAll li_account",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"emailAddress": {
								DisplayName:  "emailAddress",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"firstName": {
								DisplayName:  "firstName",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
							"lastName": {
								DisplayName:  "lastName",
								ValueType:    "other",
								ProviderType: "",
								ReadOnly:     false,
								Values:       nil,
							},
						},
						FieldsMap: map[string]string{
							"id":           "id",
							"emailAddress": "emailAddress",
							"firstName":    "firstName",
							"lastName":     "lastName",
						},
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
	connector, err := NewConnector(common.Parameters{
		Module:              common.ModuleRoot,
		AuthenticatedClient: http.DefaultClient,
	})
	if err != nil {
		return nil, err
	}

	testroutines.OverrideURLOrigin(&connector.URLManager, connector.ProviderInfo(), serverURL)

	return connector, nil
}
