package okta

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:   "Successfully describe users object",
			Input:  []string{"users"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						FieldsMap: map[string]string{
							"id":              "id",
							"status":          "status",
							"created":         "created",
							"activated":       "activated",
							"statusChanged":   "statusChanged",
							"lastLogin":       "lastLogin",
							"lastUpdated":     "lastUpdated",
							"passwordChanged": "passwordChanged",
							"type":            "type",
							"profile":         "profile",
							"credentials":     "credentials",
							"_links":          "_links",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe groups object",
			Input:  []string{"groups"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"groups": {
						DisplayName: "Groups",
						FieldsMap: map[string]string{
							"id":                    "id",
							"created":               "created",
							"lastUpdated":           "lastUpdated",
							"lastMembershipUpdated": "lastMembershipUpdated",
							"objectClass":           "objectClass",
							"type":                  "type",
							"profile":               "profile",
							"_links":                "_links",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
			ExpectedErrs: nil,
		},
		{
			Name:   "Successfully describe multiple objects",
			Input:  []string{"users", "groups"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						FieldsMap: map[string]string{
							"id":     "id",
							"status": "status",
						},
					},
					"groups": {
						DisplayName: "Groups",
						FieldsMap: map[string]string{
							"id":   "id",
							"type": "type",
						},
					},
				},
				Errors: nil,
			},
			Comparator:   testroutines.ComparatorSubsetMetadata,
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
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: &http.Client{},
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
