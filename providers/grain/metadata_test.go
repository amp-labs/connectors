package grain

import (
	"testing"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
)

func TestListObjectMetadata(t *testing.T) {
	t.Parallel()

	tests := []testconn.TestCaseListObjectMetadata{
		{
			Name: "Successful metadata for all Grain objects",
			Input: []string{
				"recordings",
				"uploads",
				"hooks",
				"users",
				"teams",
				"meeting_types",
			},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"recordings": {
						DisplayName: "Recordings",
						Fields: map[string]common.FieldMetadata{
							"start_datetime": {
								DisplayName:  "start_datetime",
								ValueType:    "datetime",
								ProviderType: "datetime",
							},
							"media_type": {
								DisplayName:  "media_type",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "video", DisplayValue: "video"},
									{Value: "audio", DisplayValue: "audio"},
									{Value: "transcript", DisplayValue: "transcript"},
								},
							},
							"source": {
								DisplayName:  "source",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "aircall", DisplayValue: "aircall"},
									{Value: "local_capture", DisplayValue: "local_capture"},
									{Value: "meet", DisplayValue: "meet"},
									{Value: "teams", DisplayValue: "teams"},
									{Value: "upload", DisplayValue: "upload"},
									{Value: "webex", DisplayValue: "webex"},
									{Value: "zoom", DisplayValue: "zoom"},
									{Value: "other", DisplayValue: "other"},
								},
							},
							"share_state": {
								DisplayName:  "share_state",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "public", DisplayValue: "public"},
									{Value: "workspace", DisplayValue: "workspace"},
									{Value: "restricted", DisplayValue: "restricted"},
								},
							},
							"sensitivity": {
								DisplayName:  "sensitivity",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"uploads": {
						DisplayName: "Uploads",
						Fields: map[string]common.FieldMetadata{
							"updated_datetime": {
								DisplayName:  "updated_datetime",
								ValueType:    "datetime",
								ProviderType: "datetime",
							},
							"status": {
								DisplayName:  "status",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "awaiting", DisplayValue: "awaiting"},
									{Value: "processing", DisplayValue: "processing"},
									{Value: "staged", DisplayValue: "staged"},
									{Value: "promoting", DisplayValue: "promoting"},
									{Value: "success", DisplayValue: "success"},
									{Value: "failure", DisplayValue: "failure"},
								},
							},
						},
					},
					"hooks": {
						DisplayName: "Hooks",
						Fields: map[string]common.FieldMetadata{
							"hook_url": {
								DisplayName:  "hook_url",
								ValueType:    "string",
								ProviderType: "string",
							},
							"hook_type": {
								DisplayName:  "hook_type",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "recording_added", DisplayValue: "recording_added"},
									{Value: "recording_updated", DisplayValue: "recording_updated"},
									{Value: "recording_deleted", DisplayValue: "recording_deleted"},
									{Value: "highlight_added", DisplayValue: "highlight_added"},
									{Value: "highlight_updated", DisplayValue: "highlight_updated"},
									{Value: "highlight_deleted", DisplayValue: "highlight_deleted"},
									{Value: "story_added", DisplayValue: "story_added"},
									{Value: "story_updated", DisplayValue: "story_updated"},
									{Value: "story_deleted", DisplayValue: "story_deleted"},
									{Value: "upload_added", DisplayValue: "upload_added"},
									{Value: "upload_updated", DisplayValue: "upload_updated"},
								},
							},
						},
					},
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"email": {
								DisplayName:  "email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"teams": {
						DisplayName: "Teams",
						Fields: map[string]common.FieldMetadata{
							"name": {
								DisplayName:  "name",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"meeting_types": {
						DisplayName: "Meeting Types",
						Fields: map[string]common.FieldMetadata{
							"scope": {
								DisplayName:  "scope",
								ValueType:    "singleSelect",
								ProviderType: "string",
								Values: []common.FieldValue{
									{Value: "internal", DisplayValue: "internal"},
									{Value: "external", DisplayValue: "external"},
								},
							},
						},
					},
				},
				Errors: map[string]error{},
			},
		},
		{
			Name:         "Empty objects returns missing objects error",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:       "Unsupported object returns object not supported error",
			Input:      []string{"users", "stories"},
			Server:     mockserver.Dummy(),
			Comparator: testconn.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "id",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
				},
				Errors: map[string]error{
					"stories": mockutils.ExpectedSubsetErrors{common.ErrObjectNotSupported},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableMetadataReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		common.ConnectorParams{
			Module:              common.ModuleRoot,
			AuthenticatedClient: mockutils.NewClient(),
		},
	)
	if err != nil {
		return nil, err
	}

	connector.SetUnitTestMockServerBaseUrl(serverURL)

	return connector, nil
}
