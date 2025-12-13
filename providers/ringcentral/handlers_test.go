package ringcentral

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop,maintidx
	t.Parallel()

	okResponse := testutils.DataFromFile(t, "meetings.json")
	unsupportedObjects := testutils.DataFromFile(t, "404.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},

		{
			Name:  "Server response must have at least one field",
			Input: []string{"butterflies"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusNotFound, unsupportedObjects),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"butterflies": common.ErrRetryable,
				},
			},
		},

		{
			Name:  "Successfully describe Meetings metadata",
			Input: []string{"meetings"},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, okResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"meetings": {
						DisplayName: "Meetings",
						Fields: map[string]common.FieldMetadata{
							"bridgeId": {
								DisplayName:  "bridgeId",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"chatContentUrl": {
								DisplayName:  "chatContentUrl",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"displayName": {
								DisplayName:  "displayName",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"duration": {
								DisplayName:  "duration",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "float",
								Values:       nil,
							},
							"hostInfo": {
								DisplayName:  "hostInfo",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"id": {
								DisplayName:  "id",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"participants": {
								DisplayName:  "participants",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "other",
								Values:       nil,
							},
							"shortId": {
								DisplayName:  "shortId",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
							"startTime": {
								DisplayName:  "startTime",
								IsCustom:     nil,
								IsRequired:   nil,
								ProviderType: "",
								ReadOnly:     nil,
								ValueType:    "string",
								Values:       nil,
							},
						},
						FieldsMap: nil,
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
