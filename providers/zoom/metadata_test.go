package zoom

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetaUserModule(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Metadata{ // nolint:gochecknoglobals
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unknown object requested",
			Input:        []string{"godzilla"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"users", "groups"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						FieldsMap: map[string]string{
							"display_name": "display_name",
							"dept":         "dept",
							"email":        "email",
							"status":       "status",
						},
					},
					"groups": {
						DisplayName: "Groups",
						FieldsMap: map[string]string{
							"name": "name",
							"id":   "id",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL, providers.ModuleZoomUser)
			})
		})
	}
}

func TestListObjectMetaMeetingModule(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Metadata{ // nolint:gochecknoglobals
		{
			Name:       "Successfully describe multiple objects with metadata",
			Input:      []string{"activities_report", "device_groups"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"activities_report": {
						DisplayName: "Activities Report",
						FieldsMap: map[string]string{
							"client_type": "client_type",
							"type":        "type",
							"email":       "email",
							"version":     "version",
						},
					},
					"device_groups": {
						DisplayName: "Device Groups",
						FieldsMap: map[string]string{
							"name":        "name",
							"description": "description",
						},
					},
				},
				Errors: nil,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()
			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL, providers.ModuleZoomMeeting)
			})
		})
	}
}

func constructTestConnector(serverURL string, moduleID common.ModuleID) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
		WithModule(moduleID),
	)
	if err != nil {
		return nil, err
	}
	// for testing we want to redirect calls to our mock server.
	connector.setBaseURL(serverURL)

	return connector, nil
}
