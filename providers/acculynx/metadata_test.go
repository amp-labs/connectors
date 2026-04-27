package acculynx

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unknown object returns ErrObjectNotSupported",
			Input:      []string{"nonexistent"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"nonexistent": common.ErrObjectNotSupported,
				},
			},
		},
		{
			Name:   "Successfully describe top-level jobs object",
			Input:  []string{"jobs"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"jobs": {
						DisplayName: "Jobs",
						FieldsMap: map[string]string{
							"id":               "id",
							"createdDate":      "createdDate",
							"currentMilestone": "currentMilestone",
						},
					},
				},
			},
			Comparator: testroutines.ComparatorSubsetMetadata,
		},
		{
			Name:   "Successfully describe top-level contacts object",
			Input:  []string{"contacts"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						FieldsMap: map[string]string{
							"id":        "id",
							"firstName": "firstName",
							"lastName":  "lastName",
						},
					},
				},
			},
			Comparator: testroutines.ComparatorSubsetMetadata,
		},
		{
			Name:   "Slash-named nested object resolves correctly",
			Input:  []string{"jobs/contacts"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"jobs/contacts": {
						DisplayName: "Job Contacts",
						FieldsMap: map[string]string{
							"id": "id",
						},
					},
				},
			},
			Comparator: testroutines.ComparatorSubsetMetadata,
		},
		{
			Name:   "Successfully describe multiple objects at once",
			Input:  []string{"jobs", "users", "calendars"},
			Server: mockserver.Dummy(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"jobs": {
						DisplayName: "Jobs",
						FieldsMap: map[string]string{
							"id": "id",
						},
					},
					"users": {
						DisplayName: "Users",
						FieldsMap: map[string]string{
							"id":          "id",
							"displayName": "displayName",
							"email":       "email",
						},
					},
					"calendars": {
						DisplayName: "Calendars",
						FieldsMap: map[string]string{
							"id":   "id",
							"name": "name",
						},
					},
				},
			},
			Comparator: testroutines.ComparatorSubsetMetadata,
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
		Module:              common.ModuleRoot,
		AuthenticatedClient: &http.Client{},
	})
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
