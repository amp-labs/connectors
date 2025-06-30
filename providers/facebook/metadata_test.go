package facebook

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
						Fields:      map[string]common.FieldMetadata{},
						FieldsMap: map[string]string{
							"id":    "id",
							"name":  "name",
							"tasks": "tasks",
						},
					},
					"adimages": {
						DisplayName: "Adimages",
						Fields:      map[string]common.FieldMetadata{},
						FieldsMap: map[string]string{
							"id":   "id",
							"hash": "hash",
						},
					},
					"system_users": {
						DisplayName: "System_users",
						Fields:      map[string]common.FieldMetadata{},
						FieldsMap: map[string]string{
							"id":   "id",
							"name": "name",
							"role": "role",
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
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Metadata: map[string]string{
			"adAccountId": "1214321106978726",
			"businessId":  "1190021932394709",
		},
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
