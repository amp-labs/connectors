package closecrm

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

	leadsResponse := testutils.DataFromFile(t, "leads.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be provided",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successful describe supported objects",
			Input: []string{"lead"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/api/v1/lead"),
					Then: mockserver.Response(http.StatusOK, leadsResponse),
				}},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"lead": {
						DisplayName: "lead",
						FieldsMap: map[string]string{
							"addresses":       "addresses",
							"contacts":        "contacts",
							"created_by":      "created_by",
							"created_by_name": "created_by_name",
							"custom":          "custom",
							"custom.cf_9RsqbFDfApgIOJsiSCwRuKw5gdVibt6H7qEvn4nQDjN": "custom.cf_9RsqbFDfApgIOJsiSCwRuKw5gdVibt6H7qEvn4nQDjN",
							"custom.cf_VrDASliEOprzQEnsi88KGAEVjGUOB00ea12dpa63hlr": "custom.cf_VrDASliEOprzQEnsi88KGAEVjGUOB00ea12dpa63hlr",
							"custom.cf_X4BeCHhnzEzwqZoZoYi5l80jcA99DsC0YXf1B8wYWYn": "custom.cf_X4BeCHhnzEzwqZoZoYi5l80jcA99DsC0YXf1B8wYWYn",
							"custom.cf_fEpuyPj0nNEvtYFotU9xHwNvcMe7vJJ22S6H094WoxY": "custom.cf_fEpuyPj0nNEvtYFotU9xHwNvcMe7vJJ22S6H094WoxY",
							"custom.cf_uOfms5LQFL39IzS0XDhqaTuxsrCyTVTlbzvANKFQ8Lu": "custom.cf_uOfms5LQFL39IzS0XDhqaTuxsrCyTVTlbzvANKFQ8Lu",
							"date_created":      "date_created",
							"date_updated":      "date_updated",
							"description":       "description",
							"display_name":      "display_name",
							"html_url":          "html_url",
							"id":                "id",
							"integration_links": "integration_links",
							"name":              "name",
							"opportunities":     "opportunities",
							"organization_id":   "organization_id",
							"status_id":         "status_id",
							"status_label":      "status_label",
							"tasks":             "tasks",
							"updated_by":        "updated_by",
							"updated_by_name":   "updated_by_name",
							"url":               "url",
						},
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
		WithAuthenticatedClient(mockutils.NewClient()),
	)
	if err != nil {
		return nil, err
	}

	// for testing we want to redirect calls to our mock server
	connector.setBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
