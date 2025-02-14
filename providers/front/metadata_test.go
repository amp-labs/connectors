package front

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	teams := testutils.DataFromFile(t, "teams.json")
	companyRules := testutils.DataFromFile(t, "notfound.json")
	meme := testutils.DataFromFile(t, "notfound.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe supported & unsupported objects",
			Input: []string{"teams", "meme", "company_rules"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If:   mockcond.PathSuffix("/teams"),
						Then: mockserver.Response(http.StatusOK, teams),
					}, {
						If:   mockcond.PathSuffix("/company_rules"),
						Then: mockserver.Response(http.StatusNotFound, companyRules),
					}, {
						If:   mockcond.PathSuffix("/meme"),
						Then: mockserver.Response(http.StatusNotFound, meme),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"teams": {
						DisplayName: "Teams",
						FieldsMap: map[string]string{
							"_links": "_links",
							"id":     "id",
							"name":   "name",
						},
					},
					"company_rules": {
						DisplayName: "company_rules",
						FieldsMap: map[string]string{
							"_links":     "_links",
							"actions":    "actions",
							"id":         "id",
							"is_private": "is_private",
							"name":       "name",
						},
					},
				},
				Errors: map[string]error{
					"meme": mockutils.ExpectedSubsetErrors{
						common.ErrObjectNotSupported,
					},
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
		WithAuthenticatedClient(http.DefaultClient),
	)
	if err != nil {
		return nil, err
	}

	connector.BaseURL = serverURL

	return connector, nil
}
