package flatfile

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
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
			Input:      []string{"godzilla"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"godzilla": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"users", "prompts"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields:      nil,
						FieldsMap: map[string]string{
							"id":         "id",
							"name":       "name",
							"email":      "email",
							"metadata":   "metadata",
							"accountId":  "accountId",
							"idp":        "idp",
							"idpRef":     "idpRef",
							"createdAt":  "createdAt",
							"updatedAt":  "updatedAt",
							"lastSeenAt": "lastSeenAt",
							"dashboard":  "dashboard",
						},
					},
					"prompts": {
						DisplayName: "Prompts",
						Fields:      nil,
						FieldsMap: map[string]string{
							"id":            "id",
							"createdById":   "createdById",
							"accountId":     "accountId",
							"promptType":    "promptType",
							"prompt":        "prompt",
							"createdAt":     "createdAt",
							"updatedAt":     "updatedAt",
							"environmentId": "environmentId",
							"spaceId":       "spaceId",
							"deletedAt":     "deletedAt",
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
	})
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
