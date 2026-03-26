package microsoft

import (
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
			Name:       "Successful metadata for multiple objects",
			Input:      []string{"users", "groups", "me/events", "me/messages"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"users": {
						DisplayName: "Users",
						Fields: map[string]common.FieldMetadata{
							"city": {
								DisplayName:  "city",
								ValueType:    "string",
								ProviderType: "string",
							},
							"skills": {
								DisplayName:  "skills",
								ValueType:    "other",
								ProviderType: "array",
							},
							"surname": {
								DisplayName:  "surname",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"groups": {
						DisplayName: "Groups",
						Fields: map[string]common.FieldMetadata{
							"isSubscribedByMail": {
								DisplayName:  "isSubscribedByMail",
								ValueType:    "boolean",
								ProviderType: "boolean",
							},
							"mail": {
								DisplayName:  "mail",
								ValueType:    "string",
								ProviderType: "string",
							},
							"createdDateTime": {
								DisplayName:  "createdDateTime",
								ValueType:    "string",
								ProviderType: "string",
							},
						},
					},
					"me/events": {
						DisplayName: "Events",
						Fields: map[string]common.FieldMetadata{
							"attendees": {
								DisplayName:  "attendees",
								ValueType:    "other",
								ProviderType: "array",
							},
						},
					},
					"me/messages": {
						DisplayName: "Messages",
						Fields: map[string]common.FieldMetadata{
							"from": {
								DisplayName: "from",
								ValueType:   "other",
							},
							"body": {
								DisplayName: "body",
								ValueType:   "other",
							},
						},
					},
				},
				Errors: map[string]error{},
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
	connector, err := NewConnector(common.ConnectorParams{
		Module:              common.ModuleRoot,
		AuthenticatedClient: mockutils.NewClient(),
		Workspace:           "test-workspace",
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
