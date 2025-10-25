package meta

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestFacebookListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	usersResponse := testutils.DataFromFile(t, "users.json")
	adimagesResponse := testutils.DataFromFile(t, "adimages.json")
	systemUsersResponse := testutils.DataFromFile(t, "system_users.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe multiple objects with metadata",
			Input: []string{"users", "adimages", "system_users"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/v19.0/act_1214321106978726/users"),
					Then: mockserver.Response(http.StatusOK, usersResponse),
				}, {
					If:   mockcond.Path("/v19.0/act_1214321106978726/adimages"),
					Then: mockserver.Response(http.StatusOK, adimagesResponse),
				}, {
					If:   mockcond.Path("/v19.0/1190021932394709/system_users"),
					Then: mockserver.Response(http.StatusOK, systemUsersResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
							"name": {
								DisplayName: "name",
								ValueType:   "other",
							},
							"tasks": {
								DisplayName: "tasks",
								ValueType:   "other",
							},
						},
						FieldsMap: nil,
					},
					"adimages": {
						DisplayName: "Adimages",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
							"hash": {
								DisplayName: "hash",
								ValueType:   "other",
							},
						},
						FieldsMap: nil,
					},
					"system_users": {
						DisplayName: "System_users",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName: "id",
								ValueType:   "other",
							},
							"name": {
								DisplayName: "name",
								ValueType:   "other",
							},
							"role": {
								DisplayName: "role",
								ValueType:   "other",
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
				return constructTestFacebookConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestFacebookConnector(serverURL string) (*Connector, error) {
	return constructTestConnector(serverURL, providers.ModuleFacebook)
}

func constructTestConnector(serverURL string, moduleID common.ModuleID) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              moduleID,
			AuthenticatedClient: mockutils.NewClient(),
			Metadata: map[string]string{
				"adAccountId": "1214321106978726",
				"businessId":  "1190021932394709",
			},
		},
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setUnitTestBaseURL(mockutils.ReplaceURLOrigin(connector.ModuleInfo().BaseURL, serverURL))

	return connector, nil
}
