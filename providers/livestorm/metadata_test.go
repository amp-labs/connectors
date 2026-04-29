package livestorm

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
			Name:       "Successful metadata for events and people",
			Input:      []string{"events", "people", "people_attributes", "jobs", "session_chat_messages"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"events": {
						DisplayName: "Events",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Event Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"title": {
								DisplayName:  "Title",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"people": {
						DisplayName: "People",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Person Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"email": {
								DisplayName:  "Email",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"people_attributes": {
						DisplayName: "People Attributes",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "People Attribute Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"slug": {
								DisplayName:  "Slug",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"jobs": {
						DisplayName: "Jobs",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Job Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"status": {
								DisplayName:  "Status",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"session_chat_messages": {
						DisplayName: "Session Chat Messages",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Chat Message Id",
								ValueType:    "string",
								ProviderType: "string",
							},
							"text": {
								DisplayName:  "Text",
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
			Input:      []string{"events", "unknown_object"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"events": {
						DisplayName: "Events",
						Fields: map[string]common.FieldMetadata{
							"id": {
								DisplayName:  "Event Id",
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
		tt := tt
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

	connector.SetUnitTestBaseURL(serverURL)

	return connector, nil
}
