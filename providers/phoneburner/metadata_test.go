package phoneburner

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successful metadata for contacts, folders, members, voicemails, and dialsession",
			Input:      []string{"contacts", "folders", "members", "voicemails", "dialsession", "tags"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"contact_user_id": {
								DisplayName:  "Contact User Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"first_name": {
								DisplayName:  "First Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"folders": {
						DisplayName: "Folders",
						Fields: map[string]common.FieldMetadata{
							"folder_id": {
								DisplayName:  "Folder Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"folder_name": {
								DisplayName:  "Folder Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"members": {
						DisplayName: "Members",
						Fields: map[string]common.FieldMetadata{
							"user_id": {
								DisplayName:  "User Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"email_address": {
								DisplayName:  "Email Address",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"voicemails": {
						DisplayName: "Voicemails",
						Fields: map[string]common.FieldMetadata{
							"recording_id": {
								DisplayName:  "Recording Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"name": {
								DisplayName:  "Name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"dialsession": {
						DisplayName: "Dial Sessions",
						Fields: map[string]common.FieldMetadata{
							"dialsession_id": {
								DisplayName:  "Dialsession Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"start_when": {
								DisplayName:  "Start When",
								ValueType:    "datetime",
								ProviderType: "datetime",
							},
						},
					},
					"tags": {
						DisplayName: "Tags",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Tag Id",
								ValueType:    "int",
								ProviderType: "integer",
							},
							"title": {
								DisplayName:  "Title",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:         "Empty objects returns missing objects error",
			Input:        nil,
			Server:       mockserver.Dummy(),
			Expected:     nil,
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unsupported object returns object not supported error",
			Input:      []string{"contacts", "unknown_object"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"contacts": {
						DisplayName: "Contacts",
						Fields: map[string]common.FieldMetadata{
							"contact_user_id": {
								DisplayName:  "Contact User Id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{
					"unknown_object": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
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
			Module:              common.ModuleRoot,
			AuthenticatedClient: &http.Client{},
			Workspace:           "test-workspace",
		},
	)
	if err != nil {
		return nil, err
	}

	// Override the base URL to point to the test server.
	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
