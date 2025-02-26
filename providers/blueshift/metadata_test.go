package blueshift

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen,gocognit,cyclop
	t.Parallel()

	tests := []testroutines.Metadata{
		{
			Name:         "Atleast one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "Unknown object requested",
			Input:        []string{"groot"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrObjectNotSupported},
		},

		{
			Name:       "Successfully describe multiple object with metadata",
			Input:      []string{"campaigns", "email_templates", "segments/list"},
			Server:     mockserver.Dummy(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"campaigns": {
						DisplayName: "Campaigns",
						FieldsMap: map[string]string{
							"archived": "archived",
							"author":   "author",
							"name":     "name",
						},
					},
					"email_templates": {
						DisplayName: "Email Templates",
						FieldsMap: map[string]string{
							"author": "author",
							"uuid":   "uuid",
							"name":   "name",
						},
					},
					"segments/list": {
						DisplayName: "Segments/List",
						FieldsMap: map[string]string{
							"approxusers": "approxusers",
							"status":      "status",
							"name":        "name",
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
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(serverURL)

	return connector, nil
}
