package marketo

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	channelsResponse := testutils.DataFromFile(t, "channels.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe supported object",
			Input: []string{"channels"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.PathSuffix("v1/channels.json"),
					Then: mockserver.Response(http.StatusOK, channelsResponse),
				}},
			}.Server(),
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"channels": {
						DisplayName: "channels",
						FieldsMap: map[string]string{
							"applicableProgramType": "applicableProgramType",
							"createdAt":             "createdAt",
							"id":                    "id",
							"name":                  "name",
							"progressionStatuses":   "progressionStatuses",
							"updatedAt":             "updatedAt",
						},
					},
				},
				Errors: make(map[string]error),
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
		WithAuthenticatedClient(http.DefaultClient),
		WithWorkspace("test-workspace"),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
