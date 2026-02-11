package phoneburner

import (
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:       "Successful metadata for contacts, folders, members, voicemails, dialsession and tags",
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
						},
					},
				},
			},
		},
		{
			Name:         "Object must be included",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unsupported object returns metadata error entry",
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
